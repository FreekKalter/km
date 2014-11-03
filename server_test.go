package km

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

var (
	config Config
	s      *Server
	db     *gorp.DbMap
)

func initServer(t *testing.T) {
	var err error
	config = Config{Env: "production", Log: "./test.log"}
	s, err = NewServer("km_test", config)
	if err != nil {
		t.Fatal(err.Error())
	}
	db = s.Dbmap
}

func TestServerInitErrors(t *testing.T) {
	_, err := NewServer("test", Config{Log: "/test.log"})
	t.Log(err)
	if err == nil {
		t.Error("NewServer should throw error on unwritable logfile input")
	}

	ss, err := NewServer("test", Config{Env: "testing", Log: "./test.log"})
	if err != nil {
		t.Error("newserver with 'testing' environment fails to init: %s", err.Error())
	}
	if ss == nil {
		t.Error("server struct returned is nil")
	}

}

func clearTable(t *testing.T, tableName string) {
	_, err := db.Exec("truncate kilometers")
	if err != nil {
		t.Fatal("truncating db: ", err)
	}
	slice, err := db.Select(Kilometers{}, "select * from kilometers")
	if len(slice) > 0 {
		t.Errorf("expected empty kilometers table to start with")
	}
	if err != nil {
		t.Errorf("could not select from kilometers")
	}
}

func tableDrivenTest(t *testing.T, table []*TestCombo) {
	for _, tc := range table {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, tc.req)
		resp := tc.resp

		if w.Code != resp.Code {
			t.Fatalf("%s : code = %d, want %d", tc.req.URL, w.Code, resp.Code)
		}
		if b := w.Body.String(); resp.Regex != nil && !resp.Regex.MatchString(b) {
			t.Fatalf("%s: body = %q, want '%s'", tc.req.URL, b, resp.String())
		}
	}
}

type TestCombo struct {
	req  *http.Request
	resp Response
}

func NewTestCombo(url string, resp Response) *TestCombo {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	return &TestCombo{req, resp}
}

//func TestDelete(t *testing.T) {
//initServer(t)
//clearTable(t, "kilometers")

//// add a row, save id
//dayWithoutPadding := "112014" // 1 januray 2014 (without padding day and month)
//goodDate := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
//now := time.Now()
//k := Kilometers{Date: now, Begin: 1234}
//err := db.Insert(&k)
//if err != nil {
//t.Fatal("TestDelete: dberror on insert", err)
//}
//id := strconv.FormatInt(k.Id, 10)

//var table []*TestCombo = []*TestCombo{
//NewTestCombo("/delete/1", InvalidId),
//NewTestCombo("/delete/a", InvalidId),
//NewTestCombo("/delete/-1", InvalidId),
//// delete saved row, compare returned id
//NewTestCombo("/delete/"+id, InvalidId),
//NewTestCombo("/delete/"+dayWithoutPadding, InvalidId),
//NewTestCombo("/delete/"+dateFormat(goodDate), Response{Code: 200}),
//}
//tableDrivenTest(t, table)
//}

//func dateFormat(t time.Time) string {
//return fmt.Sprintf("%02d%02d%d", t.Day(), t.Month(), t.Year())
//}

//func TestSaveReturnCodes(t *testing.T) {
//initServer(t)
//clearTable(t, "kilometers")

//goodDate := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
//todayStr := dateFormat(goodDate)
//var table []*TestCombo = []*TestCombo{
//NewTestCombo("/save", NotFound),
//NewTestCombo("/save/a", NotFound),         // only respond to post not get
//NewTestCombo("/save/"+todayStr, NotFound), // only respond to post not get
//}
//req, _ := http.NewRequest("POST", "/save/kilometers/today", strings.NewReader(`{"Name": "Begin", "Value": 1234}`))
//table = append(table, &TestCombo{req, Response{Code: 404}})

//req, _ = http.NewRequest("POST", "/save/"+todayStr, strings.NewReader(`{"Name": "Begin", "Km": "abc"}`))
//table = append(table, &TestCombo{req, NotParsable})

