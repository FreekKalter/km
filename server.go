package km

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"text/template"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Config struct {
	Env string
	Log string
}

type Server struct {
	mux.Router
	Dbmap     *gorp.DbMap
	templates *template.Template
	config    Config
}

func NewServer(dbName string, config Config) *Server {
	var (
		logFile *os.File
		err     error
	)
	// Set up logging
	if config.Log != "" {
		logFile, err = os.OpenFile(config.Log, syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT, 0666)
		if err != nil {
			log.Panic(err)
		}
		log.SetOutput(logFile)
		log.SetPrefix("km-app:\t")
	}

	db, err := sql.Open("postgres", "user=docker dbname="+dbName+" password=docker sslmode=disable")
	if err != nil {
		fmt.Println("init:", err)
		log.Fatal("init:", err)
	}
	var Dbmap *gorp.DbMap
	Dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	Dbmap.AddTable(Kilometers{}).SetKeys(true, "Id")
	Dbmap.AddTable(Times{}).SetKeys(true, "Id")

	var templates *template.Template
	if config.Env == "testing" {
		Dbmap.TraceOn("[gorp]", log.New(logFile, "DB:\t", log.LstdFlags))
	} else {
		templates = template.Must(template.ParseFiles("index.html"))
	}
	s := &Server{Dbmap: Dbmap, templates: templates, config: config}

	// static files get served directly
	if config.Env == "testing" {
		s.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
		s.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
		s.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))
		s.PathPrefix("/partials/").Handler(http.StripPrefix("/partials/", http.FileServer(http.Dir("partials/"))))
		s.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "favicon.ico") }))
	}

	s.HandleFunc("/", s.homeHandler).Methods("GET")
	s.HandleFunc("/state/{date}", s.stateHandler).Methods("GET")
	s.HandleFunc("/save/{date}", s.saveHandler).Methods("POST")
	//s.HandleFunc("/save/kilometers/{id}", s.saveKilometersHandler).Methods("POST")
	//s.HandleFunc("/save/times/{id}", s.saveTimesHandler).Methods("POST")
	s.HandleFunc("/overview/{category}/{year}/{month}", s.overviewHandler).Methods("GET")
	s.HandleFunc("/delete/{date}", s.deleteHandler).Methods("GET")
	//s.HandleFunc("/csv/{year}/{month}", s.csvHandler).Methods("GET")
	return s
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.Env == "testing" {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, s.config)
	} else {
		s.templates.Execute(w, s.config)
	}
}

