package main

import (
	"encoding/base64"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func initHTTP() {
	e := echo.New()
	e.HideBanner = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	e.Use(middleware.BodyLimit("100K"))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"success": true,
			"message": "let's reverse the login process",
		})
	})

	/**
	 * Get the information of an inbox (temp-inbox)
	 */
	e.GET("/inbox/info/:inboxId", func(c echo.Context) error {
		id := c.Param("inboxId")
		return c.JSON(200, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"uuid":    id,
				"inbox":   id + "@" + *flagDomain,
				"status":  strings.ToLower(redisClient.Get(redisKeyPrefix(id+":status")).Val()) == "done",
				"email":   redisClient.Get(redisKeyPrefix(id + ":email")).Val(),
				"expires": redisClient.Get(redisKeyPrefix(id + ":expires")).Val(),
			},
		})
	})

	/**
	 * Generates a temp-inbox for the specified email
	 */
	e.GET("/inbox/generate/:email", func(c echo.Context) error {
		id := uuid.New().String()
		if c.QueryParam("suffix") != "" {
			id += "-" + c.QueryParam("suffix")
		}
		id = base64.RawStdEncoding.EncodeToString([]byte(id))
		ttl, _ := strconv.Atoi(c.QueryParam("ttl"))
		if ttl < 1 {
			ttl = *flagDefaultTTL
		}
		exp := time.Now().Unix() + int64(ttl)
		email := strings.ToLower(c.Param("email"))
		redisDur := time.Second * 3600

		redisClient.Set(redisKeyPrefix(email+":inbox"), id, redisDur).Val()
		redisClient.Set(redisKeyPrefix(id+":status"), "pending", redisDur).Val()
		redisClient.Set(redisKeyPrefix(id+":email"), email, redisDur).Val()
		redisClient.Set(redisKeyPrefix(id+":expires"), exp, redisDur).Val()

		return c.JSON(200, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"uuid":    id,
				"inbox":   id + "@" + *flagDomain,
				"status":  false,
				"email":   email,
				"expires": exp,
			},
		})
	})

	log.Fatal(e.Start(*flagHTTPAddr))
}
