package edulink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"slices"
	"time"

	"github.com/eu-evops/edulink/pkg/cache"
	"github.com/eu-evops/edulink/pkg/cache/common"
)

const (
	Day     = 24 * time.Hour
	Month   = 30 * Day
	Year    = 365 * Day
	Decade  = 10 * Year
	Century = 10 * Decade
)

type Reporter struct {
	options *ReporterOptions

	teacherPhotos    []TeacherPhoto
	teachers         []Employee
	behaviourTypes   []BehaviourType
	achievementTypes []AchievementType

	templatesPrepared bool
	template          *template.Template

	cache *cache.Cache
}

type ReporterOptions struct {
	Username string
	Password string

	Cache *cache.Cache
}

func NewReporter(o *ReporterOptions) *Reporter {
	return &Reporter{
		options:          o,
		teacherPhotos:    []TeacherPhoto{},
		teachers:         []Employee{},
		behaviourTypes:   []BehaviourType{},
		achievementTypes: []AchievementType{},
	}
}

func (r *Reporter) SetAchievementTypes(achievementTypes []AchievementType) {
	r.achievementTypes = achievementTypes
}
func (r *Reporter) SetBehaviourTypes(behaviourTypes []BehaviourType) {
	r.behaviourTypes = behaviourTypes
}

func (r *Reporter) updateTeacherPhotos(schoolReport *SchoolReport) {
	for _, slTeacherPhoto := range schoolReport.TeacherPhotos {
		found := false
		for _, mTeacherPhoto := range r.teacherPhotos {
			if mTeacherPhoto.ID == slTeacherPhoto.ID {
				found = true
			}
		}
		if !found {
			r.teacherPhotos = append(r.teacherPhotos, slTeacherPhoto)
		}
	}
}

func (r *Reporter) updateTeachers(schoolReport *SchoolReport) {
	for _, slTeacher := range schoolReport.Teachers {
		found := false
		for _, mTeacher := range r.teachers {
			if mTeacher.ID == slTeacher.ID {
				found = true
			}
		}
		if !found {
			r.teachers = append(r.teachers, slTeacher)
		}
	}
}

