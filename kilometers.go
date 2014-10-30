package km

import (
	"fmt"
	"time"

	"github.com/coopernurse/gorp"
)

type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

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

func SaveKilometers(dbmap *gorp.DbMap, date time.Time, fields []Field) (err error) {
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	kms := new(Kilometers)
	err = dbmap.SelectOne(kms, "select * from kilometers where date=$1", dateStr)
	if err == nil { // there is already data for km (so use update)
		kms.AddFields(fields)
		_, err = dbmap.Update(kms)
		if err != nil {
			return
		}
	} else { // nog niks opgeslagen voor vandaag}
		kms := new(Kilometers)
		kms.Date = date
		kms.AddFields(fields)
		err = dbmap.Insert(kms)
		if err != nil {
			return
		}
	}
	return nil
}
