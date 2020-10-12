package main

import (
	"io/ioutil"
	"log"
	"net/url"
	"time"

	memdb "github.com/felts94/go-cache"
	"gopkg.in/yaml.v2"
)

// Conf ...
type Conf struct {
	Destination string `yaml:"destination"`
	Origin      *url.URL
	Rules       []Rule `yaml:"rules"`
}

// Rule ...
type Rule struct {
	Name                string        `yaml:"name"`
	Mode                string        `yaml:"mode"`
	UARegex             string        `yaml:"pattern"`
	TimeBetweenRequests string        `yaml:"delay"`
	Delay               time.Duration `yaml:"-"`
}

// ParseConf ...
func ParseConf() Conf {
	var err error
	conf := Conf{}
	if *yamlFile != "" {
		b, err := ioutil.ReadFile(*yamlFile)
		if err != nil {
			log.Fatal("could not read yaml file: ", *yamlFile, ":", err)
		}
		err = yaml.Unmarshal(b, &conf)
		if err != nil {
			log.Fatal("Could not unmarshal yaml file", *yamlFile, err)
		}
	}

	for i := range conf.Rules {
		conf.Rules[i].Delay, err = time.ParseDuration(conf.Rules[i].TimeBetweenRequests)
		if err != nil {
			log.Fatal("could not parse duration", conf.Rules[i].TimeBetweenRequests, err)
		}
	}

	if *rdsURL == "" {
		kvDB = memdb.New(1*time.Second, 5*time.Minute)
	} else {
		var err error
		kvDB, err = NewRC(*rdsURL)
		if err != nil {
			panic(err)
		}
	}

	conf.Origin, err = url.Parse(conf.Destination)
	if err != nil {
		log.Fatal("Could not parse destination url", conf.Destination, err)
	}

	return conf
}
