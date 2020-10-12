# apguard
distributed proxy that acts as a rate limiter for downstream http services

# Proxy a service in docker compose
docker-compose.yml
```yml
version: "3.8"
services:
  apguard:
    image: apguard/rpx:latest
    ports:
      - "9001:80"
    environment:
        - PROXY_CONF=/etc/proxy/conf.yml
    volumes:
        - "./example-conf.yml:/etc/proxy/conf.yml"

  greeter:
    image: apguard/greeter:latest
    ports:
        - "9000:80"
```

example-conf.yml
```yml
destination: http://greeter
rules:
    - name: vegeta-ua
      mode: ua
      pattern: "vegeta"
      delay: 1s
```


Other modes are 
- `ip+ua`: means the same ip could have different user agents allowed to query the api
- `ip`:    rate limits on ip address only




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
```
go mod vendor
make build && make ex
brew install vegeta # install vegeta if you don't already have it
vegeta attack -duration=15s -rate=1000/s -targets=example/proxy-targets.conf | tee results.bin | vegeta report
```


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