func (r *Reporter) prepareTemplates(schoolReport *SchoolReport) {
	r.updateTeacherPhotos(schoolReport)
	r.updateTeachers(schoolReport)

	if r.templatesPrepared {
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
		"teacher": func(teacherID string) *Employee {
			for _, employee := range r.teachers {
				if employee.ID == teacherID {
					return &employee
				}
			}
			return nil
		},
		"teacherPhoto": func(teacherID string) *string {
			for _, teacherPhoto := range r.teacherPhotos {
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
			for _, activity := range r.achievementTypes {
				if activity.ID == activityID {
					return &activity.Description
				}
			}
			return nil
		},
		"behaviour": func(behaviourID string) *string {
			for _, behaviour := range r.behaviourTypes {
				if behaviour.ID == behaviourID {
					return &behaviour.Description
				}
			}
			return nil
		},
	}

	reportTemplate := []string{"templates/edulink.schoolreport.go.tmpl"}
	r.template = template.Must(template.New("edulink.schoolreport.go.tmpl").Funcs(fmap).ParseFiles(reportTemplate...))
	r.templatesPrepared = true
}

func (r *Reporter) Prepare() *[]SchoolReport {

	schoolReports := []SchoolReport{}
	alreadySeenBehaviourIDs := []string{}
	alreadySeenAchievementIDs := []string{}

	r.options.Cache.Get(context.Background(), "alreadySeenBehaviourIDs", &alreadySeenBehaviourIDs)
	r.options.Cache.Get(context.Background(), "alreadySeenAchievementIDs", &alreadySeenAchievementIDs)

	fmt.Println("Already seen behaviour IDs:", alreadySeenBehaviourIDs)
	fmt.Println("Already seen achievement IDs:", alreadySeenAchievementIDs)

	defer (func() {
		r.options.Cache.Set(&common.Item{
			Ctx:   context.Background(),
			Key:   "alreadySeenBehaviourIDs",
			Value: alreadySeenBehaviourIDs,
			TTL:   Century,
		})

		r.options.Cache.Set(&common.Item{
			Ctx:   context.Background(),
			Key:   "alreadySeenAchievementIDs",
			Value: alreadySeenAchievementIDs,
			TTL:   Century,
		})

		fmt.Println("Already seen behaviour IDs:", alreadySeenBehaviourIDs)
		fmt.Println("Already seen achievement IDs:", alreadySeenAchievementIDs)
	})()

	loginReq := LoginRequest{
		RequestBase: RequestBase{
			ID:      1,
			JsonRPC: "2.0",
			Method:  "EduLink.Login",
		},
		Params: LoginRequestParams{
			Username:        r.options.Username,
			Password:        r.options.Password,
			EstablishmentID: SCHOOL_ID,
		},
	}

	var loginResponse LoginResponse
	if err := Call(context.Background(), loginReq, &loginResponse); err != nil {
		panic(err)
	}

	schoolDetailsReq := SchoolDetailsRequest{
		RequestBase: RequestBase{
			ID:      1,
			JsonRPC: "2.0",
			Method:  "EduLink.SchoolDetails",
		},
		Params: SchoolDetailsRequestParams{
			EstablishmentID: SCHOOL_ID,
		},
	}
	var schoolDetailsResp SchoolDetailsResponse
	if err := Call(context.Background(), schoolDetailsReq, &schoolDetailsResp); err != nil {
		panic(err)
	}

	achievementBehaviourLookups := AchievementBehaviourLookupsRequest{
		RequestBase: RequestBase{
			ID:        1,
			JsonRPC:   "2.0",
			Method:    "EduLink.AchievementBehaviourLookups",
			AuthToken: loginResponse.Result.AuthToken,
		},
	}
	var achievementBehaviourLookupsResponse AchievementBehaviourLookupsResponse
	if err := Call(context.Background(), achievementBehaviourLookups, &achievementBehaviourLookupsResponse); err != nil {
		panic(err)
	}

	r.SetAchievementTypes(achievementBehaviourLookupsResponse.Result.AchievementTypes)
	r.SetBehaviourTypes(achievementBehaviourLookupsResponse.Result.BehaviourTypes)

	for _, child := range loginResponse.Result.Children {
		fmt.Printf("Child: %+v\n", child)

		photoReq := &LearnerPhotosRequest{
			RequestBase: RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.LearnerPhotos",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: LearnerPhotosRequestParams{
				LearnerIDs: []string{child.ID},
				Size:       256,
			},
		}
		var photoResponse LearnerPhotosResponse
		if err := Call(context.Background(), photoReq, &photoResponse); err != nil {
			panic(err)
		}

		schoolReport := &SchoolReport{
			Child:         child,
			Photo:         photoResponse.Result.LearnerPhotos[0].Photo,
			School:        schoolDetailsResp.Result.Establishment,
			Behaviour:     []Behaviour{},
			Achievement:   []Achievement{},
			Teachers:      []Employee{},
			TeacherPhotos: []TeacherPhoto{},
		}

		behaviourReq := BehaviourRequest{
			RequestBase: RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.Behaviour",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: BehaviourRequestParams{
				LearnerID: child.ID,
				Format:    2,
			},
		}

		var behaviourResponse BehaviourResponse
		err := Call(context.Background(), behaviourReq, &behaviourResponse)
		if err != nil {
			panic(err)
		}

		for _, behaviour := range behaviourResponse.Result.Behaviour {
			if !slices.Contains(alreadySeenBehaviourIDs, behaviour.ID) {
				schoolReport.Behaviour = append(schoolReport.Behaviour, behaviour)
				alreadySeenBehaviourIDs = append(alreadySeenBehaviourIDs, behaviour.ID)
			}
		}

		achievementReq := AchievementRequest{
			RequestBase: RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.Achievement",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: AchievementRequestParams{
				LearnerID: child.ID,
				Format:    2,
			},
		}

		var achievementResponse AchievementResponse
		if err := Call(context.Background(), achievementReq, &achievementResponse); err != nil {
			panic(err)
		}

		for _, achievement := range achievementResponse.Result.Achievement {
			if !slices.Contains(alreadySeenAchievementIDs, achievement.ID) {
				schoolReport.Achievement = append(schoolReport.Achievement, achievement)
				alreadySeenAchievementIDs = append(alreadySeenAchievementIDs, achievement.ID)
			}
		}
		involvedTeachers := []Employee{}
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

		teachersPhotosRequest := &TeacherPhotosRequest{
			RequestBase: RequestBase{
				ID:        1,
				JsonRPC:   "2.0",
				Method:    "EduLink.TeacherPhotos",
				AuthToken: loginResponse.Result.AuthToken,
			},
			Params: TeacherPhotosRequestParams{
				EmployeeIDs: involvedTeacherIDs,
				Size:        256,
			},
		}
		var teachersPhotosResponse TeacherPhotosResponse
		if err := Call(context.Background(), teachersPhotosRequest, &teachersPhotosResponse); err != nil {
			panic(err)
		}

		schoolReport.Teachers = involvedTeachers
		schoolReport.TeacherPhotos = teachersPhotosResponse.Result.TeacherPhotos

		schoolReports = append(schoolReports, *schoolReport)

		if len(schoolReport.Behaviour) == 0 && len(schoolReport.Achievement) == 0 {
			log.Printf("There are no new achievements or behaviours for %s to report on.\n", child.Forename)
		}
	}

	return &schoolReports
}

func (r *Reporter) Generate(schoolReport *SchoolReport) string {
	r.prepareTemplates(schoolReport)

	style, err := os.ReadFile("templates/style.css")
	if err != nil {
		panic(err)
	}

	schoolReportViewData := &SchoolReportViewData{
		SchoolReport: *schoolReport,
		Style:        template.CSS(style),
	}

	var tmpl bytes.Buffer
	if err := r.template.Execute(&tmpl, schoolReportViewData); err != nil {
		panic(err)
	}

	template.New("schoolReport")
	return tmpl.String()
}