//req, _ = http.NewRequest("POST", "/save/"+todayStr, strings.NewReader(`{"Name": "InvalidFieldname", "Km": 1234}`))
//table = append(table, &TestCombo{req, NotParsable})

//req, _ = http.NewRequest("POST", "/save/"+todayStr, strings.NewReader(""))
//table = append(table, &TestCombo{req, NotParsable})

//req, _ = http.NewRequest("POST", "/save/blaat", strings.NewReader(`{"Name": "Begin", "Km": 1234}`))
//table = append(table, &TestCombo{req, InvalidId})

//tableDrivenTest(t, table)
//}

//func TestHome(t *testing.T) {
//initServer(t)
//var table []*TestCombo = []*TestCombo{
//NewTestCombo("/", Response{Code: 200}),
//}
//tableDrivenTest(t, table)
//}

//func TestState(t *testing.T) {
//initServer(t)
//goodDate := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
//dateStr := dateFormat(goodDate)
//now := time.Now()
//k := Kilometers{Date: now, Begin: 1234}
//err := db.Insert(&k)
//if err != nil {
//t.Fatal("TestDelete: dberror on insert", err)
//}
//var table []*TestCombo = []*TestCombo{
//NewTestCombo("/state", NotFound),
//NewTestCombo("/state/2234a", InvalidId),
//NewTestCombo("/state/today", InvalidId),
//NewTestCombo("/state/"+dateStr, Response{Code: 200}),
//}
//tableDrivenTest(t, table)

//req, _ := http.NewRequest("GET", "/state/"+dateStr, nil)
//w := httptest.NewRecorder()
//s.ServeHTTP(w, req)
//var unMarschalled interface{}
//err = json.Unmarshal(w.Body.Bytes(), &unMarschalled)
//if err != nil {
//t.Fatal("/state/"+dateStr+" not a valid json response:", err)
//}
//}

//func TestOverview(t *testing.T) {
//initServer(t)
//var table []*TestCombo = []*TestCombo{
//NewTestCombo("/overview", NotFound),
//NewTestCombo("/overview/invalidCategory/2013/01", InvalidUrl),
//NewTestCombo("/overview/tijden/201a/01", InvalidUrl),
//NewTestCombo("/overview/tijden/2013/0a", InvalidUrl),
//NewTestCombo("/overview/1/2013/01", InvalidUrl),
//}
//tableDrivenTest(t, table)
//}

func BenchmarkSave(b *testing.B) {
	req, _ := http.NewRequest("POST", "/save", strings.NewReader(`{"Name": "Begin", "Value": 1234}`))
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ServeHTTP(w, req)
	}
}

func BenchmarkOverview(b *testing.B) {
	//req, _ := http.NewRequest("POST", "/save", strings.NewReader(`{"Name": "Begin", "Value": 1234}`))
	req, _ := http.NewRequest("GET", "/overview/kilometers/2013/12", nil)
	w := httptest.NewRecorder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ServeHTTP(w, req)
	}
}

func TestNewTimeRow(t *testing.T) {
	tr := NewTimeRow()
	tt := TimeRow{Begin: "-", CheckIn: "-", CheckOut: "-", Laatste: "-"}
	if tr != tt {
		t.Error("NewTimeRow should return a TimeRow struct with all fields initialized to '-'")
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
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, date1, 1388577600, 0, 0, 0))
	rows, err := getAllTimes(dbmap, year, month)
	if err != nil {
		t.Errorf("getAllTimes returned: %s", err)
	}
	if len(rows) != 1 {
		t.Errorf("getAlltimes returned unexpected number of rows")
	}
	rowExpected := TimeRow{Id: 1, Date: date1, Begin: "13:00", CheckIn: "-", CheckOut: "-", Laatste: "-"}
	if rows[0] != rowExpected {
		t.Errorf("row expected: %+v, got: %+v", rowExpected, rows[0])
	}
	if err = dbmap.Db.Close(); err != nil {
		t.Errorf("Error '%s' was not expected while closing the database", err)
	}
}
