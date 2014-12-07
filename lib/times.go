package km

import (
	"fmt"
	"log"
	"time"

	"github.com/coopernurse/gorp"
)

// Times represents a db row in the times table
type Times struct {
	Id                                int64
	Date                              time.Time
	Begin, CheckIn, CheckOut, Laatste int64
}

// TimeRow is a Times row converted to the format displayed in the frontend
type TimeRow struct {
	Id                                int64
	Date                              time.Time
	Begin, CheckIn, CheckOut, Laatste string
	Hours                             float64
}

// NewTimeRow creates new initialized TimeRow
func NewTimeRow() TimeRow {
	var t TimeRow
	t.Begin = "-"
	t.CheckIn = "-"
	t.CheckOut = "-"
	t.Laatste = "-"
	return t
}

// UpdateObject updates the times struct with posted data coming from the user
// this struct can used to update the current state in the db
func (t *Times) UpdateObject(date string, fields []Field) error {
	loc, _ := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
	for _, field := range fields {
		if field.Time == "" {
			continue
		}
		fieldLocalTime, err := time.ParseInLocation("1-2-2006 15:04", fmt.Sprintf("%s %s", date, field.Time), loc)
		if err != nil {
			return CustomResponse(NotParsable, err)
		}
		fieldTime := fieldLocalTime.UTC().Unix()
		switch field.Name {
		case "Begin":
			t.Begin = fieldTime
		case "Eerste":
			t.CheckIn = fieldTime
		case "Laatste":
			t.CheckOut = fieldTime
		case "Terug":
			t.Laatste = fieldTime
		}
	}
	return nil
}

// SaveTimes saves a given fields array to the db backend
func SaveTimes(dbmap *gorp.DbMap, date time.Time, fields []Field) (err error) {
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	times := new(Times)
	err = dbmap.SelectOne(times, "select * from times where date=$1", dateStr)
	//if err != nil && err.Error() == "sql: no rows in result set" {
	if err == nil {
		log.Printf("times object to update VOOR invoegen van de op te slaan velden: %+v\n", times)
		err = times.UpdateObject(dateStr, fields)
		if err != nil {
			return CustomResponse(DbError, err)
		}
		log.Printf("times object to update NA invoegen van de op te slaan velden: %+v\n", times)
		var count int64
		count, err = dbmap.Update(times)
		if err != nil {
			return CustomResponse(DbError, err)
		}
		if count != 1 {
			return CustomResponse(DbError, fmt.Errorf("update did not return a count of 1, instead: %d", count))
		}
	} else {
		if err.Error() != "sql: no rows in result set" {
			return CustomResponse(DbError, err)
		}
		times := new(Times)
		times.Date = date
		times.UpdateObject(dateStr, fields)
		times.Id = -1
		log.Printf("object to be insterted: %+v\n", times)
		err = dbmap.Insert(times)
	}
	return nil
}

// GetAllTimes pulls all rows for a given month from the db and converts it all to TimeRow for
// displaying in the frontend
func GetAllTimes(dbmap *gorp.DbMap, year, month int64) (rows []TimeRow, err error) {
	var all []Times
	rows = make([]TimeRow, 0)
	_, err = dbmap.Select(&all, "select * from times where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
	if err != nil {
		return rows, err
	}
	loc, _ := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
	for _, c := range all {
		row := NewTimeRow()
		row.Id = c.Id
		row.Date = c.Date
		if c.Begin != 0 {
			row.Begin = time.Unix(c.Begin, 0).In(loc).Format("15:04")
		}
		if c.CheckIn != 0 {
			row.CheckIn = time.Unix(c.CheckIn, 0).In(loc).Format("15:04")
		}
		if c.CheckOut != 0 {
			row.CheckOut = time.Unix(c.CheckOut, 0).In(loc).Format("15:04")
		}
		if c.Laatste != 0 {
			row.Laatste = time.Unix(c.Laatste, 0).In(loc).Format("15:04")

		}
		if hours := (time.Duration(c.CheckOut-c.CheckIn) * time.Second).Hours(); hours > 0 && hours < 24 {
			row.Hours = hours
		}
		rows = append(rows, row)
	}
	return rows, nil
}
