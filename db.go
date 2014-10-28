package km

import (
	"time"

	"github.com/coopernurse/gorp"
)

const timeFmt = "01-02-2006"

type Db struct {
	today *Kilometers
	dbMap *gorp.DbMap
}

func (db *Db) getToday() *Kilometers {
	if db.today.Date.Format(timeFmt) == time.Now().Format(timeFmt) {
		return db.today
	}
	var today Kilometers = Kilometers{}
	err := db.dbMap.SelectOne(&today, "select * from kilometers where date=$1", time.Now().UTC().Format(timeFmt))
	if err != nil {
		panic(err)
	}
	return &today
}
