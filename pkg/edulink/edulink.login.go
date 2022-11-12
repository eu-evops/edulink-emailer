package edulink

type LoginRequestParams struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	EstablishmentID int    `json:"establishment_id"`
}
type LoginRequest struct {
	RequestBase
	Params LoginRequestParams `json:"params"`
}

type Child struct {
	ID               string `json:"id"`
	CommunityGroupID string `json:"community_group_id"`
	FormGroupID      string `json:"form_group_id"`
	Forename         string `json:"forename"`
	Surname          string `json:"surname"`
	Gender           string `json:"gender"`
	YearGroupID      string `json:"year_group_id"`
}

type Establishment struct {
	Name string `json:"name"`
	Logo string `json:"logo"`

	Rooms []struct {
		ID   string `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"rooms"`

	YearGroups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"year_groups"`

	CommunityGroups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"community_groups"`

	DiscoverGroups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"discover_groups"`

	ApplicantAdmissionGroups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"applicant_admission_groups"`

	ApplicantIntakeGroups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"applicant_intake_groups"`

	FormGroups []struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		EmployeeID   string   `json:"employee_id"`
		RoomID       string   `json:"room_id"`
		YearGroupIDs []string `json:"year_group_ids"`
	} `json:"form_groups"`

	TeachingGroups []struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		EmployeeID   string   `json:"employee_id"`
		YearGroupIDs []string `json:"year_group_ids"`
	} `json:"teaching_groups"`

	Subjects []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Active bool   `json:"active"`
	} `json:"subjects"`

	ReportCardTargetTypes []struct {
		ID          string `json:"id"`
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"report_card_target_types"`
}

type LoginResponse struct {
	ResponseBase

	Result struct {
		ResultBase

		ApiVersion int    `json:"api_version"`
		AuthToken  string `json:"authtoken"`

		User struct {
			ID                        string   `json:"id"`
			EstablishmentID           string   `json:"establishment_id"`
			Gender                    string   `json:"gender"`
			Title                     string   `json:"title"`
			Forename                  string   `json:"forename"`
			Surname                   string   `json:"surname"`
			Types                     []string `json:"types"`
			Username                  string   `json:"username"`
			RememberPasswordPermitted bool     `json:"remember_password_permitted"`
		} `json:"user"`

		PersonalMenu []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"personal_menu"`

		LearnerMenu []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"learner_menu"`

		SubMenu struct {
			Label string `json:"label"`
		} `json:"sub_menu"`

		LoginMethod               string `json:"login_method"`
		LoginMethodChangePassword bool   `json:"login_method_change_password"`

		Children []Child `json:"children"`

		Establishment Establishment `json:"establishment"`
	} `json:"result"`
}

func (r LoginRequest) GetBaseRequest() RequestBase {
	return r.RequestBase
}

func (r LoginResponse) GetBaseResponse() ResponseBase {
	return r.ResponseBase
}

func (r LoginResponse) GetBaseResult() ResultBase {
	return r.Result.ResultBase
}
