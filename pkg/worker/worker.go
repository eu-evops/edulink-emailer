package worker

import (
	"fmt"
	"os"

	"github.com/eu-evops/edulink/pkg/cache"
	"github.com/eu-evops/edulink/pkg/edulink"
	"github.com/eu-evops/edulink/pkg/mailer"
)

type Worker struct {
	cache           *cache.Cache
	edulinkUsername string
	edulinkPassword string
	mailgunApiKey   string
}

type WorkerOptions struct {
	Cache           *cache.Cache
	EdulinkUsername string
	EdulinkPassword string
	MailgunApiKey   string
}

func NewWorker(o *WorkerOptions) *Worker {
	return &Worker{
		cache:           o.Cache,
		edulinkUsername: o.EdulinkUsername,
		edulinkPassword: o.EdulinkPassword,
		mailgunApiKey:   o.MailgunApiKey,
	}
}

func (w *Worker) Start() error {

	if os.Getenv("SEND_EMAIL") != "true" {
		fmt.Println("Not sending email because SEND_EMAIL is not set to true")
		return nil
	}

	mailer := mailer.NewMailer(&mailer.MailerOptions{
		MailgunApiKey: w.mailgunApiKey,
	})

	reporter := edulink.NewReporter(&edulink.ReporterOptions{
		Username: w.edulinkUsername,
		Password: w.edulinkPassword,
		Cache:    w.cache,
	})

	schoolReports := reporter.Prepare(nil)

	for _, report := range *schoolReports {
		if len(report.Achievement) > 0 || len(report.Behaviour) > 0 {
			mailer.Send(&report, reporter.Generate(&report))
		}
	}

	return nil
}
