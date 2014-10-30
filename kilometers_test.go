package km

import (
	"log"
	"os"
	"syscall"
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

func TestSaveKilometersMock(t *testing.T) {
	//log.SetOutput(ioutil.Discard)
	db, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)

	}
	columns := []string{"Id", "Date", "Begin", "Eerste", "Laatste", "Terug", "Comment"}
	date := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	sqlmock.ExpectQuery("select \\* from kilometers where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date, 1234, 0, 0, 0, ""))

	sqlmock.ExpectExec("update \"kilometers\" set \"date\"=(.+)").
		WithArgs(date, 1234, 12345, 0, 0, "", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var Dbmap *gorp.DbMap
	Dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	tm := Dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")
	log.Printf("%+v", tm.Columns[1])

	logFile, _ := os.OpenFile("db.log", syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT, 0666)
	Dbmap.TraceOn("[gorp]", log.New(logFile, "DB:\t", log.LstdFlags))

	fields := []Field{Field{Name: "Eerste", Km: 12345, Time: "13:00"}}
	err = SaveKilometers(Dbmap, date, fields)
	if err != nil {
		t.Errorf("SaveKilometers returned: %s", err)
	}
}
