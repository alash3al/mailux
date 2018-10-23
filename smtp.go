package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/alash3al/go-smtpsrv"
	"github.com/zaccone/spf"
)

func initSMTP() {
	handler := func(req *smtpsrv.Request) error {
		fromEmail, toEmail := req.From, req.To
		fromEmail = strings.ToLower(fromEmail)

		if spf.Pass != req.SPFResult && *flagSPFChecker {
			return errors.New("ERR_CHEATING_DETECTED")
		}

		inbox := strings.Split(toEmail[0], "@")[0]
		expires, _ := redisClient.Get(redisKeyPrefix(inbox + ":expires")).Int64()
		if time.Now().Unix() >= expires {
			return errors.New("ERR_EXPIRED")
		}

		if strings.ToLower(fromEmail) != redisClient.Get(redisKeyPrefix(inbox+":email")).Val() {
			return errors.New("ERR_MISS_MATCH")
		}

		if strings.ToLower(redisClient.Get(redisKeyPrefix(inbox+":status")).Val()) == "pending" {
			ttl := redisClient.TTL(redisKeyPrefix(inbox + ":status")).Val()
			redisClient.Set(redisKeyPrefix(inbox+":status"), "done", ttl).Val()
		}

		return nil
	}

	srv := &smtpsrv.Server{
		Addr:        *flagSMTPAddr,
		MaxBodySize: 1024 * 1024, // 1 MB
		Handler:     handler,
	}

	log.Fatal(srv.ListenAndServe())
}
