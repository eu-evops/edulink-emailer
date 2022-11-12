package edulink

import (
	"strings"
	"time"
)

type DateOnly time.Time

func (d DateOnly) Format(format string) string {
	return time.Time(d).Format(format)
}

func (d DateOnly) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`) //get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02", value) //parse time
	if err != nil {
		return err
	}
	*d = DateOnly(t) //set result using the pointer
	return nil
}
