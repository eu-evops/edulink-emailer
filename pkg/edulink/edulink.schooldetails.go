package edulink

type SchoolDetailsRequestParams struct {
	EstablishmentID int `json:"establishment_id"`
}
type SchoolDetailsRequest struct {
	RequestBase
	Params SchoolDetailsRequestParams `json:"params"`
}

type SchoolDetailsResponse struct {
	ResponseBase

	Result struct {
		ResultBase
		Establishment Establishment `json:"establishment"`
	} `json:"result"`
}

func (r SchoolDetailsResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r SchoolDetailsResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}

func (r SchoolDetailsRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}
