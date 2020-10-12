package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
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

type blocked struct {
	r    *http.Request
	w    *http.ResponseWriter
	rule Rule
	key  string
}

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	flag.Parse()

	proxyConf = ParseConf()

	proxy := httputil.NewSingleHostReverseProxy(proxyConf.Origin)

	// http.HandleFunc("/", proxy.ServeHTTP)

	http.HandleFunc("/", limiter(proxy.ServeHTTP))

	go func() {
		for l := range logCh {
			log.Println(l)
		}
	}()

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
			ioutil.ReadAll(r.Body)
			r.Body.Close()
			w.WriteHeader(http.StatusTooManyRequests)
			// w.Write([]byte(fmt.Sprintf("Not Ready: must wait %v between requests\n", rule.Delay)))
			logCh <- fmt.Sprintf("blocked key=%s rule=%s\n", key, rule.Name)
			return
		}
		kvDB.Set(key, true, rule.Delay)
		next.ServeHTTP(w, r)
	}
}
