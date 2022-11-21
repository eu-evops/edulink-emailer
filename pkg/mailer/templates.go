package mailer

import (
	"html/template"

	"github.com/eu-evops/edulink/pkg/edulink"
)

type SchoolReportViewData struct {
	SchoolReport edulink.SchoolReport
	Style        template.CSS
}
