package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"

	"github.com/nse-go/db"
	fetchnse "github.com/nse-go/nse"
	"github.com/sirupsen/logrus"
)

type config struct {
	Parser   fetchnse.Parser `json:"parser"`
	Db       string          `json:"db"`
	Holidays []string        `json:"holidays"`
}

func main() {
	var c config
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: false,
		DisableTimestamp: true})

	logf := flag.String("c", "", "location of the config file")
	flag.Parse()
	_, err := os.Stat(*logf)
	if err != nil {
		log.Fatal(err)
	}

	bt, err := ioutil.ReadFile(*logf)
	if err != nil {
		log.Fatalln("Error while reading config file. ", err)
	}
	if err := json.Unmarshal(bt, &c); err != nil {
		log.Fatalln("Error while reading data from config file. ", err)
	}
	db := db.New(c.Db, log.WithField("module", "db"))
	nse := fetchnse.New(c.Parser, db, log.WithField("module", "fetchnse"), c.Holidays)
	for {
		select {
		case <-nse.Done:
			go nse.Schedule()

		}
	}
}
