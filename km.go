package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/FreekKalter/km/lib"
	"launchpad.net/goyaml"
)

func main() {
	workdir := flag.String("workdir", fmt.Sprintf("%s/src/github.com/FreekKalter/km/", os.Getenv("GOPATH")), "directory where support files are located (js/img/index.html)")
	configFile := flag.String("config", "./config.yml", "location of configuration file")
	flag.Parse()
	err := os.Chdir(*workdir)
	if err != nil {
		log.Fatalf("could not chdir to workingdir: %s", *workdir)
	}

	// Load config
	config, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := km.NewServer("km", config)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer s.Dbmap.Db.Close()

	http.Handle("/", s)
	log.Printf("started... (%s)\n", config.Env)

	listener, _ := net.Listen("tcp", ":4001")
	if config.Env == "testing" {
		http.Serve(listener, nil)
	} else {
		http.Serve(listener, nil)
		//fcgi.Serve(listener, nil)
	}
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
