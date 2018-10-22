package main

import (
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/alash3al/go-mailbox"
	"github.com/zaccone/spf"
)

func initSMTP() {
	smtp.HandleFunc("*@"+*flagDomain, func(req *smtp.Envelope) error {
		ipAddr, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			return err
		}

		fromEmail, toEmail := req.MessageFrom, req.MessageTo
		fromEmail = strings.ToLower(fromEmail)
		_, fromDomain, err := smtp.SplitAddress(fromEmail)
		if err != nil {
			return err
		}

		res, _, err := spf.CheckHost(net.ParseIP(ipAddr), fromDomain, fromEmail)
		if (err != nil || res != spf.Pass) && *flagSPFChecker {
			return errors.New("ERR_CHEATING_DETECTED")
		}

		inbox := strings.Split(toEmail, "@")[0]
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
	})

	log.Fatal(smtp.ListenAndServe(*flagSMTPAddr, nil))
}
