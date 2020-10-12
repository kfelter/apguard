package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

var (
	p         = flag.String("port", "80", "port")
	rdsURL    = flag.String("ru", os.Getenv("REDIS_URL"), "redis url for distributed mode default to env var REDIS_URL")
	yamlFile  = flag.String("conf", os.Getenv("PROXY_CONF"), "configuration file for proxy")
	logCh     = make(chan string, 10000)
	kvDB      DB
	proxyConf Conf
)

func main() {
	flag.Parse()

	proxyConf = ParseConf()

	proxy := httputil.NewSingleHostReverseProxy(proxyConf.Origin)

	http.HandleFunc("/", limiter(proxy.ServeHTTP))

	go logger()

	log.Fatal(http.ListenAndServe(":"+*p, nil))
}

func limiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rule, key, found := proxyConf.MatchesRule(r)
		if !found {
			next.ServeHTTP(w, r)
			return
		}

		if _, err := kvDB.Get(key); err == nil {
			w.WriteHeader(http.StatusTooManyRequests)
			// w.Write([]byte(fmt.Sprintf("Not Ready: must wait %v between requests\n", rule.Delay)))
			logCh <- fmt.Sprintf("blocked key=%s rule=%s\n", key, rule.Name)
			return
		}
		kvDB.Set(key, true, rule.Delay)
		next.ServeHTTP(w, r)
	}
}

func logger() {
	for l := range logCh {
		log.Println(l)
	}
}
