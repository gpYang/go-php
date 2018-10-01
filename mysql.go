package php

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	// DriverName only mysql
	DriverName string = "mysql"

	// QuerySelect query of select
	QuerySelect string = "SELECT %s FROM %s"
	// QueryInsert query of insert
	QueryInsert string = "INSERT INTO %s %s"
	// QueryUpdate query of update
	QueryUpdate string = "UPDATE %s SET %s"
	// QueryDelete query of delete
	QueryDelete string = "DELETE FROM %s"

	// DefalutFields defalut fields of select
	DefalutFields string = "*"
)

var (
	logSQL  func(q string)
	drivers map[string]*DB
)

// Condition the params in func Where or WhereString
type Condition struct {
	field     string
	match     string
	query     string
	connector string
	value     interface{}
}

// Order sort by field
type Order struct {
	field string
	sort  string
}

// Table the params in func From or Join
type Table struct {
	name     string
	alias    string
	join     string
	joinType string
}

// DB db handle
type DB struct {
	transactions []*sql.Tx
	tx           *sql.Tx
	dsn          string
	err          error
	wheres       []Condition
	limit        int
	offset       int
	fields       string
	tables       []Table
	expired      time.Time
	timeout      int
	orders       []Order
	groups       []Order
	havings      []Condition
	rollup       bool
	params       []interface{}
	lastSQL      string
	connection   *sql.DB
}

// LogSQL set function of log sql
func LogSQL(log func(q string)) {
	logSQL = log
}

// Field define the fields hope to select
// .eg DB.Field("`u`.`name`, `c`.`name`")
func (d *DB) Field(fields string) *DB {
	d.fields = fields
	return d
}

// WhereString func Where with full string
// .eg DB.WhereString("`uid` = \"2\"", "or")
func (d *DB) WhereString(c string, params ...string) *DB {
	conditionString(d, false, c, params...)
	return d
}

// HavingString func Having with full string
// .eg DB.Having("SUM(`uid`) = \"2\"", "or")
func (d *DB) HavingString(c string, params ...string) *DB {
	conditionString(d, true, c, params...)
	return d
}

// Where func Where
// .eg DB.Where("u.uid", "=", "2", "or")
func (d *DB) Where(field, match string, value interface{}, params ...string) *DB {
	condition(d, false, field, match, value, params...)
	return d
}

// Having func Having
// .eg DB.Having("SUM(uid)", "=", "2", "or")
func (d *DB) Having(field, match string, value interface{}, params ...string) *DB {
	condition(d, true, field, match, value, params...)
	return d
}

// From func Form
// .eg DB.From("user", "u")
// .eg DB.From("user")
func (d *DB) From(name string, params ...string) *DB {
	f := Table{
		name: name,
		join: "",
	}
	if len(params) > 0 {
		f.alias = params[0]
	}
	d.tables = append(d.tables, f)
	return d
}

// Join func Join
// .eg DB.Join("class", "u.class_id = c.id", "c", "inner")
// .eg DB.Join("class", "user.class_id = class.id")
func (d *DB) Join(name, join string, params ...string) *DB {
	j := Table{
		name: name,
		join: join,
	}
	if len(params) > 0 {
		j.alias = params[0]
	}
	if len(params) > 1 {
		j.joinType = params[1]
	}
	d.tables = append(d.tables, j)
	return d
}

// Limit func Limit
// .eg DB.Limit(1)
func (d *DB) Limit(limit int) *DB {
	d.limit = limit
	return d
}

// Offset func Offset
// .eg DB.Offset(100)
func (d *DB) Offset(offset int) *DB {
	d.offset = offset
	return d
}

// Order func Order
// .eg DB.Order("age", "desc")
func (d *DB) Order(field, sort string) *DB {
	d.orders = append(d.orders, Order{
		field: field,
		sort:  sort,
	})
	return d
}

// Group func Group
// .eg DB.Group("age", "desc")
func (d *DB) Group(field, sort string) *DB {
	d.groups = append(d.groups, Order{
		field: field,
		sort:  sort,
	})
	return d
}

// Rollup add with rollup where using grouping
// .eg DB.Rollup()
func (d *DB) Rollup() *DB {
	d.rollup = true
	return d
}

