package main

import (
	"github.com/aud/redis_complete/autocomplete"
	"github.com/garyburd/redigo/redis"
	"testing"
	"time"
)

func BenchmarkReadWordListFunction(b *testing.B) {
	for n := 0; n < b.N; n++ {
		readWordList("wordlist.txt")
	}
}

func BenchmarkHandleResultValuesFunction(b *testing.B) {
	pool := &redis.Pool{
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

	defer pool.Close()

	a := autocomplete.New(pool, "abc", "newKey", 1)
	a.AddToList([]string{"abc", "def", "hello"})

	values, _ := a.LexicographicalOrder()

	for n := 0; n < b.N; n++ {
		handleResultValues(values)
	}
}
