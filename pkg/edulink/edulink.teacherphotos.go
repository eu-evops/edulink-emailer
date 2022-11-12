package edulink

type TeacherPhotosRequestParams struct {
	EmployeeIDs []string `json:"employee_ids"`
	Size        int      `json:"size"`
}
type TeacherPhotosRequest struct {
	RequestBase
	Params TeacherPhotosRequestParams `json:"params"`
}

type TeacherPhoto struct {
	ID    string `json:"id"`
	Cache string `json:"cache"`
	Photo string `json:"photo"`
}

type TeacherPhotosResponse struct {
	ResponseBase
	Result struct {
		ResultBase
		TeacherPhotos []TeacherPhoto `json:"employee_photos"`
	} `json:"result"`
}

func (r TeacherPhotosResponse) GetByID(id string) (*TeacherPhoto, error) {
	for _, v := range r.Result.TeacherPhotos {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, &ErrNotFound{}
}

func (r TeacherPhotosRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}

func (r TeacherPhotosResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r TeacherPhotosResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}
