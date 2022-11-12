package edulink

type EmployeesRequestParams struct {
	LearnerIDs []string `json:"learner_ids"`
}
type EmployeesRequest struct {
	RequestBase
	Params EmployeesRequestParams `json:"params"`
}

type Employee struct {
	Email       string `json:"email"`
	Forename    string `json:"forename"`
	Gender      string `json:"gender"`
	ID          string `json:"id"`
	MobilePhone string `json:"mobile_phone"`
	Phone       string `json:"phone"`
	Surname     string `json:"surname"`
	Title       string `json:"title"`
}

type EmployeesResponse struct {
	ResponseBase

	Result struct {
		ResultBase
		Employees []Employee `json:"employees"`
	} `json:"result"`
}

func (r EmployeesResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r EmployeesResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}

func (r EmployeesRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}
