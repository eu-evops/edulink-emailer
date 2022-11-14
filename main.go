package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/eu-evops/edulink/pkg/cache"
	"github.com/eu-evops/edulink/pkg/cache/common"
	"github.com/eu-evops/edulink/pkg/edulink"
	mailgun "github.com/mailgun/mailgun-go/v4"
	"golang.org/x/exp/slices"
)

var (
	USERNAME        string
	PASSWORD        string
	MAILGUN_API_KEY string

	Cache *cache.Cache
)

func init() {
	USERNAME = os.Getenv("EDULINK_USERNAME")
	PASSWORD = os.Getenv("EDULINK_PASSWORD")
	MAILGUN_API_KEY = os.Getenv("MAILGUN_API_KEY")

	if USERNAME == "" || PASSWORD == "" {
		fmt.Println("Please set EDULINK_USERNAME and EDULINK_PASSWORD environment variables")
		os.Exit(1)
	}

	if MAILGUN_API_KEY == "" {
		fmt.Println("Please set MAILGUN_API_KEY environment variable")
		os.Exit(1)
	}

	Cache = cache.New(&common.CacheOptions{
		CacheType:     common.Redis,
		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisUsername: os.Getenv("REDIS_USERNAME"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
	})

	if err := Cache.Initialise(); err != nil {
		panic(err)
	}
}

func main() {
	fmt.Printf("Welcome to EduLink scanner.\n")

	alreadySeenBehaviourIDs := []string{}
	alreadySeenAchievementIDs := []string{}

	Cache.Get(context.Background(), "alreadySeenBehaviourIDs", &alreadySeenBehaviourIDs)
	Cache.Get(context.Background(), "alreadySeenAchievementIDs", &alreadySeenAchievementIDs)

	fmt.Println("Already seen behaviour IDs:", alreadySeenBehaviourIDs)
	fmt.Println("Already seen achievement IDs:", alreadySeenAchievementIDs)

	defer (func() {
		Cache.Set(&common.Item{
			Ctx:   context.Background(),
			Key:   "alreadySeenBehaviourIDs",
			Value: alreadySeenBehaviourIDs,
			TTL:   0,
		})

		Cache.Set(&common.Item{
			Ctx:   context.Background(),
			Key:   "alreadySeenAchievementIDs",
			Value: alreadySeenAchievementIDs,
			TTL:   0,
		})

		fmt.Println("Already seen behaviour IDs:", alreadySeenBehaviourIDs)
		fmt.Println("Already seen achievement IDs:", alreadySeenAchievementIDs)
	})()

	schoolDetailsReq := edulink.SchoolDetailsRequest{
		RequestBase: edulink.RequestBase{
			ID:      1,
			JsonRPC: "2.0",
			Method:  "EduLink.SchoolDetails",
		},
		Params: edulink.SchoolDetailsRequestParams{
			EstablishmentID: edulink.SCHOOL_ID,
		},
	}
	var schoolDetailsResp edulink.SchoolDetailsResponse
	err = call(schoolDetailsReq, &schoolDetailsResp)
	if err != nil {
		panic(err)
	}

	loginReq := edulink.LoginRequest{
		RequestBase: edulink.RequestBase{
			ID:      1,
			JsonRPC: "2.0",
			Method:  "EduLink.Login",
		},
		Params: edulink.LoginRequestParams{
			Username:        USERNAME,
			Password:        PASSWORD,
			EstablishmentID: edulink.SCHOOL_ID,
		},
	}
	var loginResponse edulink.LoginResponse
	err = call(loginReq, &loginResponse)
	if err != nil {
		panic(err)
	}

	achievementBehaviourLookups := edulink.AchievementBehaviourLookupsRequest{
		RequestBase: edulink.RequestBase{
			ID:        1,
			JsonRPC:   "2.0",
			Method:    "EduLink.AchievementBehaviourLookups",
			AuthToken: loginResponse.Result.AuthToken,
		},
	}
	var achievementBehaviourLookupsResponse edulink.AchievementBehaviourLookupsResponse
	err = call(achievementBehaviourLookups, &achievementBehaviourLookupsResponse)
	if err != nil {
		panic(err)
	}

	for _, child := range loginResponse.Result.Children {
		fmt.Printf("Child: %+v\n", child)

		photoReq := &edulink.LearnerPhotosRequest{
			RequestBase: edulink.RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.LearnerPhotos",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: edulink.LearnerPhotosRequestParams{
				LearnerIDs: []string{child.ID},
				Size:       256,
			},
		}
		var photoResponse edulink.LearnerPhotosResponse
		err = call(photoReq, &photoResponse)
		if err != nil {
			panic(err)
		}

		schoolReport := &edulink.SchoolReport{
			Child:       child,
			Photo:       photoResponse.Result.LearnerPhotos[0].Photo,
			School:      schoolDetailsResp.Result.Establishment,
			Behaviour:   []edulink.Behaviour{},
			Achievement: []edulink.Achievement{},
		}

		behaviourReq := edulink.BehaviourRequest{
			RequestBase: edulink.RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.Behaviour",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: edulink.BehaviourRequestParams{
				LearnerID: child.ID,
				Format:    2,
			},
		}

		var behaviourResponse edulink.BehaviourResponse
		err := call(behaviourReq, &behaviourResponse)
		if err != nil {
			panic(err)
		}

		for _, behaviour := range behaviourResponse.Result.Behaviour {
			if !slices.Contains(alreadySeenBehaviourIDs, behaviour.ID) {
				schoolReport.Behaviour = append(schoolReport.Behaviour, behaviour)
				alreadySeenBehaviourIDs = append(alreadySeenBehaviourIDs, behaviour.ID)
			}
		}

		achievementReq := edulink.AchievementRequest{
			RequestBase: edulink.RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.Achievement",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: edulink.AchievementRequestParams{
				LearnerID: child.ID,
				Format:    2,
			},
		}

		var achievementResponse edulink.AchievementResponse
		err = call(achievementReq, &achievementResponse)
		if err != nil {
			panic(err)
		}

		for _, achievement := range achievementResponse.Result.Achievement {
			if !slices.Contains(alreadySeenAchievementIDs, achievement.ID) {
				schoolReport.Achievement = append(schoolReport.Achievement, achievement)
				alreadySeenAchievementIDs = append(alreadySeenAchievementIDs, achievement.ID)
			}
		}
		involvedTeachers := []edulink.Employee{}
		involvedTeacherIDs := []string{}

		for _, employee := range behaviourResponse.Result.Employees {
			if !slices.Contains(involvedTeacherIDs, employee.ID) {
				involvedTeachers = append(involvedTeachers, employee)
				involvedTeacherIDs = append(involvedTeacherIDs, employee.ID)
			}
		}

		for _, employee := range achievementResponse.Result.Employees {
			if !slices.Contains(involvedTeacherIDs, employee.ID) {
				involvedTeachers = append(involvedTeachers, employee)
				involvedTeacherIDs = append(involvedTeacherIDs, employee.ID)
			}
		}

		teachersPhotosRequest := &edulink.TeacherPhotosRequest{
			RequestBase: edulink.RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.TeacherPhotos",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: edulink.TeacherPhotosRequestParams{
				EmployeeIDs: involvedTeacherIDs,
				Size:        256,
			},
		}
		var teachersPhotosResponse edulink.TeacherPhotosResponse
		err = call(teachersPhotosRequest, &teachersPhotosResponse)
		if err != nil {
			panic(err)
		}

		if len(schoolReport.Behaviour) == 0 || len(schoolReport.Achievement) == 0 {
			fmt.Printf("There are no new achievements or behaviours for %s to report on.\n", child.Forename)
			continue
		}

		if os.Getenv("SEND_EMAIL") != "true" {
			fmt.Println("Not sending email because SEND_EMAIL is not set to true")
			continue
		}

		fmt.Printf("Sending email to %s\n", MAILGUN_API_KEY)
		mg := mailgun.NewMailgun("evops.eu", MAILGUN_API_KEY)

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
				for _, employee := range involvedTeachers {
					if employee.ID == teacherID {
						return &employee
					}
				}
				return nil
			},
			"teacherPhoto": func(teacherID string) *string {
				for _, teacherPhoto := range teachersPhotosResponse.Result.TeacherPhotos {
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
				for _, activity := range achievementBehaviourLookupsResponse.Result.AchievementTypes {
					if activity.ID == activityID {
						return &activity.Description
					}
				}
				return nil
			},
			"behaviour": func(behaviourID string) *string {
				for _, behaviour := range achievementBehaviourLookupsResponse.Result.BehaviourTypes {
					if behaviour.ID == behaviourID {
						return &behaviour.Description
					}
				}
				return nil
			},
		}

		reportTemplate := []string{"templates/edulink.schoolreport.go.tmpl"}
		t := template.Must(template.New("edulink.schoolreport.go.tmpl").Funcs(fmap).ParseFiles(reportTemplate...))

		fmt.Printf("Executing a template: %+v\n", t)

		f, _ := os.OpenFile(fmt.Sprintf("%s.report.html", strings.ToLower(child.Forename)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		defer f.Close()

		style, _ := ioutil.ReadFile("templates/style.css")
		schoolReportViewData := &SchoolReportViewData{
			SchoolReport: *schoolReport,
			Style:        string(style),
		}

		err = t.Execute(f, schoolReportViewData)
		if err != nil {
			panic(err)
		}

		sender := "EduLink <edulink@evops.eu>"
		subject := fmt.Sprintf("EduLink School Report: %s", child.Forename)

		recipients := os.Getenv("EMAIL_RECIPIENTS")
		if recipients == "" {
			fmt.Println("No recipients specified, skipping email")
			continue
		}

		recipientsList := strings.Split(recipients, ",")

		message := mg.NewMessage(sender, subject, "html", recipientsList...)

		template.New("schoolReport")

		reportHTML, _ := ioutil.ReadFile(fmt.Sprintf("%s.report.html", strings.ToLower(child.Forename)))
		message.SetHtml(string(reportHTML))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		// Send the message with a 10 second timeout
		resp, id, err := mg.Send(ctx, message)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("ID: %s Resp: %s\n", id, resp)
	}
}
