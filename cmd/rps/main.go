package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	memdb "github.com/felts94/go-cache"
	"github.com/go-redis/redis/v8"
)

var (
	dest   = flag.String("dest", "http://localhost:9000/", "destination url")
	p      = flag.String("port", "9001", "port")
	dbFile = flag.String("sqlite", ":memory:", "sqlite file name")
	defExp = flag.String("exp", "1s", "default expiration")
	rdsURL = flag.String("ru", os.Getenv("REDIS_URL"), "redis url for distributed mode default to env var REDIS_URL")
	kvDB   DB
)

func main() {
	flag.Parse()
	d, _ := time.ParseDuration(*defExp)
	if *rdsURL == "" {
		kvDB = memdb.New(d, 10*time.Minute)
	} else {
		var err error
		kvDB, err = NewRC(*rdsURL)
		if err != nil {
			panic(err)
		}
	}

	origin, _ := url.Parse(*dest)

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ipOnly := parseIP(r.RemoteAddr)

		if _, err := kvDB.Get(ipOnly); err == nil {
			ioutil.ReadAll(r.Body)
			r.Body.Close()
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(fmt.Sprintf("Not Ready: must wait %v between requests\n", *defExp)))
			fmt.Printf("blocked ip=%s(%s) ua=%s\n", r.RemoteAddr, ipOnly, r.Header.Get("User-Agent"))
			return
		}
		kvDB.Set(ipOnly, true, d)
		proxy.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":"+*p, nil))
}

func parseIP(s string) string {
	ip, _, err := net.SplitHostPort(s)
	if err != nil {
		return s
	}
	return ip
}

type DB interface {
	Set(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, error)
}

type redisClient struct {
	standaloneClient *redis.Client
	clusterClient    *redis.ClusterClient
}

func NewRC(url string) (*redisClient, error) {
	sp := strings.Split(url, ":")
	if len(sp) < 2 {
		return nil, fmt.Errorf("malformed redis url: %s", url)
	}
	if sp[0] == "redis" {
		opts, err := redis.ParseURL(url)
		if err != nil {
			return nil, fmt.Errorf("malformed redis url opts: %s: %v", url, err)
		}
		c := redis.NewClient(opts)
		if err := c.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("could not connect to redis url: %s: %v", url, err)
		}
		return &redisClient{
			standaloneClient: c,
		}, nil
	}

	if sp[0] == "cluster" {
		c := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{"redis:" + sp[1]},
		})

		if err := c.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("could not connect to %s: %v", url, err)
		}
	}

	return nil, fmt.Errorf("not implemented %s", url)
}

func (rc *redisClient) Set(key string, value interface{}, ttl time.Duration) {
	if rc.standaloneClient != nil {
		rc.standaloneClient.Set(context.Background(), key, value, ttl)
	}
	if rc.clusterClient != nil {
		rc.clusterClient.Set(context.Background(), key, value, ttl)
	}
}

func (rc *redisClient) Get(key string) (interface{}, error) {
	if rc.standaloneClient != nil {
		return rc.standaloneClient.Get(context.Background(), key).Result()
	}
	if rc.clusterClient != nil {
		return rc.clusterClient.Get(context.Background(), key).Result()
	}
	return nil, fmt.Errorf("rcs are nil")
}