// Find return result row and error
// select the first row of result
// .eg DB.Find()
func (d *DB) Find() (map[string]string, error) {
	var err error
	defer d.Clear()
	d.limit = 1
	row := make(map[string]string)
	query := fmt.Sprintf(QuerySelect, d.fields, parseTable(d)+parseWhere(d, false)+parseOrder(d, true)+parseWhere(d, true)+parseOrder(d, false)+parseLimit(d)+parseOffset(d))
	q, err := search(d, query)
	defer q.Close()
	if err != nil {
		return row, err
	}
	cols, _ := q.Columns()
	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}
	for q.Next() {
		if err := q.Scan(scans...); err != nil {
			return row, err
		}
		for k, v := range values {
			key := cols[k]
			row[key] = string(v)
		}
	}
	return row, err
}

// Select return result map and error
// support to select by the query given in
// .eg DB.Select()
// .eg DB.Select("select * from user")
func (d *DB) Select(params ...string) (map[int]map[string]string, error) {
	var (
		query string
		err   error
	)
	defer d.Clear()
	if len(params) > 0 {
		query = params[0]
	} else {
		query = fmt.Sprintf(QuerySelect, d.fields, parseTable(d)+parseWhere(d, false)+parseOrder(d, true)+parseWhere(d, true)+parseOrder(d, false)+parseLimit(d)+parseOffset(d))
	}
	q, err := search(d, query)
	defer q.Close()
	cols, _ := q.Columns()
	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}
	rows := make(map[int]map[string]string)
	i := 0
	for q.Next() {
		if err := q.Scan(scans...); err != nil {
			return rows, err
		}
		row := make(map[string]string)
		for k, v := range values {
			key := cols[k]
			if v != nil {
				row[key] = string(v)
			}
		}
		rows[i] = row
		i++
	}
	return rows, err
}

// Count return count and error
// .eg DB.Count()
// .eg DB.Count("id")
func (d *DB) Count(params ...string) (int64, error) {
	var (
		err   error
		count int64
		field = DefalutFields
	)
	defer d.Clear()
	if len(params) > 0 {
		field = params[0]
	}
	query := fmt.Sprintf(QuerySelect, fmt.Sprintf("COUNT(%s)", field), parseTable(d)+parseWhere(d, false)+parseWhere(d, true))
	q, err := search(d, query)
	if err != nil {
		return count, err
	}
	defer q.Close()
	for q.Next() {
		if err := q.Scan(&count); err != nil {
			return count, err
		}
	}
	return count, err
}

// Insert return lastid and error
// .eg DB.Insert([]string{"name", "age", "class_id"}, map[int]map[string]interface{}{0: map[string]interface{}{"name": "apple", "age": 12}, 1: map[string]interface{}{"name": "king"}})
func (d *DB) Insert(fields []string, data map[int]map[string]interface{}) (int64, error) {
	var (
		res    sql.Result
		err    error
		lastID int64
	)
	defer d.Clear()
	d.params = make([]interface{}, 0)
	d.fields = "`" + strings.Join(fields, "`,`") + "`"
	query := fmt.Sprintf(QueryInsert, parseTable(d), "("+d.fields+")"+" VALUE "+parseInsert(d, data))
	res, err = exec(d, query)
	if err != nil {
		return lastID, err
	}
	return res.LastInsertId()
}

// InsertSelect return lastid and error
// .eg DB.InsertSelect("SELECT * FROM user_bak")
func (d *DB) InsertSelect(s string) (int64, error) {
	var (
		res    sql.Result
		err    error
		lastID int64
	)
	defer d.Clear()
	d.params = make([]interface{}, 0)
	query := fmt.Sprintf(QueryInsert, parseTable(d), s)
	res, err = exec(d, query)
	if err != nil {
		return lastID, err
	}
	return res.LastInsertId()
}

// Update return affect-rows and error
// .eg DB.Update()
func (d *DB) Update(data map[string]interface{}) (int64, error) {
	var (
		res          sql.Result
		err          error
		rowsAffected int64
	)
	defer d.Clear()
	d.params = make([]interface{}, 0)
	query := fmt.Sprintf(QueryUpdate, parseTable(d), parseUpdate(d, data)+parseWhere(d, false)+parseOrder(d, false)+parseLimit(d))
	res, err = exec(d, query)
	if err != nil {
		return rowsAffected, err
	}
	return res.RowsAffected()
}

// Delete return affect-rows and error
// .eg DB.Delete()
func (d *DB) Delete() (int64, error) {
	query := fmt.Sprintf(QueryDelete, parseTable(d)+parseWhere(d, false))
	var (
		res          sql.Result
		err          error
		rowsAffected int64
	)
	defer d.Clear()
	res, err = exec(d, query)
	if err != nil {
		return rowsAffected, err
	}
	return res.RowsAffected()
}

