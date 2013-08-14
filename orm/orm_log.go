package orm

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

type Log struct {
	*log.Logger
}

func NewLog(out io.Writer) *Log {
	d := new(Log)
	d.Logger = log.New(out, "[ORM]", 1e9)
	return d
}

func debugLogQueies(alias *alias, operaton, query string, t time.Time, err error, args ...interface{}) {
	sub := time.Now().Sub(t) / 1e5
	elsp := float64(int(sub)) / 10.0
	flag := "  OK"
	if err != nil {
		flag = "FAIL"
	}
	con := fmt.Sprintf(" - %s - [Queries/%s] - [%s / %11s / %7.1fms] - [%s]", t.Format(format_DateTime), alias.Name, flag, operaton, elsp, query)
	cons := make([]string, 0, len(args))
	for _, arg := range args {
		cons = append(cons, fmt.Sprintf("%v", arg))
	}
	if len(cons) > 0 {
		con += fmt.Sprintf(" - `%s`", strings.Join(cons, "`, `"))
	}
	if err != nil {
		con += " - " + err.Error()
	}
	DebugLog.Println(con)
}

type stmtQueryLog struct {
	alias *alias
	query string
	stmt  stmtQuerier
}

var _ stmtQuerier = new(stmtQueryLog)

func (d *stmtQueryLog) Close() error {
	a := time.Now()
	err := d.stmt.Close()
	debugLogQueies(d.alias, "st.Close", d.query, a, err)
	return err
}

func (d *stmtQueryLog) Exec(args ...interface{}) (sql.Result, error) {
	a := time.Now()
	res, err := d.stmt.Exec(args...)
	debugLogQueies(d.alias, "st.Exec", d.query, a, err, args...)
	return res, err
}

func (d *stmtQueryLog) Query(args ...interface{}) (*sql.Rows, error) {
	a := time.Now()
	res, err := d.stmt.Query(args...)
	debugLogQueies(d.alias, "st.Query", d.query, a, err, args...)
	return res, err
}

func (d *stmtQueryLog) QueryRow(args ...interface{}) *sql.Row {
	a := time.Now()
	res := d.stmt.QueryRow(args...)
	debugLogQueies(d.alias, "st.QueryRow", d.query, a, nil, args...)
	return res
}

func newStmtQueryLog(alias *alias, stmt stmtQuerier, query string) stmtQuerier {
	d := new(stmtQueryLog)
	d.stmt = stmt
	d.alias = alias
	d.query = query
	return d
}

type dbQueryLog struct {
	alias *alias
	db    dbQuerier
	tx    txer
	txe   txEnder
}

var _ dbQuerier = new(dbQueryLog)
var _ txer = new(dbQueryLog)
var _ txEnder = new(dbQueryLog)

func (d *dbQueryLog) Prepare(query string) (*sql.Stmt, error) {
	a := time.Now()
	stmt, err := d.db.Prepare(query)
	debugLogQueies(d.alias, "db.Prepare", query, a, err)
	return stmt, err
}

func (d *dbQueryLog) Exec(query string, args ...interface{}) (sql.Result, error) {
	a := time.Now()
	res, err := d.db.Exec(query, args...)
	debugLogQueies(d.alias, "db.Exec", query, a, err, args...)
	return res, err
}

func (d *dbQueryLog) Query(query string, args ...interface{}) (*sql.Rows, error) {
	a := time.Now()
	res, err := d.db.Query(query, args...)
	debugLogQueies(d.alias, "db.Query", query, a, err, args...)
	return res, err
}

func (d *dbQueryLog) QueryRow(query string, args ...interface{}) *sql.Row {
	a := time.Now()
	res := d.db.QueryRow(query, args...)
	debugLogQueies(d.alias, "db.QueryRow", query, a, nil, args...)
	return res
}

func (d *dbQueryLog) Begin() (*sql.Tx, error) {
	a := time.Now()
	tx, err := d.db.(txer).Begin()
	debugLogQueies(d.alias, "db.Begin", "START TRANSACTION", a, err)
	return tx, err
}

func (d *dbQueryLog) Commit() error {
	a := time.Now()
	err := d.db.(txEnder).Commit()
	debugLogQueies(d.alias, "tx.Commit", "COMMIT", a, err)
	return err
}

func (d *dbQueryLog) Rollback() error {
	a := time.Now()
	err := d.db.(txEnder).Rollback()
	debugLogQueies(d.alias, "tx.Rollback", "ROLLBACK", a, err)
	return err
}

func (d *dbQueryLog) SetDB(db dbQuerier) {
	d.db = db
}

func newDbQueryLog(alias *alias, db dbQuerier) dbQuerier {
	d := new(dbQueryLog)
	d.alias = alias
	d.db = db
	return d
}
