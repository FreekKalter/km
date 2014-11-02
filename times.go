package km

import (
	"fmt"
	"log"
	"time"

	"github.com/coopernurse/gorp"
)

/// Save times
type Times struct {
	Id                                int64
	Date                              time.Time
	Begin, CheckIn, CheckOut, Laatste int64
}

func (t *Times) UpdateObject(date string, fields []Field) error {
	loc, _ := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
	for _, field := range fields {
		if field.Time == "" {
			continue
		}
		fieldLocalTime, err := time.ParseInLocation("1-2-2006 15:04", fmt.Sprintf("%s %s", date, field.Time), loc)
		if err != nil {
			return err
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

func SaveTimes(dbmap *gorp.DbMap, date time.Time, fields []Field) (err error) {
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	times := new(Times)
	err = dbmap.SelectOne(times, "select * from times where date=$1", dateStr)
	if err != nil && err.Error() == "sql: no rows in result set" {
		times := new(Times)
		times.Date = date
		times.UpdateObject(dateStr, fields)
		times.Id = -1
		log.Printf("object to be insterted: %+v\n", times)
		err = dbmap.Insert(times)
	} else if err == nil {
		log.Printf("times object to update VOOR invoegen van de op te slaan velden: %+v\n", times)
		err = times.UpdateObject(dateStr, fields)
		if err != nil {
			return
		}
		log.Printf("times object to update NA invoegen van de op te slaan velden: %+v\n", times)
		var count int64
		count, err = dbmap.Update(times)
		if err != nil {
			return
		}
		if count != 1 {
			return fmt.Errorf("Update did not return a count of 1, instead: %d", count)
		}
	}
	if err != nil {
		return
	}
	return nil
}
