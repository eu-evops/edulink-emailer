package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/eu-evops/edulink/pkg/cache"
	"github.com/eu-evops/edulink/pkg/cache/common"
	"github.com/eu-evops/edulink/pkg/util"
	"github.com/eu-evops/edulink/pkg/web"
	"github.com/eu-evops/edulink/pkg/worker"
)

var (
	EdulinkUsername string
	EdulinkPassword string
	MailgunApiKey   string

	Cache *cache.Cache
)

func init() {
	EdulinkUsername = os.Getenv("EDULINK_USERNAME")
	EdulinkPassword = os.Getenv("EDULINK_PASSWORD")
	MailgunApiKey = os.Getenv("MAILGUN_API_KEY")

	if EdulinkUsername == "" || EdulinkPassword == "" {
		fmt.Println("Please set EDULINK_USERNAME and EDULINK_PASSWORD environment variables")
		os.Exit(1)
	}

	if MailgunApiKey == "" {
		fmt.Println("Please set MAILGUN_API_KEY environment variable")
		os.Exit(1)
	}

	Cache = cache.New(&common.CacheOptions{
		CacheType:     common.Redis,
		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisUsername: os.Getenv("REDIS_USERNAME"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	})

	util.Cache = Cache

	if err := Cache.Initialise(); err != nil {
		panic(err)
	}
}

func main() {
	fmt.Printf("Welcome to EduLink scanner.\n")

	webserverPort := flag.Int("port", 8080, "Port to listen on")

	webServer := web.NewServer(*webserverPort)
	if err := webServer.Start(); err != nil {
		panic(err)
	}

	workerOptions := &worker.WorkerOptions{
		EdulinkUsername: EdulinkUsername,
		EdulinkPassword: EdulinkPassword,
		Cache:           Cache,
		MailgunApiKey:   MailgunApiKey,
	}
	worker := worker.NewWorker(workerOptions)
	if err := worker.Start(); err != nil {
		panic(err)
	}

}
