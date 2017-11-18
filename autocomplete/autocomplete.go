package autocomplete

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
)

type Autocomplete struct {
	pool   *redis.Pool
	prefix string
	key    string
	limit  int
}

func New(pool *redis.Pool, prefix string, key string, limit int) *Autocomplete {
	return &Autocomplete{
		pool:   pool,
		prefix: prefix,
		key:    key,
		limit:  limit,
	}
}

// Check if the Redis key exists so we don't redundantly read the wordlist
// into memory and try to readd the values, as it's an expensive operation.
func (a *Autocomplete) KeyExists() (int64, error) {
	conn := a.pool.Get()
	defer conn.Close()

	result, err := conn.Do("EXISTS", a.key)

	return result.(int64), err
}

// Create a range using the string the user provides for the 'start' of the range,
// and the same string with a trailing byte set to 255, as \xff, for the end of the range.
// This way we get all the strings that start with the provided string.
func createQuery(prefix string) (string, string) {
	return fmt.Sprintf("[%s", prefix), fmt.Sprintf("[%s\xff", prefix)
}

// Iterates through wordlist, ZADDs each list item with the same initial score of 1.
// This operation can get pretty expensive (O(n)), depending on size of wordlist.
func (a *Autocomplete) AddToList(list []string) {
	conn := a.pool.Get()
	defer conn.Close()

	for _, l := range list {
		conn.Do("ZADD", a.key, 0, fmt.Sprintf("%s:%v", l, 1))
	}
}

// Elements with the same score are sorted comparing the raw values of their bytes,
// byte after byte. If the first byte is the same, the second is checked and so forth.
// If the common prefix of two strings is the same then the longer string is considered
// the greater of the two, so "foobar" is greater than "foo".
// This is a complexity of O(log(n) + m), with n being the sorted set and
// m being the number of elements returned.
func (a *Autocomplete) LexicographicalOrder() ([]interface{}, error) {
	conn := a.pool.Get()
	defer conn.Close()

	prefixOne, prefixTwo := createQuery(a.prefix)

	values, err := redis.Values(conn.Do("ZRANGEBYLEX", a.key, prefixOne, prefixTwo, "LIMIT", 0, a.limit))

	return values, err
}

// Check if initial search prefix has a 1:1 match, if so remove
// key and readd it with frequency incremented by 1.
func (a *Autocomplete) HandleExactMatchFrequency(results []string) {
	for _, r := range results {
		// Split at frequency to get normalized string.
		normalized := strings.Split(string(r), ":")

		// Check if initial search prefix has a 1:1 match, if so remove
		// key and readd it with an incremented by 1 frequency.
		if a.prefix == normalized[0] {
			conn := a.pool.Get()
			defer conn.Close()

			conn.Do("ZREM", a.key, 0, string(r))

			freq, _ := strconv.Atoi(normalized[1])
			conn.Do("ZADD", a.key, 0, fmt.Sprintf("%s:%v", normalized[0], freq+1))
		}
	}
}
