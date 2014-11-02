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
		t.Fatal(err)
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
