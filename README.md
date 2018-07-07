This is definitely not meant to be used by anyone. Just some example.

Given a wordlist seperated by `\n`, this will attempt to traverse through the list and `ZADD` [1] each item to the index with an initial frequency of 1. If the command `EXISTS` [2] returns true for the key, this step is redundant and will be skipped.

```
ZADD index 0 foo:1
```

The autocomplete function works by executing a range query using the command `ZRANGEBYLEX` [3] with the string provided, which returns all elements in the set sorted lexicographically between the given range. As an example, if the query provided is `foo`, we want to return all the results starting with `foo`. In this case the range would have a min of `[foo` with a max of `[foo\xff` (trailing byte).

```
ZRANGEBYLEX index "[foo" "[foo\xff"
```

This will then check if the initial search prefix was a 1:1 match with the result, if so we `ZREM` [4] the value and `ZADD` it back with an incremented frequency of 1, making the new value `foo:2`.

- [1] https://redis.io/commands/zadd
- [2] https://redis.io/commands/exists
- [3] https://redis.io/commands/zrangebylex
- [4] https://redis.io/commands/zrem

## Options
```
$ go run main.go \
--query="foo" \
--limit=10 \
--list="path_to_wordlist.txt" \
--port=6379
```
## Benchmarks

`$ go test ./... --bench=.`
```
BenchmarkReadWordListFunction-8         	     100	  13850440 ns/op
BenchmarkHandleResultValuesFunction-8   	20000000	        74.8 ns/op
BenchmarkLexicographicalOrderFunction-8        	   20000	     68123 ns/op
BenchmarkAddToListFunction-8                   	   10000	    118421 ns/op
BenchmarkCreateQueryFunction-8                 	 5000000	       262 ns/op
BenchmarkKeyExistsFunction-8                   	   30000	     55203 ns/op
BenchmarkHandleExactMatchFrequencyFunction-8   	   20000	     89283 ns/op
```

