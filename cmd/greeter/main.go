package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	p = flag.String("port", "9000", "server port")
)

func main() {
	http.HandleFunc("/", greetHandler)
	panic(http.ListenAndServe(":"+*p, nil))
}

func greetHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Hello there %s\n", req.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Hello there` + "\n"))
	ioutil.ReadAll(req.Body)
	req.Body.Close()
}
