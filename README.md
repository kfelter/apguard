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

# default uses in memory as backend but set REDIS_URL to point to a redis instance to have many proxies that know about other requests