func (s *Server) saveHandler(w http.ResponseWriter, r *http.Request) {
	// parse date
	vars := mux.Vars(r)
	date, err := time.Parse("02012006", vars["date"])
	if err != nil {
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	log.Println(dateStr)

	// parse posted data
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, NotParsable.String(), NotParsable.Code)
		log.Println(err)
		return
	}
	var fields []Field
	err = json.Unmarshal(body, &fields)
	if err != nil {
		http.Error(w, NotParsable.String(), NotParsable.Code)
		log.Println(err)
		return
	}
	log.Printf("parsed array of fields to save: %+v\n", fields)
	//TODO: sanitize input

	/// Save kilometers
	SaveKilometers(s.Dbmap, date, fields)
	// save Times
	times := new(Times)
	err = s.Dbmap.SelectOne(times, "select * from times where date=$1", dateStr)
	log.Println("err na selectONe from times:", err)
	if err != nil && err.Error() == "sql: no rows in result set" {
		times := new(Times)
		times.Date = date
		times.UpdateObject(dateStr, fields)
		times.Id = -1
		log.Printf("object to be insterted: %+v\n", times)
		err = s.Dbmap.Insert(times)
	} else if err == nil {
		log.Printf("times object to update VOOR invoegen van de op te slaan velden: %+v\n", times)
		err = times.UpdateObject(dateStr, fields)
		if err != nil {
			log.Println(err)
		}
		log.Printf("times object to update NA invoegen van de op te slaan velden: %+v\n", times)
		_, err = s.Dbmap.Update(times)
	}
	if err != nil {
		http.Error(w, DbError.String(), DbError.Code)
		log.Println(err)
		return
	}
	log.Printf("%+v\n", times)
	// sla eerste stand van vandaag op als laatste stand van gister (als die vergeten is)
}
func (s *Server) stateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date, err := time.Parse("02012006", vars["date"])
	log.Println(date)
	if err != nil {
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	type State struct {
		Fields       []Field
		LastDayError string
		LastDayKm    int
	}
	var state State
	state.Fields = make([]Field, 4)

	// Get data save for this date
	var today Kilometers
	err = s.Dbmap.SelectOne(&today, "select * from kilometers where date=$1", dateStr)
	log.Println(err)
	switch {
	case err != nil && err.Error() != "sql: no rows in result set":
		http.Error(w, DbError.String(), DbError.Code)
		log.Println("stateHandler:", err)
		return
	case err != nil && err.Error() == "sql: no rows in result set": // today not saved yet
		log.Println("no today")
		var lastDay Kilometers
		err := s.Dbmap.SelectOne(&lastDay, "select * from kilometers where date = (select max(date) as date from kilometers)")
		if err != nil {
			http.Error(w, "Database error", 500)
			log.Println("stateHandler:", err)
			return
		}
		if lastDay != (Kilometers{}) { // Nothing in db yet
			log.Println("nothing in db yet for todag:", dateStr)
			state.LastDayKm = lastDay.getMax()
			state.Fields[0] = Field{Name: "Begin"}
			state.Fields[1] = Field{Name: "Eerste"}
			state.Fields[2] = Field{Name: "Laatste"}
			state.Fields[3] = Field{Name: "Terug"}
		}
		var lastDayTimes Times
		err = s.Dbmap.SelectOne(&lastDayTimes, "select * from times where date=(select max(date) as date from times)")
		log.Println("na select laatste tijden:", err, lastDayTimes)
		if lastDayTimes.CheckIn == 0 || lastDayTimes.CheckOut == 0 {
			state.LastDayError = fmt.Sprintf("input/%02d%02d%04d", lastDayTimes.Date.Day(), lastDayTimes.Date.Month(), lastDayTimes.Date.Year())
		}

	default: // Something is already filled in for today
		log.Println("today:", today)
		var times Times
		err = s.Dbmap.SelectOne(&times, "select * from times where date=$1", dateStr)
		loc, err := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
		if err != nil {
			log.Println(err)
		}
		convertTime := func(t int64) string {
			ret := ""
			if t != 0 {
				ret = time.Unix(t, 0).In(loc).Format("15:04")
			}
			return ret
		}
		state.Fields[0] = Field{Km: today.Begin, Name: "Begin", Time: convertTime(times.Begin)}
		state.Fields[1] = Field{Km: today.Eerste, Name: "Eerste", Time: convertTime(times.CheckIn)}
		state.Fields[2] = Field{Km: today.Laatste, Name: "Laatste", Time: convertTime(times.CheckOut)}
		state.Fields[3] = Field{Km: today.Terug, Name: "Terug", Time: convertTime(times.Laatste)}
		log.Printf("state: %+v", state)

		var lastDayTimes []Times
		_, err = s.Dbmap.Select(&lastDayTimes, "select * from times order by date desc limit 2")
		if err != nil {
			log.Println("probeer laatste twee tijden op te halen:", err)
		}
		if len(lastDayTimes) > 1 {
			log.Println("tijden van gisteren, (vandaag al half ingevuld):", err, lastDayTimes[1])
			if lastDayTimes[1].CheckIn == 0 || lastDayTimes[1].CheckOut == 0 {
				state.LastDayError = fmt.Sprintf("input/%02d%02d%04d", lastDayTimes[1].Date.Day(), lastDayTimes[1].Date.Month(), lastDayTimes[1].Date.Year())
			}

		}
	}
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(state)
}

