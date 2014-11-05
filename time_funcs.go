package km

import "time"

func ParseUrlDate(dateStr string) (err error, date time.Time) {
	date, err = time.Parse("02012006", dateStr)
	if err != nil {
		return CustomResponse(InvalidDate, err), time.Time{}
	}
	return nil, date
}
