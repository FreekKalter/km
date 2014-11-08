package km

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestUpdateKilometers(t *testing.T) {
	// successfull update
	err, dbmap, columns := MockSetup("kilometers")
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

	// update fail test
	err, dbmap, columns = MockSetup("kilometers")
	if err != nil {
		t.Error(err)
	}
	sqlmock.ExpectQuery("select \\* from kilometers where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date, 1234, 0, 0, 0, ""))
	sqlmock.ExpectExec("update \"kilometers\" set \"date\"=(.+)").
		WithArgs(date, 1234, 0, 0, 12345, "", 1).
		WillReturnError(fmt.Errorf("failed update"))
	fields = []Field{Field{Name: "Terug", Km: 12345, Time: "13:00"}}
	err = SaveKilometers(dbmap, date, fields)
	if err == nil {
		t.Errorf("Updating kilometers passed without error, when it should have returned one")
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}

	// select fails
	err, dbmap, columns = MockSetup("kilometers")
	if err != nil {
		t.Error(err)
	}
	sqlmock.ExpectQuery("select \\* from kilometers where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnError(fmt.Errorf("failed select"))
	fields = []Field{Field{Name: "Terug", Km: 12345, Time: "13:00"}}
	err = SaveKilometers(dbmap, date, fields)
	if err == nil {
		t.Errorf("Updating kilometers passed without error, when it should have returned one")
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}
}

func TestInsertKilometers(t *testing.T) {
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	err, dbmap, columns := MockSetup("kilometers")
	if err != nil {
		t.Error(err)
	}
	date := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	sqlmock.ExpectQuery("select \\* from kilometers where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))

	// INSERT is Query aparently, not Exec as my long struggle to get this working discovered
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

	// test failure on insert
	err, dbmap, columns = MockSetup("kilometers")
	if err != nil {
		t.Error(err)
	}
	sqlmock.ExpectQuery("select \\* from kilometers where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))
	// INSERT is Query aparently, not Exec as my long struggle to get this working discovered
	sqlmock.ExpectQuery("insert into \"kilometers\"(.+)").
		WithArgs(date, 0, 0, 0, 12345, ""). //autoincrement field (id in this case) not given to WithArgs
		WillReturnError(fmt.Errorf("failed instert"))
	fields = []Field{Field{Name: "Terug", Km: 12345, Time: "13:00"}}
	err = SaveKilometers(dbmap, date, fields)
	if err == nil {
		t.Errorf("Inserting kilometers passed without error, when it should have returned one")
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}
}
