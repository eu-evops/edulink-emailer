package edulink

type AchievementRequestParams struct {
	LearnerID string `json:"learner_id"`
	Format    int    `json:"format"`
}
type AchievementRequest struct {
	RequestBase
	Params AchievementRequestParams `json:"params"`
}

type Achievement struct {
	ID                  string   `json:"id"`
	ActivityID          string   `json:"activity_id"`
	Date                DateOnly `json:"date"`
	Comments            string   `json:"comments"`
	InvolvedEmployeeIDs []string `json:"involved_employee_ids"`
	LessonInformation   string   `json:"lesson_information"`
	Points              int      `json:"points"`
	Source              string   `json:"source"`
	TypeIDs             []string `json:"type_ids"`
	Recorded            struct {
		Date       DateOnly `json:"date"`
		EmployeeID string   `json:"employee_id"`
	}
}

type AchievementResponse struct {
	ResponseBase
	Result struct {
		ResultBase
		Achievement []Achievement `json:"achievement"`
		Employees   []Employee    `json:"employees"`
	} `json:"result"`
}

func (r AchievementRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}

func (r AchievementResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r AchievementResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}
