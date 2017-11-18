package autocomplete

import (
	"github.com/garyburd/redigo/redis"
	"testing"
	"time"
)

func createPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   20,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
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

func BenchmarkLexicographicalOrderFunction(b *testing.B) {
	pool := createPool()
	defer pool.Close()

	a := New(pool, "abc", "newKey", 1)

	for n := 0; n < b.N; n++ {
		a.LexicographicalOrder()
	}
}

func BenchmarkAddToListFunction(b *testing.B) {
	pool := createPool()
	defer pool.Close()

	a := New(pool, "abc", "newKey", 1)

	for n := 0; n < b.N; n++ {
		a.AddToList([]string{"abc", "def", "hello"})
	}
}

func BenchmarkCreateQueryFunction(b *testing.B) {
	for n := 0; n < b.N; n++ {
		createQuery("term")
	}
}

func BenchmarkKeyExistsFunction(b *testing.B) {
	pool := createPool()
	defer pool.Close()

	a := New(pool, "abc", "newKey", 1)

	for n := 0; n < b.N; n++ {
		a.KeyExists()
	}
}

func BenchmarkHandleExactMatchFrequencyFunction(b *testing.B) {
	pool := createPool()
	defer pool.Close()

	a := New(pool, "abc", "newKey", 1)

	results := []string{"abc:1", "def:1", "hello:1"}

	for n := 0; n < b.N; n++ {
		a.HandleExactMatchFrequency(results)
	}
}
