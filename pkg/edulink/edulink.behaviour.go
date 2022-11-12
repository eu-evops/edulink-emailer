package edulink

type BehaviourRequestParams struct {
	LearnerID string `json:"learner_id"`
	Format    int    `json:"format"`
}
type BehaviourRequest struct {
	RequestBase
	Params BehaviourRequestParams `json:"params"`
}

type Behaviour struct {
	ID                  string   `json:"id"`
	ActivityID          string   `json:"activity_id"`
	BullyingTypeID      string   `json:"bullying_type_id"`
	Comments            string   `json:"comments"`
	Date                DateOnly `json:"date"`
	InvolvedEmployeeIDs []string `json:"involved_employee_ids"`
	LocationID          string   `json:"location_id"`
	LessonInformation   string   `json:"lesson_information"`
	Points              int      `json:"points"`
	Source              string   `json:"source"`
	StatusID            string   `json:"status_id"`
	TimeID              string   `json:"time_id"`
	TypeIDs             []string `json:"type_ids"`
	Recorded            struct {
		Date       DateOnly `json:"date"`
		EmployeeID string   `json:"employee_id"`
	}
}

type BehaviourResponse struct {
	ResponseBase
	Result struct {
		ResultBase
		Behaviour  []Behaviour `json:"behaviour"`
		Detentions []struct {
			ID                  string   `json:"id"`
			Attended            string   `json:"attended"`
			Date                DateOnly `json:"date"`
			Description         string   `json:"description"`
			StartTime           string   `json:"start_time"`
			EndTime             string   `json:"end_time"`
			NonAttendanceReason string   `json:"non_attendance_reason"`
			Location            string   `json:"location"`
		} `json:"detentions"`

		Employees []Employee `json:"employees"`
	} `json:"result"`
}

func (r BehaviourRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}

func (r BehaviourResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r BehaviourResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}
