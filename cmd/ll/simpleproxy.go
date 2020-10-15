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
