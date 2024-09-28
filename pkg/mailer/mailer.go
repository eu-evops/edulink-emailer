package mailer

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eu-evops/edulink/pkg/edulink"
	"github.com/mailgun/mailgun-go/v4"
)

type Mailer struct {
	mailgunApiKey string
	mailGun       *mailgun.MailgunImpl

	teachers      []edulink.Employee
	teacherPhotos []edulink.TeacherPhoto

	behaviourTypes   []edulink.BehaviourType
	achievementTypes []edulink.AchievementType
}

type MailerOptions struct {
	MailgunApiKey string

	Teachers         []edulink.Employee
	TeacherPhotos    []edulink.TeacherPhoto
	BehaviourTypes   []edulink.BehaviourType
	AchievementTypes []edulink.AchievementType
}

func NewMailer(o *MailerOptions) *Mailer {
	return &Mailer{
		mailgunApiKey:    o.MailgunApiKey,
		mailGun:          mailgun.NewMailgun("evops.eu", o.MailgunApiKey),
		teachers:         o.Teachers,
		teacherPhotos:    o.TeacherPhotos,
		behaviourTypes:   o.BehaviourTypes,
		achievementTypes: o.AchievementTypes,
	}
}

func (m *Mailer) Send(schoolReport *edulink.SchoolReport, mail string) {
	sender := "EduLink <edulink@evops.eu>"
	subject := fmt.Sprintf("EduLink School Report: %s", schoolReport.Child.Forename)

	recipients := os.Getenv("EMAIL_RECIPIENTS")
	if recipients == "" {
		fmt.Println("No recipients specified, skipping email")
		return
	}

	recipientsList := strings.Split(recipients, ",")

	message := m.mailGun.NewMessage(sender, subject, "html", recipientsList...)

	template.New("schoolReport")

	message.SetHtml(mail)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	resp, id, err := m.mailGun.Send(ctx, message)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s Resp: %s\n", id, resp)
}
