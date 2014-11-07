package km

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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

type StateGetter func(dbmap *gorp.DbMap, dateStr string) (err error, state State)
type SaveInterface func(dbmap *gorp.DbMap, date time.Time, fields []Field) (err error)
type GetTimesInterface func(dbmap *gorp.DbMap, year, month int64) (rows []TimeRow, err error)

type Server struct {
	mux.Router
	Dbmap     *gorp.DbMap
	templates *template.Template
	config    Config
	StateFunc StateGetter
	SaveKilos SaveInterface
	SaveTimes SaveInterface
	GetTimes  GetTimesInterface
}

func NewServer(dbName string, config Config) (s *Server, err error) {
	var logFile *os.File
	// Set up logging
	if config.Log != "" {
		logFile, err = os.OpenFile(config.Log, syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT, 0666)
		if err != nil {
			return nil, fmt.Errorf("could not open logfile: %s", err.Error())
		}
		log.SetOutput(logFile)
		log.SetPrefix("km-app:\t")
	}

	db, creating_db_error := sql.Open("postgres", "user=docker dbname="+dbName+" password=docker sslmode=disable")
	testDbRegex := regexp.MustCompile("_test$")
	err = db.Ping()
	if !testDbRegex.MatchString(dbName) && err != nil {
		return nil, fmt.Errorf("ping result: %s\nsql.Open result: %s", err, creating_db_error)
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
	s = &Server{Dbmap: Dbmap,
		templates: templates,
		config:    config,
		StateFunc: GetState,
		SaveKilos: SaveKilometers,
		SaveTimes: SaveTimes,
		GetTimes:  GetAllTimes,
	}

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
	s.HandleFunc("/overview/{category}/{year}/{month}", s.overviewHandler).Methods("GET")
	s.HandleFunc("/delete/{date}", s.deleteHandler).Methods("GET")
	return s, nil
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.Env == "testing" {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, s.config)
	} else {
		s.templates.Execute(w, s.config)
	}
}

type State struct {
	Fields       []Field
	LastDayError string
	LastDayKm    int
}

func ParseJsonBody(bodyReader io.Reader) (err error, fields []Field) {
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return CustomResponse(NotParsable, err), []Field{}
	}
	err = json.Unmarshal(body, &fields)
	if err != nil {
		return CustomResponse(NotParsable, err), []Field{}
	}
	return
}

func (s *Server) saveHandler(w http.ResponseWriter, r *http.Request) {
	// parse date
	vars := mux.Vars(r)
	err, date := ParseUrlDate(vars["date"])
	if err != nil {
		myError := err.(Response)
		http.Error(w, myError.String(), myError.Code)
		return
	}

	// parse posted data
	err, fields := ParseJsonBody(r.Body)
	if err != nil {
		response := err.(Response)
		http.Error(w, response.Error(), response.Code)
		return
	}

	/// Save kilometers
	err = s.SaveKilos(s.Dbmap, date, fields)
	if err != nil {
		response := err.(Response)
		http.Error(w, response.Error(), response.Code)
		return
	}
	// save Times
	err = s.SaveTimes(s.Dbmap, date, fields)
	if err != nil {
		response := err.(Response)
		http.Error(w, response.Error(), response.Code)
		return
	}
	w.Write([]byte("ok\n"))
	// sla eerste stand van vandaag op als laatste stand van gister (als die vergeten is)
}

func (s *Server) stateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err, date := ParseUrlDate(vars["date"])
	if err != nil {
		myError := err.(Response)
		http.Error(w, myError.String(), myError.Code)
		return
	}
	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	err, state := s.StateFunc(s.Dbmap, dateStr)
	if err != nil {
		response := err.(Response)
		http.Error(w, response.Error(), response.Code)
	}
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(state)
}

func GetState(dbmap *gorp.DbMap, dateStr string) (err error, state State) {
	state.Fields = make([]Field, 4)
	// Get data save for this date
	var today Kilometers
	err = dbmap.SelectOne(&today, "select * from kilometers where date=$1", dateStr)
	switch {
	case err != nil && err.Error() != "sql: no rows in result set":
		return CustomResponse(DbError, err), State{}
	case err != nil && err.Error() == "sql: no rows in result set": // today not saved yet
		var lastDay Kilometers
		err := dbmap.SelectOne(&lastDay, "select * from kilometers where date = (select max(date) as date from kilometers)")
		if err != nil {
			return CustomResponse(DbError, err), State{}
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
		err = dbmap.SelectOne(&lastDayTimes, "select * from times where date=(select max(date) as date from times)")
		log.Println("na select laatste tijden:", err, lastDayTimes)
		if lastDayTimes.CheckIn == 0 || lastDayTimes.CheckOut == 0 {
			state.LastDayError = fmt.Sprintf("input/%02d%02d%04d", lastDayTimes.Date.Day(), lastDayTimes.Date.Month(), lastDayTimes.Date.Year())
		}

	default: // Something is already filled in for today
		log.Println("today:", today)
		var times Times
		err = dbmap.SelectOne(&times, "select * from times where date=$1", dateStr)
		if err != nil {
			return CustomResponse(DbError, err), State{}
		}
		loc, _ := time.LoadLocation("Europe/Amsterdam") // should not be hardcoded but idgaf
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
		_, err = dbmap.Select(&lastDayTimes, "select * from times order by date desc limit 2")
		if err != nil {
			return CustomResponse(DbError, err), State{}
		}
		if len(lastDayTimes) > 1 {
			log.Println("tijden van gisteren, (vandaag al half ingevuld):", err, lastDayTimes[1])
			if lastDayTimes[1].CheckIn == 0 || lastDayTimes[1].CheckOut == 0 {
				state.LastDayError = fmt.Sprintf("input/%02d%02d%04d", lastDayTimes[1].Date.Day(), lastDayTimes[1].Date.Month(), lastDayTimes[1].Date.Year())
			}

		}
	}
	return nil, state
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
			http.Error(w, fmt.Sprintf("%s\n%s", DbError.String(), err), DbError.Code)
			log.Println("overview:", err)
			return
		}
		jsonEncoder.Encode(all)
	case "tijden":
		rows, err := s.GetTimes(s.Dbmap, year, month)
		if err != nil {
			http.Error(w, DbError.String(), DbError.Code)
			log.Println("overview tijden getalltimes return:", err)
			return
		}
		jsonEncoder.Encode(rows)
	default:
		http.Error(w, InvalidUrl.String(), InvalidUrl.Code)
		return
	}
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err, date := ParseUrlDate(vars["date"])
	if err != nil {
		myError := err.(Response)
		http.Error(w, myError.String(), myError.Code)
		return
	}

	dateStr := fmt.Sprintf("%d-%d-%d", date.Month(), date.Day(), date.Year())
	err = DeleteAllForDate(s.Dbmap, dateStr)
	if err != nil {
		myError := err.(Response)
		http.Error(w, myError.String(), myError.Code)
	}
}

func DeleteAllForDate(dbmap *gorp.DbMap, dateStr string) (err error) {
	_, err = dbmap.Exec("delete from kilometers where date=$1", dateStr)
	if err != nil {
		return CustomResponse(DbError, err)
	}
	_, err = dbmap.Exec("delete from times where date=$1", dateStr)
	if err != nil {
		return CustomResponse(DbError, err)
	}
	return nil
}
