package fetchnse

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/nse-go/db"
	"github.com/sirupsen/logrus"
)

var offdays []string

//Endpoints is for exporting
type Endpoints struct {
	URL    string   `json:"url"`
	Shares []string `json:"shares"`
}

//Parser is the overall leader of this pack
type Parser struct {
	Workers  int               `json:"workers"`
	Refresh  int               `json:"refreshInterval"`
	Indexes  Endpoints         `json:"indexes"`
	Equities Endpoints         `json:"equities"`
	Headers  map[string]string `json:"headers"`
	Timeout  int               `json:"httpTimeout"`
}

//Nseparser will be the object for coordination
type Nseparser struct {
	p       Parser
	log     *logrus.Entry
	payload chan fetchshare
	Done    chan struct{}
	wg      sync.WaitGroup
	lock    sync.RWMutex
	cookie  []string
	db      db.Db
}

type fetchshare struct {
	url   string
	stype string
}

//New creates an object for communicating with this module
func New(Parser Parser, db db.Db, log *logrus.Entry, holidays []string) *Nseparser {
	offdays = holidays
	p := Nseparser{
		p:       Parser,
		log:     log,
		payload: make(chan fetchshare),
		Done:    make(chan struct{}),
		wg:      sync.WaitGroup{},
		db:      db,
	}
	go p.workers()
	return &p
}

func (np *Nseparser) fetchdata(d fetchshare) error {
	np.log.Info("fetching data from ", d.url)
	bt, err := np.getfromnse(d)
	if err != nil {
		return err
	}
	var hd nsejson
	if err := json.Unmarshal(bt, &hd); err != nil {
		return err
	}

	for _, v := range hd.Records.Data {
		if v.CE.Underlying != "" {
			err := np.db.Write(v.CE, "CE")
			if err != nil {
				np.log.Error(err)
			}
		} else if v.PE.Underlying != "" {
			err := np.db.Write(v.PE, "PE")
			if err != nil {
				np.log.Error(err)
			}
		}
	}
	return nil
}
func (np *Nseparser) workers() {
	for i := 0; i < np.p.Workers; i++ {
		np.log.Info("starting worker ", i)
		go func() {
			for {
				i, k := <-np.payload
				if !k {
					return
				}
				err := np.fetchdata(i)
				np.wg.Done()
				if err != nil {
					np.log.Error(err)
				}
			}
		}()
	}

	go func() { np.Done <- struct{}{} }()
}

func (np *Nseparser) isRunnable(stime time.Time) bool {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		np.log.Error(err)
		return false
	}

	for _, v := range offdays {
		if stime.Format("02-Jan-2006") == v {
			return false
		}
	}

	start, err := time.Parse("2/1/2006 15:04:05 MST", strconv.Itoa(stime.Day())+"/"+strconv.Itoa(int(stime.Month()))+"/"+strconv.Itoa(stime.Year())+" 9:00:00 IST")

	if err != nil {
		np.log.Error(err)
		return false
	}
	end, err := time.Parse("2/1/2006 15:04:05 MST", strconv.Itoa(stime.Day())+"/"+strconv.Itoa(int(stime.Month()))+"/"+strconv.Itoa(stime.Year())+" 15:31:59 IST")
	if err != nil {
		np.log.Error(err)
		return false
	}

	if time.Now().Sub(start.In(loc)).Seconds() > 0 && time.Now().Sub(end.In(loc)).Seconds() < 0 {
		if int(stime.Weekday()) > 0 && int(stime.Weekday()) <= 5 {
			return true
		}
	}
	return false
}

//Schedule will call all the endpoints
func (np *Nseparser) Schedule() {
	stime := time.Now()
	np.log.Info("starting task at ", stime.Format("02-Jan-2006 15:04:05"))
	if np.isRunnable(stime) {
		for _, v := range np.p.Equities.Shares {
			np.wg.Add(1)
			np.payload <- fetchshare{url: np.p.Equities.URL + v, stype: "equities"}
		}

		for _, v := range np.p.Indexes.Shares {
			np.wg.Add(1)
			np.payload <- fetchshare{url: np.p.Indexes.URL + v, stype: "indexes"}
		}
		np.wg.Wait()
		np.log.Println("All workers have completed their tasks.")

	}

	np.log.Info("Completed task in ", time.Now().Sub(stime).Nanoseconds(), "ns")

	if stime.Minute() == time.Now().Minute() {
		time.Sleep(time.Duration(60-time.Now().Second()) * time.Second)
	}

	np.Done <- struct{}{}
}
