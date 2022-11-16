package worker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/eu-evops/edulink/pkg/cache"
	"github.com/eu-evops/edulink/pkg/cache/common"
	"github.com/eu-evops/edulink/pkg/edulink"
	"github.com/eu-evops/edulink/pkg/mailer"
	"github.com/eu-evops/edulink/pkg/util"

	"golang.org/x/exp/slices"
)

const (
	Day     = 24 * time.Hour
	Month   = 30 * Day
	Year    = 365 * Day
	Decade  = 10 * Year
	Century = 10 * Decade
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
}

func NewWorker(o *WorkerOptions) *Worker {
	return &Worker{
		cache:           o.Cache,
		edulinkUsername: o.EdulinkUsername,
		edulinkPassword: o.EdulinkPassword,
	}
}

func (w *Worker) Start() error {
	mailer := mailer.NewMailer(&mailer.MailerOptions{
		MailgunApiKey: w.mailgunApiKey,
	})

	alreadySeenBehaviourIDs := []string{}
	alreadySeenAchievementIDs := []string{}

	w.cache.Get(context.Background(), "alreadySeenBehaviourIDs", &alreadySeenBehaviourIDs)
	w.cache.Get(context.Background(), "alreadySeenAchievementIDs", &alreadySeenAchievementIDs)

	fmt.Println("Already seen behaviour IDs:", alreadySeenBehaviourIDs)
	fmt.Println("Already seen achievement IDs:", alreadySeenAchievementIDs)

	defer (func() {
		w.cache.Set(&common.Item{
			Ctx:   context.Background(),
			Key:   "alreadySeenBehaviourIDs",
			Value: alreadySeenBehaviourIDs,
			TTL:   Century,
		})

		w.cache.Set(&common.Item{
			Ctx:   context.Background(),
			Key:   "alreadySeenAchievementIDs",
			Value: alreadySeenAchievementIDs,
			TTL:   Century,
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
	if err := util.Call(context.Background(), schoolDetailsReq, &schoolDetailsResp); err != nil {
		panic(err)
	}

	loginReq := edulink.LoginRequest{
		RequestBase: edulink.RequestBase{
			ID:      1,
			JsonRPC: "2.0",
			Method:  "EduLink.Login",
		},
		Params: edulink.LoginRequestParams{
			Username:        w.edulinkUsername,
			Password:        w.edulinkPassword,
			EstablishmentID: edulink.SCHOOL_ID,
		},
	}
	var loginResponse edulink.LoginResponse
	if err := util.Call(context.Background(), loginReq, &loginResponse); err != nil {
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
	if err := util.Call(context.Background(), achievementBehaviourLookups, &achievementBehaviourLookupsResponse); err != nil {
		panic(err)
	}

	mailer.SetAchievementTypes(achievementBehaviourLookupsResponse.Result.AchievementTypes)
	mailer.SetBehaviourTypes(achievementBehaviourLookupsResponse.Result.BehaviourTypes)

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
		if err := util.Call(context.Background(), photoReq, &photoResponse); err != nil {
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
		err := util.Call(context.Background(), behaviourReq, &behaviourResponse)
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
		if err := util.Call(context.Background(), achievementReq, &achievementResponse); err != nil {
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
		if err := util.Call(context.Background(), teachersPhotosRequest, &teachersPhotosResponse); err != nil {
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

		mailer.Send(schoolReport)
	}

	return nil
}
