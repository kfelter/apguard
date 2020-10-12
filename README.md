# apguard
distributed proxy that acts as a rate limiter for downstream http services

# Usage
1. `go mod vendor`
2. start the greeter service `go run cmd/greeter/main.go`
3. start the apguard `go run cmd/rps/main.go -exp=30s`
4. curl apguard a few times and see the blocked requests `curl localhost:9001`

### greeter logs
```
kylefelter@iMac apguard % go run cmd/greeter/main.go
2020/10/11 01:57:13 Hello there [::1]:53393
```
only one request go through to the greeter


### apguard logs
```
kylefelter@iMac apguard % go run cmd/rps/main.go -exp=30s
blocked ip=[::1]:53384(::1) ua=curl/7.64.1
blocked ip=[::1]:53395(::1) ua=curl/7.64.1
```
3 total requests were made but 2 of the three were blocked because they came from the same ip address

# Production usage
default uses in memory as backend but set REDIS_URL to point to a redis instance to have many proxies that know about other requests

# Example
find an example in the `/example` folder

to run the example
1. `go mod vendor`
2. `make build && make ex`
3. (install vegeta if you don't already have it) `brew install vegeta`
4 `vegeta attack -duration=15s -rate=1000/s -targets=example/proxy-targets.conf | tee results.bin | vegeta report`


Vegeta report
```
Requests      [total, rate, throughput]         75000, 5000.06, 1.07
Duration      [total, attack, wait]             15.002s, 15s, 1.862ms
Latencies     [min, mean, 50, 90, 95, 99, max]  531.896µs, 34.129ms, 12.518ms, 100.251ms, 140.86ms, 226.763ms, 1.203s
Bytes In      [total, mean]                     192, 0.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           0.02%
Status Codes  [code:count]                      200:16  429:74984  
Error Set:
429 Too Many Requests
```

Vegeta plot
![](vegeta-plot.png?raw=true)
