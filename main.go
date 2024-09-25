package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/eu-evops/edulink/pkg/cache"
	"github.com/eu-evops/edulink/pkg/cache/common"
	"github.com/eu-evops/edulink/pkg/edulink"
	"github.com/eu-evops/edulink/pkg/web"
	"github.com/eu-evops/edulink/pkg/worker"
)

var (
	EdulinkUsername string
	EdulinkPassword string
	MailgunApiKey   string

	appCache *cache.Cache
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

	appCache = cache.New(&common.CacheOptions{
		CacheType:     common.Redis,
		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisUsername: os.Getenv("REDIS_USERNAME"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	})

	edulink.Cache = appCache

	if err := appCache.Initialise(); err != nil {
		panic(err)
	}
}

func main() {
	fmt.Printf("Welcome to EduLink scanner.\n")

	webserverEnabled := flag.Bool("webserver", false, "Enable webserver")
	webserverPort := flag.Int("port", 8080, "Port to listen on")

	flag.Parse()

	webServer := web.NewServer(*webserverPort)
	if err := webServer.Start(); err != nil {
		panic(err)
	}

	workerOptions := &worker.WorkerOptions{
		EdulinkUsername: EdulinkUsername,
		EdulinkPassword: EdulinkPassword,
		Cache:           appCache,
		MailgunApiKey:   MailgunApiKey,
	}

	worker := worker.NewWorker(workerOptions)
	if err := worker.Start(); err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	if *webserverEnabled {
		wg.Add(1)
		fmt.Println("Webserver enabled, listening on port", *webserverPort)
		s := make(chan os.Signal, 1)
		signal.Notify(s, os.Interrupt)

		go func() {
			<-s
			wg.Done()
			webServer.Stop()
		}()

		wg.Wait()
	}
}
