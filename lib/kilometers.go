package km

import (
	"fmt"
	"time"

	"github.com/coopernurse/gorp"
	// import postgres for its side effects
	_ "github.com/lib/pq"
)

// Kilometers is the struct representing a db row in the kilometers table
type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

// Field holds the data for 1 row in the ui form
type Field struct {
	Km   int
	Time string
	Name string
}

func (k *Kilometers) getMax() int {
	if k.Terug > 0 {
		return k.Terug
	}
	if k.Laatste > 0 {
		return k.Laatste
	}
	if k.Eerste > 0 {
		return k.Eerste
	}
	if k.Begin > 0 {
		return k.Begin
	}
	return 0
}

// AddFields updates an existing Kilometers truct to update/insert into db
// with data posted by user
func (k *Kilometers) AddFields(fields []Field) {
	for _, field := range fields {
		switch field.Name {
		case "Begin":
			k.Begin = field.Km
		case "Eerste":
			k.Eerste = field.Km
		case "Laatste":
			k.Laatste = field.Km
		case "Terug":
			k.Terug = field.Km
		}

	}
}

// SaveKilometers saves a the given Field array (wich is supplied by the user)
// if no data is saved for today it results in an insert, otherwise a update of
// the already saved data is done
func SaveKilometers(dbmap *gorp.DbMap, date time.Time, fields []Field) (err error) {
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	kms := new(Kilometers)
	err = dbmap.SelectOne(kms, "select * from kilometers where date=$1", dateStr)
	if err == nil { // there is already data for km (so use update)
		kms.AddFields(fields)
		_, err = dbmap.Update(kms)
		if err != nil {
			return CustomResponse(DbError, err)
		}
	} else { // nog niks opgeslagen voor vandaag}
		if err.Error() != "sql: no rows in result set" {
			return CustomResponse(DbError, err)
		}
		kms := new(Kilometers)
		kms.Date = date
		kms.AddFields(fields)
		err = dbmap.Insert(kms)
		if err != nil {
			return CustomResponse(DbError, err)
		}
	}
	return nil
}
