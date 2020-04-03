To run benchmarks, go to the project root and do:

```shell script
go test -bench=BenchmarkClientGzip/compression=0 -count 5 -benchmem -benchtime=10s | tee -a bench/result-0.txt
go test -bench=BenchmarkClientGzip/compression=3 -count 5 -benchmem -benchtime=10s | tee -a bench/result-3.txt
go test -bench=BenchmarkClientGzip/compression=6 -count 5 -benchmem -benchtime=10s | tee -a bench/result-6.txt
go test -bench=BenchmarkClientGzip/compression=9 -count 5 -benchmem -benchtime=10s | tee -a bench/result-9.txt

benchstat bench/result-0.txt bench/result-3.txt
benchstat bench/result-0.txt bench/result-6.txt
benchstat bench/result-9.txt bench/result-9.txt
```


go test -bench=BenchmarkClientGzip/compression=0 -count 1 -benchmem -benchtime=5s > bench/result-0.txt