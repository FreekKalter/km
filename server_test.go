package km

import (
	"encoding/json"
	"fmt"

	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

var config Config = Config{Env: "production", Log: ""}
var s *Server = NewServer("km_test", config)
var db *gorp.DbMap = s.Dbmap

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

func TestDelete(t *testing.T) {
	clearTable(t, "kilometers")

	// add a row, save id
	todayStr := getTodayString()
	now := time.Now()
	k := Kilometers{Date: now, Begin: 1234}
	err := db.Insert(&k)
	if err != nil {
		t.Fatal("TestDelete: dberror on insert", err)
	}
	id := strconv.FormatInt(k.Id, 10)

	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/delete/1", InvalidId),
		NewTestCombo("/delete/a", InvalidId),
		NewTestCombo("/delete/-1", InvalidId),
		// delete saved row, compare returned id
		NewTestCombo("/delete/"+id, InvalidId),
		NewTestCombo("/delete/"+todayStr, Response{Code: 200}),
	}
	tableDrivenTest(t, table)
}
func getTodayString() string {
	today := time.Now()
	return fmt.Sprintf("%d%d%d", today.Day(), today.Month(), today.Year())
}

func TestSave(t *testing.T) {
	clearTable(t, "kilometers")

	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/save", NotFound),
		NewTestCombo("/save/a", NotFound),
	}
	todayStr := getTodayString()
	req, _ := http.NewRequest("POST", "/save/kilometers/today", strings.NewReader(`{"Name": "Begin", "Value": 1234}`))
	table = append(table, &TestCombo{req, Response{Code: 404}})

	req, _ = http.NewRequest("POST", "/save/"+todayStr, strings.NewReader(`{"Name": "Begin", "Value": "abc"}`))
	table = append(table, &TestCombo{req, NotParsable})

	req, _ = http.NewRequest("POST", "/save/"+todayStr, strings.NewReader(`{"Name": "InvalidFieldname", "Value": 1234}`))
	table = append(table, &TestCombo{req, NotParsable})
	//TODO: add more test here

	tableDrivenTest(t, table)
}

func TestHome(t *testing.T) {
	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/", Response{Code: 200}),
	}
	tableDrivenTest(t, table)
}

func TestState(t *testing.T) {
	todayStr := getTodayString()
	now := time.Now()
	k := Kilometers{Date: now, Begin: 1234}
	err := db.Insert(&k)
	if err != nil {
		t.Fatal("TestDelete: dberror on insert", err)
	}
	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/state", NotFound),
		NewTestCombo("/state/2234a", InvalidId),
		NewTestCombo("/state/today", InvalidId),
		NewTestCombo("/state/"+todayStr, Response{Code: 200}),
	}
	tableDrivenTest(t, table)

	req, _ := http.NewRequest("GET", "/state/"+todayStr, nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)
	var unMarschalled interface{}
	err = json.Unmarshal(w.Body.Bytes(), &unMarschalled)
	if err != nil {
		t.Fatal("/state/"+todayStr+" not a valid json response:", err)
	}
}

func TestOverview(t *testing.T) {
	var table []*TestCombo = []*TestCombo{
		NewTestCombo("/overview", NotFound),
		NewTestCombo("/overview/invalidCategory/2013/01", InvalidUrl),
		NewTestCombo("/overview/tijden/201a/01", InvalidUrl),
		NewTestCombo("/overview/tijden/2013/0a", InvalidUrl),
		NewTestCombo("/overview/1/2013/01", InvalidUrl),
	}
	tableDrivenTest(t, table)
}

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
