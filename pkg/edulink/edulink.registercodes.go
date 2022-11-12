package edulink

type RegisterCodesRequest struct {
	RequestBase
	Params struct{} `json:"params"`
}

type RegisterCodesResponse struct {
	ResponseBase

	Result struct {
		ResultBase

		CodesProtectFromFloodFill []string `json:"codes_protect_from_flood_fill"`
		CodesToPromptChange       []string `json:"codes_to_prompt_change"`
		HideComments              bool     `json:"hide_comments"`
		LessonCodes               []struct {
			Active              bool   `json:"active"`
			Code                string `json:"code"`
			IsAuthorisedAbsence bool   `json:"is_authorised_absence"`
			IsLate              bool   `json:"is_late"`
			IsStatistical       bool   `json:"is_statistical"`
			Name                string `json:"name"`
			Present             bool   `json:"present"`
			Type                string `json:"type"`
		} `json:"lesson_codes"`

		LessonRegistersDefaultMark string `json:"lesson_registers_default_mark"`
		LessonRegistersEnabled     bool   `json:"lesson_registers_enabled"`

		StatutoryCodes []struct {
			Active              bool   `json:"active"`
			Code                string `json:"code"`
			IsAuthorisedAbsence bool   `json:"is_authorised_absence"`
			IsLate              bool   `json:"is_late"`
			IsStatistical       bool   `json:"is_statistical"`
			Name                string `json:"name"`
			Present             bool   `json:"present"`
			Type                string `json:"type"`
		} `json:"statutory_codes"`

		StatutoryRegistersDefaultMarkAm string `json:"statutory_registers_default_mark_am"`
		StatutoryRegistersDefaultMarkPm string `json:"statutory_registers_default_mark_pm"`

		StatutoryRegistersEnabled bool `json:"statutory_registers_enabled"`
		TagsAutohide              bool `json:"tags_autohide"`
	} `json:"result"`
}

func (r RegisterCodesResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r RegisterCodesResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}

func (r RegisterCodesRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}