func (s *Server) overviewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]
	year, err := strconv.ParseInt(vars["year"], 10, 64)
	if err != nil {
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		log.Println("overview:", err)
		return
	}
	month, err := strconv.ParseInt(vars["month"], 10, 64)
	if err != nil {
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		log.Println("overview:", err)
		return
	}
	log.Println("overview", year, month)

	jsonEncoder := json.NewEncoder(w)
	switch category {
	case "kilometers":
		all := make([]Kilometers, 0)
		_, err := s.Dbmap.Select(&all, "select * from kilometers where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
		if err != nil {
			http.Error(w, DbError.String(), DbError.Code)
			log.Println("overview:", err)
			return
		}
		jsonEncoder.Encode(all)
	case "tijden":
		rows, err := getAllTimes(s, year, month)
		if err != nil {
			http.Error(w, DbError.String(), DbError.Code)
			log.Println("overview tijden getalltimes return:", err)
		}
		jsonEncoder.Encode(rows)
	default:
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		return
	}
}

func getAllTimes(s *Server, year, month int64) (rows []TimeRow, err error) {
	var all []Times
	rows = make([]TimeRow, 0)
	_, err = s.Dbmap.Select(&all, "select * from times where extract (year from date)=$1 and extract (month from date)=$2 order by date desc ", year, month)
	if err != nil {
		return rows, err
	}
	loc, err := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
	if err != nil {
		log.Println(err)
	}
	for _, c := range all {
		row := NewTimeRow()
		row.Id = c.Id
		row.Date = c.Date
		if c.Begin != 0 {
			row.Begin = time.Unix(c.Begin, 0).In(loc).Format("15:04")
		}
		if c.CheckIn != 0 {
			row.CheckIn = time.Unix(c.CheckIn, 0).In(loc).Format("15:04")
		}
		if c.CheckOut != 0 {
			row.CheckOut = time.Unix(c.CheckOut, 0).In(loc).Format("15:04")
		}
		if c.Laatste != 0 {
			row.Laatste = time.Unix(c.Laatste, 0).In(loc).Format("15:04")

		}
		if hours := (time.Duration(c.CheckOut-c.CheckIn) * time.Second).Hours(); hours > 0 && hours < 24 {
			row.Hours = hours
		}
		rows = append(rows, row)
	}
	return rows, nil
}

//func (s *Server) csvHandler(w http.ResponseWriter, r *http.Request) {
//vars := mux.Vars(r)
//month, err := strconv.ParseInt(vars["month"], 10, 64)
//if err != nil {
//http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
//log.Println("overview:", err)
//return
//}
//year, err := strconv.ParseInt(vars["year"], 10, 64)
//if err != nil {
//http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
//log.Println("overview:", err)
//return
//}
//times, err := getAllTimes(s, year, month)
//if err != nil {
//http.Error(w, DbError.String(), DbError.Code)
//log.Println("overview:", err)
//}
//for _, t := range times {
//fmt.Fprintf(w, "%s,%s,%s,%.1f\n", t.Date.Format("Mon 2"), t.CheckIn, t.CheckOut, t.Hours)
//}
//}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date, err := time.Parse("02012006", vars["date"])
	log.Println(date)
	if err != nil {
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())

	_, err = s.Dbmap.Exec("delete from kilometers where date=$1", dateStr)
	if err != nil {
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}

	_, err = s.Dbmap.Exec("delete from times where date=$1", dateStr)
	if err != nil {
		log.Println("error deleting from times", err)
		http.Error(w, InvalidId.String(), InvalidId.Code)
		return
	}
	log.Println("delete:", err)
}

func getDateStr() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d-%d-%d", now.Month(), now.Day(), now.Year())
}

type TimeRow struct {
	Id                                int64
	Date                              time.Time
	Begin, CheckIn, CheckOut, Laatste string
	Hours                             float64
}

func NewTimeRow() TimeRow {
	var t TimeRow
	t.Begin = "-"
	t.CheckIn = "-"
	t.CheckOut = "-"
	t.Laatste = "-"
	return t
}
