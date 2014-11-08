package km

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/coopernurse/gorp"
)

func MockSetup(table string) (err error, dbmap *gorp.DbMap, columns []string) {
	db, err := sqlmock.New()
	if err != nil {
		return
	}
	if table == "kilometers" {
		columns = []string{"Id", "Date", "Begin", "Eerste", "Laatste", "Terug", "Comment"}
	} else if table == "times" {
		columns = []string{"Id", "Date", "Begin", "CheckIn", "CheckOut", "Laatste"}

	}
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")
	dbmap.AddTable(Times{}).SetKeys(true, "Id")
	if testing.Verbose() {
		dbmap.TraceOn("DB:\t", log.New(os.Stdout, "", log.Lshortfile))
	} else {

		dbmap.TraceOn("DB:\t", log.New(ioutil.Discard, "", log.Lshortfile))
	}
	return
}
