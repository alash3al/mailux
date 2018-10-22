package main

import (
	"flag"
	"log"

	"github.com/go-redis/redis"
)

var (
	flagSMTPAddr   = flag.String("smtp", ":25025", "the smtp (inbox) address to listen on")
	flagHTTPAddr   = flag.String("http", ":5050", "the http server address to listen on")
	flagSPFChecker = flag.Bool("spf", false, "whether to enable SPF checking or not")
	flagRedis      = flag.String("redis", "redis://localhost:6379/1", "redis url")
	flagDefaultTTL = flag.Int("ttl", 30, "the time-to-live of each auth-inbox in seconds")
	flagDomain     = flag.String("domain", "local.host", "the default domain name of the server")
)

var (
	redisClient *redis.Client

	redisKeyPrefix = func(k string) string {
		return "mailux:" + k
	}
)

func init() {
	flag.Parse()

	opt, err := redis.ParseURL(*flagRedis)
	if err != nil {
		log.Fatal("[Redis]", " ", err.Error())
	}

	redisClient = redis.NewClient(opt)
	if _, err := redisClient.Ping().Result(); err != nil {
		log.Fatal("[Redis]", " ", err.Error())
	}
}
