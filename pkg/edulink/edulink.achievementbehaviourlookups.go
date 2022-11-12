package edulink

type AchievementBehaviourLookupsRequest struct {
	RequestBase
	Params struct{} `json:"params"`
}

type AchievementBehaviourLookups struct {
	ID                string   `json:"id"`
	Date              DateOnly `json:"date"`
	Comments          string   `json:"comments"`
	LessonInformation string   `json:"lesson_information"`
	Points            int      `json:"points"`
	Recorded          struct {
		Date       DateOnly `json:"date"`
		EmployeeID string   `json:"employee_id"`
	}
}

type AchievementBehaviourLookupsResponse struct {
	ResponseBase
	Result struct {
		ResultBase
		DetentionManagementEnabled bool `json:"detentionmanagement_enabled"`
		BehaviourTypes             []struct {
			Active            bool   `json:"active"`
			ID                string `json:"id"`
			Code              string `json:"code"`
			Description       string `json:"description"`
			IncludeInRegister bool   `json:"include_in_register"`
			IsBullyingType    bool   `json:"is_bullying_type"`
			Points            int    `json:"points"`
			Position          int    `json:"position"`
			System            bool   `json:"system"`
		} `json:"behaviour_types"`

		BehaviourTimes []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"behaviour_times"`

		BehaviourStatuses []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"behaviour_statuses"`

		BehaviourRequireFields  []string `json:"behaviour_require_fields"`
		BehaviourPointsEditable bool     `json:"behaviour_points_editable"`
		BehaviourLocations      []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"behaviour_locations"`

		BehaviourHiddenFieldsOnEntry []string `json:"behaviour_hidden_fields_on_entry"`

		BehaviourBullyingTypes []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"behaviour_bullying_types"`

		BehaviourActivityTypes []struct {
			ID          string `json:"id"`
			Code        string `json:"code"`
			Active      bool   `json:"active"`
			Description string `json:"description"`
		} `json:"behaviour_activity_types"`

		BehaviourActionsTaken []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"behaviour_actions_taken"`

		AchievementTypes []struct {
			Active      bool   `json:"active"`
			Code        string `json:"code"`
			Description string `json:"description"`
			ID          string `json:"id"`
			Points      int    `json:"points"`
			Position    int    `json:"position"`
			System      bool   `json:"system"`
		} `json:"achievement_types"`
	} `json:"result"`
	AchivementRequireFields       []string `json:"achivement_require_fields"`
	AchievementPointsEditable     bool     `json:"achievement_points_editable"`
	AchivementHiddenFieldsOnEntry []string `json:"achivement_hidden_fields_on_entry"`
	AchievementAwardTypes         []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"achievement_award_types"`

	AchievementActivityTypes []struct {
		Active      bool   `json:"active"`
		ID          string `json:"id"`
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"achievement_activity_types"`
}

func (r AchievementBehaviourLookupsRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}

func (r AchievementBehaviourLookupsResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r AchievementBehaviourLookupsResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}
