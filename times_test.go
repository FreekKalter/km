package km

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestInit(t *testing.T) {
	ti := Times{}
	if ti.Id != 0 {
		t.Fatal("id should be 0 on times init")
	}
}

func TestUpdate(t *testing.T) {
	testDate := time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)

	fields := []Field{Field{Km: 123456, Time: "13:00", Name: "Begin"}, Field{}}
	cmpDate := int64(1257854400) // 10-11-2009 13:00 uur (in unix formaat)

	dateStr := fmt.Sprintf("%d-%d-%d", testDate.Month(), testDate.Day(), testDate.Year())
	ti := Times{}
	err := ti.UpdateObject(dateStr, fields)
	if err != nil {
		t.Error(err)
	}

	if ti.Begin != cmpDate {
		t.Fatalf("updating Begin field: expected %d got %d", cmpDate, ti.Begin)
	}

	if ti.CheckIn != 0 || ti.CheckOut != 0 || ti.Laatste != 0 {
		t.Fatalf("only 1 field should change %+v", ti)
	}

	ti = Times{}
	fields = []Field{Field{Km: 123456, Time: "jemoeder", Name: "Begin"}}
	if ti.UpdateObject(dateStr, fields) == nil {
		t.Fatal("updateObjects should fail on invalid time field")
	}

	fields = []Field{Field{Km: 123456, Time: "13:00", Name: "Begin"},
		Field{Km: 123456, Time: "13:00", Name: "Eerste"},
		Field{Km: 123456, Time: "13:00", Name: "Laatste"},
		Field{Km: 123456, Time: "13:00", Name: "Terug"},
	}
	ti = Times{}
	if ti.UpdateObject(dateStr, fields) != nil {
		t.Error("Error updating times struct ", err)
	}
	if ti.Begin != cmpDate || ti.CheckIn != cmpDate || ti.CheckOut != cmpDate || ti.Laatste != cmpDate {
		t.Errorf("all fields should be %d but times struct is %+v", cmpDate, ti)
	}
}

func TestTimeInsert(t *testing.T) {
	err, dbmap, columns := MockSetup("times")
	if err != nil {
		t.Error(err)
	}
	date := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)

	sqlmock.ExpectQuery("select \\* from times where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString(""))

	// INSERT is Query aparently, not Exec as my long struggle to get this working discovered
	sqlmock.ExpectQuery("insert into \"times\"(.+)").
		WithArgs(date, 1388577600, 0, 0, 0). //autoincrement field (id in this case) not given to WithArgs
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	fields := []Field{Field{Time: "13:00", Name: "Begin"}}
	err = SaveTimes(dbmap, date, fields)
	if err != nil {
		t.Errorf("SaveTimes returned: %s", err)
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}
}

func TestTimeUpdate(t *testing.T) {
	// succesful update
	err, dbmap, columns := MockSetup("times")
	if err != nil {
		t.Error(err)
	}
	date := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	sqlmock.ExpectQuery("select \\* from times where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date, 1388577600, 0, 0, 0))
	sqlmock.ExpectExec("update \"times\" set \"date\"=(.+)").
		WithArgs(date, 1388577600, 1388577720, 0, 0, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	fields := []Field{Field{Time: "13:02", Name: "Eerste"}}
	err = SaveTimes(dbmap, date, fields)
	if err != nil {
		t.Errorf("SaveTimes returned: %s", err)
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}

	// first select * returns error
	err, dbmap, columns = MockSetup("times")
	if err != nil {
		t.Error(err)
	}
	sqlmock.ExpectQuery("select \\* from times where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnError(fmt.Errorf("failed select *"))
	err = SaveTimes(dbmap, date, fields)
	if err == nil {
		t.Errorf("SaveTimes returned: %s", err)
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}

	// update returned error
	err, dbmap, columns = MockSetup("times")
	if err != nil {
		t.Error(err)
	}
	sqlmock.ExpectQuery("select \\* from times where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date, 1388577600, 0, 0, 0))
	sqlmock.ExpectExec("update \"times\" set \"date\"=(.+)").
		WithArgs(date, 1388577600, 1388577720, 0, 0, 1).
		WillReturnError(fmt.Errorf("update failed"))
	err = SaveTimes(dbmap, date, fields)
	if err == nil {
		t.Errorf("SaveTimes returned: %s", err)
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}

	//update returns 0 affected rows
	err, dbmap, columns = MockSetup("times")
	if err != nil {
		t.Error(err)
	}
	sqlmock.ExpectQuery("select \\* from times where date=(.+)").
		WithArgs("1-1-2014").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date, 1388577600, 0, 0, 0))
	sqlmock.ExpectExec("update \"times\" set \"date\"=(.+)").
		WithArgs(date, 1388577600, 1388577720, 0, 0, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err = SaveTimes(dbmap, date, fields)
	if err == nil {
		t.Errorf("SaveTimes returned: %s", err)
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}
}

func TestGetAllTimes(t *testing.T) {
	err, dbmap, columns := MockSetup("times")
	if err != nil {
		t.Errorf("setting up mock db: %s", err.Error())
	}
	var year, month int64 = 2014, 1
	date1 := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	sqlmock.ExpectQuery("select \\* from times where (.+)").
		WithArgs(year, month).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date1, 1388577600, 1388577600, 1388578800, 1388578860))
	rows, err := GetAllTimes(dbmap, year, month)
	if err != nil {
		t.Errorf("GetAllTimes returned: %s", err)
	}
	if len(rows) != 1 {
		t.Errorf("GetAlltimes returned unexpected number of rows")
	}
	rowExpected := TimeRow{Id: 1, Date: date1, Begin: "13:00", CheckIn: "13:00", CheckOut: "13:20", Laatste: "13:21", Hours: 0.33333333333333337}
	if rows[0] != rowExpected {
		t.Errorf("row expected: %+v, got: %+v", rowExpected, rows[0])
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}

	// test fail on select * from ...
	err, dbmap, columns = MockSetup("times")
	if err != nil {
		t.Errorf("setting up mock db: %s", err)
	}

	sqlmock.ExpectQuery("select \\* from times where (.+)").
		WithArgs(year, month).
		WillReturnError(fmt.Errorf("FAIL"))

	rows, err = GetAllTimes(dbmap, year, month)
	if err == nil {
		t.Error("GetAllTimes should return error when select * from times fails")
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}

}
