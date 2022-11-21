package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
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
	template      *template.Template

	teachers      []edulink.Employee
	teacherPhotos []edulink.TeacherPhoto

	behaviourTypes   []edulink.BehaviourType
	achievementTypes []edulink.AchievementType

	templatesPrepared bool
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

func (m *Mailer) SetAchievementTypes(achievementTypes []edulink.AchievementType) {
	m.achievementTypes = achievementTypes
}
func (m *Mailer) SetBehaviourTypes(behaviourTypes []edulink.BehaviourType) {
	m.behaviourTypes = behaviourTypes
}

func (m *Mailer) updateTeacherPhotos(schoolReport *edulink.SchoolReport) {
	for _, slTeacherPhoto := range schoolReport.TeacherPhotos {
		found := false
		for _, mTeacherPhoto := range m.teacherPhotos {
			if mTeacherPhoto.ID == slTeacherPhoto.ID {
				found = true
			}
		}
		if !found {
			m.teacherPhotos = append(m.teacherPhotos, slTeacherPhoto)
		}
	}
}

func (m *Mailer) updateTeachers(schoolReport *edulink.SchoolReport) {
	for _, slTeacher := range schoolReport.Teachers {
		found := false
		for _, mTeacher := range m.teachers {
			if mTeacher.ID == slTeacher.ID {
				found = true
			}
		}
		if !found {
			m.teachers = append(m.teachers, slTeacher)
		}
	}
}

func (m *Mailer) prepareTemplates(schoolReport *edulink.SchoolReport) {
	m.updateTeacherPhotos(schoolReport)
	m.updateTeachers(schoolReport)

	if m.templatesPrepared {
		return
	}

	fmap := template.FuncMap{
		"has": func(slice []string, v string) bool {
			has := false
			for _, item := range slice {
				if item == v {
					return true
				}
			}
			return has
		},
		"json": func(v interface{}) string {
			bytes, _ := json.MarshalIndent(v, "", "  ")
			return string(bytes)
		},
		"pluralize": func(val int, text string) string {
			if val == 1 {
				return fmt.Sprintf("%d %s", val, text)
			}
			return fmt.Sprintf("%d %ss", val, text)
		},
		"teacher": func(teacherID string) *edulink.Employee {
			for _, employee := range m.teachers {
				if employee.ID == teacherID {
					return &employee
				}
			}
			return nil
		},
		"teacherPhoto": func(teacherID string) *string {
			for _, teacherPhoto := range m.teacherPhotos {
				if teacherPhoto.ID == teacherID {
					return &teacherPhoto.Photo
				}
			}
			return nil
		},
		"wrap": func(pairs ...interface{}) map[string]interface{} {
			m := make(map[string]interface{})
			for i := 0; i < len(pairs); i += 2 {
				m[pairs[i].(string)] = pairs[i+1]
			}

			return m
		},
		"activity": func(activityID string) *string {
			for _, activity := range m.achievementTypes {
				if activity.ID == activityID {
					return &activity.Description
				}
			}
			return nil
		},
		"behaviour": func(behaviourID string) *string {
			for _, behaviour := range m.behaviourTypes {
				if behaviour.ID == behaviourID {
					return &behaviour.Description
				}
			}
			return nil
		},
	}

	reportTemplate := []string{"templates/edulink.schoolreport.go.tmpl"}
	m.template = template.Must(template.New("edulink.schoolreport.go.tmpl").Funcs(fmap).ParseFiles(reportTemplate...))
	m.templatesPrepared = true
}

func (m *Mailer) Send(schoolReport *edulink.SchoolReport) {
	m.prepareTemplates(schoolReport)

	style, err := ioutil.ReadFile("templates/style.css")
	if err != nil {
		panic(err)
	}

	schoolReportViewData := &SchoolReportViewData{
		SchoolReport: *schoolReport,
		Style:        template.CSS(style),
	}

	var tmpl bytes.Buffer
	if err := m.template.Execute(&tmpl, schoolReportViewData); err != nil {
		panic(err)
	}

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

	message.SetHtml(tmpl.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	resp, id, err := m.mailGun.Send(ctx, message)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s Resp: %s\n", id, resp)
}
