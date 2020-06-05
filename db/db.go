package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql" //driver for mysql
	"github.com/jmoiron/sqlx"
	"github.com/nse-go/models"
	"github.com/sirupsen/logrus"
)

//PE is call type PE
const (
	PE = "PE"
	CE = "CE"
)

//Db is interface
type Db interface {
	Write(s models.CallPut, t string) error
	Get(f time.Time, n string, t string) ([]models.CallPut, error)
}

//dbobj struct for managing
type dbobj struct {
	db  *sqlx.DB
	log *logrus.Entry
}

//New db initializer
func New(dsn string, log *logrus.Entry) Db {
	db := sqlx.MustConnect("mysql", dsn)
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	var d dbobj
	d.db = db
	d.log = log
	return &d
}

func (d *dbobj) getshareid(name string) (int, error) {
	var s int
	if err := d.db.QueryRow("select stock_id from stocks where name=?", name).Scan(&s); err != nil {
		if name == "" {
			return -1, errors.New("error creating stockname with empty value")
		}
		if err == sql.ErrNoRows {
			r, err := d.db.Exec("insert into stocks(name) value(?)", name)
			if err != nil {
				return -1, err
			}
			id, err := r.LastInsertId()

			if err != nil {
				return -1, err
			}
			return int(id), nil

		} else if err != nil {
			return -1, err
		}
	}
	return s, nil
}

//Write to write data
func (d *dbobj) Write(s models.CallPut, t string) error {
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	id, err := d.getshareid(s.Underlying)
	if err != nil {
		d.log.Error(s)
		return err
	}

	_, err = d.db.Exec("insert into stock_values (stock_id, type, strikePrice, expiryDate, openInterest, changeinOpenInterest, pchangeinOpenInterest, totalTradedVolume, impliedVolatility, lastPrice, `change`, pChange, totalBuyQuantity, totalSellQuantity, bidQty, bidprice, askQty, askPrice, underlyingValue) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", id, t, s.StrikePrice, s.ExpiryDate, s.OpenInterest, s.ChangeinOpenInterest, s.PchangeinOpenInterest, s.TotalTradedVolume, s.ImpliedVolatility, s.LastPrice, s.Change, s.PChange, s.TotalBuyQuantity, s.TotalSellQuantity, s.BidQty, s.Bidprice, s.AskQty, s.AskPrice, s.UnderlyingValue)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil

}

// Get to fetch all entries
func (d *dbobj) Get(from time.Time, name string, ctype string) ([]models.CallPut, error) {
	var a []models.CallPut
	return a, nil
}
