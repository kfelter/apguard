```go
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	p = flag.String("port", "80", "server port")
)

func main() {
	http.HandleFunc("/", greetHandler)
	panic(http.ListenAndServe(":"+*p, nil))
}

func greetHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Hello there %s\n", req.RemoteAddr)
	ioutil.ReadAll(req.Body)
	req.Body.Close()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Hello there` + "\n"))
}
```

```go
package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	p      = flag.String("port", "80", "port")
	origin = flag.String("origin", "", "url to be proxied")
)

func main() {
	flag.Parse()

	parsedURL, _ := url.Parse(*origin)

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	http.HandleFunc("/", proxy.ServeHTTP)

	log.Fatal(http.ListenAndServe(":"+*p, nil))
}
```
```go
package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	p      = flag.String("port", "80", "port")
	origin = flag.String("origin", "", "url to be proxied")
)

func main() {
	flag.Parse()

	parsedURL, _ := url.Parse(*origin)

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	http.HandleFunc("/", loggingMiddleware(proxy.ServeHTTP))

	log.Fatal(http.ListenAndServe(":"+*p, nil))
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, "proxied", r.RequestURI)
		next.ServeHTTP(w, r)
	}
}
```
```go
func limiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rule, key, found := proxyConf.MatchesRule(r)
		if !found {
			next.ServeHTTP(w, r)
			return
		}

		if _, err := kvDB.Get(key); err == nil {
			w.WriteHeader(http.StatusTooManyRequests)
			logCh <- fmt.Sprintf("blocked key=%s rule=%s\n", key, rule.Name)
			return
		}
		kvDB.Set(key, true, rule.Delay)
		next.ServeHTTP(w, r)
	}
}
```
```go
// Rule holds our configured reverse proxy rules
type Rule struct {
	Name                string        `yaml:"name"`
	Mode                string        `yaml:"mode"`
	UARegex             string        `yaml:"pattern"`
	TimeBetweenRequests string        `yaml:"delay"`
	Delay               time.Duration `yaml:"-"`
}

// MatchesRule checks all rules in the conf and tries to match on them
func (c Conf) MatchesRule(r *http.Request) (Rule, string, bool) {
	ip := parseIP(r)
	ua := parseUA(r)

	for _, rule := range c.Rules {
		matchedUA, _ := regexp.Match(rule.UARegex, []byte(ua))
    
		if matchedUA && rule.Mode == "ua" {
			return rule, ua, true
		}

		if matchedUA && rule.Mode == "ip+ua" {
			return rule, ip + ua, true
		}

		if rule.Mode == "ip" {
			return rule, ip, true
		}
	}
	return Rule{}, "", false
}
```
```yaml
destination: http://greeter
rules:
    - name: vegeta-ua
      mode: ua
      pattern: "vegeta"
      delay: 1s
    
    - name: ip
      mode: ip
      delay: 200ms
```
```yaml
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
```txt
GET http://localhost:9001/
X-Forwarded-For: 192.168.1.10
User-Agent: vegeta

GET http://localhost:9001/
X-Forwarded-For: 192.168.1.20
User-Agent: vegeta

GET http://localhost:9001/
X-Forwarded-For: 192.168.1.30
User-Agent: vegeta

GET http://localhost:9001/
X-Forwarded-For: 192.168.1.40
User-Agent: vegeta

GET http://localhost:9001/
X-Forwarded-For: 192.168.1.50
User-Agent: vegeta
```
