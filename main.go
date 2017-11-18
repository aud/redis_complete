package main

import (
	"bufio"
	"flag"
	"fmt"
	ac "github.com/aud/redis_complete/autocomplete"
	"github.com/garyburd/redigo/redis"
	"os"
	"time"
)

var (
	query string
	list  string
	port  int
	limit int
	key   = "acIndex"
)

func init() {
	q := flag.String("query", "", "input query")
	l := flag.String("list", "", "wordlist path")
	s := flag.Int("limit", 0, "wordlist limit")
	p := flag.Int("port", 6379, "redis port")
	flag.Parse()

	query = *q
	limit = *s
	list = *l
	port = *p
}

func main() {
	pool := createRedisPool()
	defer pool.Close()

	Autocomplete := ac.New(pool, query, key, limit)
	exists, err := Autocomplete.KeyExists()

	if err != nil {
		panic(err)
	}

	if exists == 0 {
		wordlist, err := readWordList(list)

		if err != nil {
			panic(err)
		}

		Autocomplete.AddToList(wordlist)
	}

	values, err := Autocomplete.LexicographicalOrder()

	if err != nil {
		panic(err)
	}

	results := handleResultValues(values)
	Autocomplete.HandleExactMatchFrequency(results)

	for _, r := range results {
		fmt.Printf("%s\n", r)
	}
}

func createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   20,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf(":%v", port))
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// Read wordlist into memory and return string slice.
func readWordList(path string) ([]string, error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	var list []string
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}

	return list, scanner.Err()
}

// Type coerces result to byte array, so we can more easily convert
// results to strings and handle the values.
func handleResultValues(values []interface{}) []string {
	var arr []string

	for _, r := range values {
		b, _ := r.([]byte)
		arr = append(arr, string(b))
	}

	return arr
}
