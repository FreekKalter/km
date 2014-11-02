package km

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

func TestGetMax(t *testing.T) {
	kiloTests := []Kilometers{
		Kilometers{1, time.Now(), 1, 0, 0, 0, "test"},
		Kilometers{1, time.Now(), 1, 2, 0, 0, "test"},
		Kilometers{1, time.Now(), 1, 2, 3, 0, "test"},
		Kilometers{1, time.Now(), 1, 2, 3, 4, "test"},
	}
	for i, k := range kiloTests {
		if v := k.getMax(); v != i+1 {
			t.Errorf("got %d: want %d", v, i+1)
		}
	}
	if v := new(Kilometers).getMax(); v != 0 {
		t.Errorf("got %d: want %d", v, 0)
	}
}

func TestAddFields(t *testing.T) {
	fields := []Field{
		Field{Km: 1, Name: "Begin"},
		Field{Km: 2, Name: "Eerste"},
		Field{Km: 3, Name: "Laatste"},
		Field{Km: 4, Name: "Terug"},
	}
	ki := new(Kilometers)
	ki.AddFields(fields)
	if ki.Begin != 1 || ki.Eerste != 2 || ki.Laatste != 3 || ki.Terug != 4 {
		t.Error("Kilometers AddFields did not do its job")
	}

}

func mockSetup() (err error, dbmap *gorp.DbMap, columns []string) {
	db, err := sqlmock.New()
	if err != nil {
		return
	}
	columns = []string{"Id", "Date", "Begin", "Eerste", "Laatste", "Terug", "Comment"}
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

func TestUpdateKilometers(t *testing.T) {
	err, dbmap, columns := mockSetup()
	if err != nil {
		t.Error(err)
	}
	date := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	sqlmock.ExpectQuery("select \\* from kilometers where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date, 1234, 0, 0, 0, ""))

	sqlmock.ExpectExec("update \"kilometers\" set \"date\"=(.+)").
		WithArgs(date, 1234, 0, 0, 12345, "", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	fields := []Field{Field{Name: "Terug", Km: 12345, Time: "13:00"}}
	err = SaveKilometers(dbmap, date, fields)
	if err != nil {
		t.Errorf("SaveKilometers returned: %s", err)
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}
}

func TestInsertKilometers(t *testing.T) {
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	err, dbmap, columns := mockSetup()
	if err != nil {
		t.Error(err)
	}
	date := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	sqlmock.ExpectQuery("select \\* from kilometers where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))

	// INSERT is Query aparently,
	sqlmock.ExpectQuery("insert into \"kilometers\"(.+)").
		WithArgs(date, 0, 0, 0, 12345, ""). //autoincrement field (id in this case) not given to WithArgs
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	fields := []Field{Field{Name: "Terug", Km: 12345, Time: "13:00"}}
	err = SaveKilometers(dbmap, date, fields)
	if err != nil {
		t.Errorf("SaveKilometers returned: %s", err)
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}
}
