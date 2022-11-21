package edulink

import "time"

const API_ENDPOINT = "https://roundwoodpark.edulinkone.com/api/"
const SCHOOL_ID = 2

type Request interface {
	GetBaseRequest() RequestBase
}

type Response interface {
	GetBaseResponse() ResponseBase
}

type Result interface {
	GetBaseResult() ResultBase
}

type RequestBase struct {
	ID        int    `json:"id"`
	JsonRPC   string `json:"jsonrpc"`
	Method    string `json:"method"`
	UUID      string `json:"uuid"`
	AuthToken string `json:"authtoken,omitempty"`
}

type ResultBase struct {
	Method  string `json:"method"`
	Success bool   `json:"success"`

	Metrics struct {
		Be       string    `json:"be"`
		Sspt     float64   `json:"sspt"`
		SsptUs   int64     `json:"sspt_us"`
		St       time.Time `json:"st"`
		UniqueID string    `json:"unique_id"`
	} `json:"metrics"`
}

type ResponseBase struct {
	ID      int    `json:"id"`
	JsonRPC string `json:"jsonrpc"`

	Result ResultBase `json:"result"`
}

type SchoolReport struct {
	Child Child `json:"child"`

	Photo string `json:"photo"`

	Behaviour     []Behaviour    `json:"behaviour"`
	Achievement   []Achievement  `json:"achievement"`
	School        Establishment  `json:"school"`
	Teachers      []Employee     `json:"teachers"`
	TeacherPhotos []TeacherPhoto `json:"teacher_photos"`
}

type ErrNotFound struct{}

func (e *ErrNotFound) Error() string {
	return "not found"
}