// Exec sql.Exec
// .eg DB.Exec("UPDATE user SET name = "Aaron" WHERE id = 1")
func (d *DB) Exec(query string, params ...interface{}) (sql.Result, error) {
	defer d.Clear()
	d.params = params
	return exec(d, query)
}

// Begin starts a transaction
func (d *DB) Begin() (*DB, error) {
	tx, err := d.connection.Begin()
	if err != nil {
		return nil, err
	}
	defer setLastSQL(d, "BEGIN")
	d.tx = tx
	d.transactions = append(d.transactions, tx)
	return d, err
}

// Commit commits the transaction
func (d *DB) Commit() (err error) {
	if d.tx == nil {
		return errors.New("no transation running")
	}
	defer setLastSQL(d, "COMMIT")
	err = d.tx.Commit()
	if err == nil {
		d.transactions = d.transactions[:len(d.transactions)-1]
		if len(d.transactions) > 0 {
			d.tx = d.transactions[len(d.transactions)-1]
		} else {
			d.tx = nil
		}
	}
	return err
}

// Rollback aborts the transaction
func (d *DB) Rollback() (err error) {
	if d.tx == nil {
		return errors.New("no transation running")
	}
	defer setLastSQL(d, "ROLLBACK")
	err = d.tx.Rollback()
	if err == nil {
		d.transactions = d.transactions[:len(d.transactions)-1]
		if len(d.transactions) > 0 {
			d.tx = d.transactions[len(d.transactions)-1]
		} else {
			d.tx = nil
		}
	}
	return err
}

// Clear return the clean db handle
// .eg DB.Clear()
func (d *DB) Clear() *DB {
	d.limit = 0
	d.offset = 0
	d.rollup = false
	d.fields = DefalutFields
	d.tables = []Table{}
	d.orders = []Order{}
	d.groups = []Order{}
	d.wheres = []Condition{}
	d.params = make([]interface{}, 0)
	d.havings = []Condition{}
	return d
}

// GetLastSQL get last sql
func (d *DB) GetLastSQL() string {
	return d.lastSQL
}

// Instance return a DB handle in driver pool
// dsn [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
func Instance(dsn string) (*DB, error) {
	var err error
	if drivers == nil {
		drivers = make(map[string]*DB)
	}
	if _, ok := drivers[dsn]; ok {
		return drivers[dsn], nil
	}
	drivers[dsn], err = connect(&DB{
		dsn:    dsn,
		fields: "*",
	})
	if err != nil {
		return drivers[dsn].Clear(), err
	}
	return drivers[dsn].Clear(), err
}

// connect always return a new connection with server
func connect(d *DB) (*DB, error) {
	db, err := sql.Open(DriverName, d.dsn)
	if err != nil {
		return d, err
	}
	if err := db.Ping(); err != nil {
		return d, err
	}
	d.connection = db
	setExpired(d)
	return d, nil
}

// setExpired set expired time from "wait_timeout"
func setExpired(d *DB) {
	var (
		k string
	)
	if d.timeout == 0 {
		d.connection.QueryRow("show variables WHERE Variable_name = \"wait_timeout\"").Scan(&k, &d.timeout)
	}
	ss, _ := time.ParseDuration(fmt.Sprintf("%ds", d.timeout-2))
	d.connection.SetConnMaxLifetime(ss)
	d.expired = time.Now().Add(ss)
}

// setLastSQL set the last sql
// finally change to an executable sql
func setLastSQL(d *DB, query string) {
	qq := strings.Replace(query, "?", "'%v'", -1)
	d.lastSQL = fmt.Sprintf(qq, d.params...)
	logSQL(d.lastSQL)
}

// search find the result
func search(d *DB, q string) (*sql.Rows, error) {
	if d.expired.Before(time.Now()) {
		connect(d)
	} else {
		setExpired(d)
	}
	return query(d, q)
}

// query do query
func query(d *DB, query string) (q *sql.Rows, err error) {
	setLastSQL(d, query)
	if strings.Contains(query, "?") {
		if d.tx != nil {
			q, err = d.tx.Query(query, d.params...)
		} else {
			q, err = d.connection.Query(query, d.params...)
		}
	} else {
		if d.tx != nil {
			q, err = d.tx.Query(query)
		} else {
			q, err = d.connection.Query(query)
		}
	}
	return q, err
}

