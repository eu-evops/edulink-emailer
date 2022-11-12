package edulink

type LearnerPhotosRequestParams struct {
	LearnerIDs []string `json:"learner_ids"`
	Size       int      `json:"size"`
}
type LearnerPhotosRequest struct {
	RequestBase
	Params LearnerPhotosRequestParams `json:"params"`
}

type LearnerPhoto struct {
	ID    string `json:"id"`
	Cache string `json:"cache"`
	Photo string `json:"photo"`
}

type LearnerPhotosResponse struct {
	ResponseBase
	Result struct {
		ResultBase
		LearnerPhotos []LearnerPhoto `json:"learner_photos"`
	} `json:"result"`
}

func (r LearnerPhotosRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}

func (r LearnerPhotosResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}
func (r LearnerPhotosResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}
