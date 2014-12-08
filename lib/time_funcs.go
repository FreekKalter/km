package km

import "time"

// ParseURLDate parse a datestring used in the url to a time.Time object
func ParseURLDate(dateStr string) (err error, date time.Time) {
	date, err = time.Parse("02012006", dateStr)
	if err != nil {
		return CustomResponse(InvalidDate, err), time.Time{}
	}
	return nil, date
}