// exec do exec
func exec(d *DB, query string) (res sql.Result, err error) {
	setLastSQL(d, query)
	if strings.Contains(query, "?") {
		if d.tx != nil {
			res, err = d.tx.Exec(query, d.params...)
		} else {
			res, err = d.connection.Exec(query, d.params...)
		}
	} else {
		if d.tx != nil {
			res, err = d.tx.Exec(query)
		} else {
			res, err = d.connection.Exec(query)
		}
	}
	return res, err
}

func conditionString(d *DB, isHaving bool, c string, params ...string) {
	cc := Condition{
		query:     c,
		connector: "and",
	}
	if len(params) > 0 {
		cc.connector = params[0]
	}
	if isHaving {
		d.havings = append(d.havings, cc)
	} else {
		d.wheres = append(d.wheres, cc)
	}
}

func condition(d *DB, isHaving bool, field, match string, value interface{}, params ...string) {
	cc := Condition{
		field:     field,
		match:     match,
		connector: "and",
		value:     value,
	}
	if len(params) > 0 {
		cc.connector = params[0]
	}
	if isHaving {
		d.havings = append(d.havings, cc)
	} else {
		d.wheres = append(d.wheres, cc)
	}
}

func parseWhere(d *DB, isHaving bool) string {
	var (
		query      string
		conditions []Condition
	)
	if isHaving {
		conditions = d.havings
		if len(conditions) > 0 {
			query += " HAVING"
		}
	} else {
		conditions = d.wheres
		if len(conditions) > 0 {
			query += " WHERE"
		}
	}
	i := 0
	for _, c := range conditions {
		if i != 0 {
			query += " " + strings.ToUpper(c.connector)
		}
		if c.query == "" {
			var field string
			if strings.Contains(c.field, "(") {
				reg := regexp.MustCompile("(\\S+)\\(([^\\)]+)\\)")
				matches := reg.FindStringSubmatch(c.field)
				field = fmt.Sprintf("%s(`%s`)", matches[1], strings.Join(strings.Split(matches[2], "."), "`.`"))
			} else {
				field = fmt.Sprintf("`%s`", strings.Join(strings.Split(c.field, "."), "`.`"))
			}
			query += fmt.Sprintf(" %s %s ?", field, c.match)
			d.params = append(d.params, c.value)
		} else {
			query += " " + c.query
		}
		i++
	}
	return query
}

func parseTable(d *DB) string {
	var (
		query string
		table string
	)
	for _, t := range d.tables {
		table = fmt.Sprintf("`%s`", t.name)
		if t.alias != "" {
			table += " `" + t.alias + "`"
		}

		if t.join == "" {
			query += table
		} else {
			query += fmt.Sprintf("%s JOIN %s ON %s", strings.ToUpper(t.joinType), table, t.join)
		}
	}
	return query
}

func parseLimit(d *DB) string {
	var query string
	if d.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", d.limit)
	}
	return query
}

func parseOffset(d *DB) string {
	var query string
	if d.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", d.offset)
	}
	return query
}

func parseOrder(d *DB, isGroup bool) string {
	var (
		query  string
		orders []Order
		extra  string
	)
	if isGroup {
		orders = d.groups
		if len(orders) > 0 {
			query += " GROUP BY"
			if d.rollup {
				extra = " WITH ROLLUP"
			}
		}
	} else {
		orders = d.orders
		if len(orders) > 0 {
			query += " ORDER BY"
		}
	}
	i := 0
	for _, c := range orders {
		if i != 0 {
			query += ", "
		}
		field := strings.Split(c.field, ".")
		query += fmt.Sprintf(" `%s` %s", strings.Join(field, "`.`"), strings.ToUpper(c.sort))
		i++
	}
	return query + extra
}

func parseInsert(d *DB, datas map[int]map[string]interface{}) string {
	var query string
	fields := strings.Split(d.fields, ",")
	i := 0
	for _, data := range datas {
		if i != 0 {
			query += ", "
		}
		query += "("
		for _, field := range fields {
			field = strings.Trim(field, "`")
			if data[field] != nil {
				query += "?,"
				d.params = append(d.params, data[field])
			} else {
				query += "NULL,"
			}
		}
		query = strings.TrimRight(query, ",") + ")"
		i++
	}
	return query
}

func parseUpdate(d *DB, data map[string]interface{}) string {
	var query string
	i := 0
	for key, value := range data {
		if i != 0 {
			query += ", "
		}
		k := strings.Split(key, ".")
		query += fmt.Sprintf("`%s` = ?", strings.Join(k, "`.`"))
		d.params = append(d.params, value)
		i++
	}
	return query
}
