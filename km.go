package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/FreekKalter/km/lib"
	"launchpad.net/goyaml"
)

var (
	workdir, configFile *string
)

func init() {
	workdir = flag.String("workdir", fmt.Sprintf("%s/src/github.com/FreekKalter/km/", os.Getenv("GOPATH")), "directory where support files are located (js/img/index.html)")
	configFile = flag.String("config", "./config.yml", "location of configuration file")
}

func main() {
	flag.Parse()
	err := os.Chdir(*workdir)
	if err != nil {
		log.Fatalf("could not chdir to workingdir: %s", *workdir)
	}

	// Load config
	config, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	var logFile *os.File
	// Set up logging
	if config.Log != "" {
		logFile, err = os.OpenFile(config.Log, syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT, 0666)
		if err != nil {
			log.Fatal("could not open logfile: %s", err.Error())
		}
		log.SetOutput(logFile)
		log.SetPrefix("km-app:\t")
	}

	s, err := km.NewServer("km", config)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Dbmap.Db.Close()

	http.Handle("/", s)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("started on port %d... (%s)\n", config.Port, config.Env)
	http.Serve(listener, nil)
}

//TODO:test this function
func parseConfig(filename string) (config km.Config, err error) {
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = goyaml.Unmarshal(configFile, &config)
	if err != nil {
		return
	}
	return
}
