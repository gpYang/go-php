package php

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	// DriverName only mysql
	DriverName string = "mysql"

	//QuerySelect query of select
	QuerySelect string = "SELECT %s FROM %s"
	//QueryInsert query of insert
	QueryInsert string = "INSERT INTO %s %s VALUES %s"

	// DefalutFields defalut fields of select
	DefalutFields string = "*"
)

var (
	drivers map[string]*DB
)

// Condition the params in func Where or WhereString
type Condition struct {
	field     string
	match     string
	value     string
	query     string
	connector string
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
	wheres     []Condition
	limit      int
	offset     int
	connection *sql.DB
	fields     string
	tables     []Table
	expired    time.Time
	key        string
	timeout    int
	orders     []Order
	groups     []Order
	havings    []Condition
	rollup     bool
	lastSQL    string
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
func (d *DB) Where(field string, match string, value string, params ...string) *DB {
	condition(d, true, field, match, value, params...)
	return d
}

// Having func Having
// .eg DB.Having("SUM(uid)", "=", "2", "or")
func (d *DB) Having(field, match, value string, params ...string) *DB {
	condition(d, true, field, match, value, params...)
	return d
}

// From func Form
// .eg DB.From("user", "u")
func (d *DB) From(name, alias string) *DB {
	d.tables = append(d.tables, Table{
		name:  name,
		alias: alias,
		join:  "",
	})
	return d
}

// Join func Join
// .eg DB.Join("class", "c", "u.class_id = c.id", "inner")
func (d *DB) Join(name, alias, join string, params ...string) *DB {
	j := Table{
		name:  name,
		alias: alias,
		join:  join,
	}
	if len(params) > 0 {
		j.joinType = params[0]
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

// Find func Find select the first row of result
// .eg DB.Find()
func (d *DB) Find() map[string]string {
	d.limit = 1
	query := fmt.Sprintf(QuerySelect, d.fields, parseTable(d)+parseWhere(d, false)+parseOrder(d, true)+parseWhere(d, true)+parseOrder(d, false)+parseLimit(d))
	q := search(d, query)
	defer q.Close()
	cols, _ := q.Columns()
	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}
	row := make(map[string]string)
	for q.Next() {
		if err := q.Scan(scans...); err != nil {
			panic(err)
		}
		for k, v := range values {
			key := cols[k]
			row[key] = string(v)
		}
	}
	return row
}

// Select func Select support to select by the query given in
// .eg DB.Select()
// .eg DB.Select("select * from user")
func (d *DB) Select(params ...string) map[int]map[string]string {
	var query string
	if len(params) > 0 {
		query = params[0]
	} else {
		query = fmt.Sprintf(QuerySelect, d.fields, parseTable(d)+parseWhere(d, false)+parseOrder(d, true)+parseWhere(d, true)+parseOrder(d, false)+parseLimit(d))
	}
	q := search(d, query)
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
			panic(err)
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
	return rows
}

// Count func Count
// .eg DB.Count()
// .eg DB.Count("id")
func (d *DB) Count(params ...string) int {
	var (
		count int
		field = DefalutFields
	)
	if len(params) > 0 {
		field = params[0]
	}
	query := fmt.Sprintf(QuerySelect, fmt.Sprintf("COUNT(%s)", field), parseTable(d)+parseWhere(d, false)+parseWhere(d, true))
	q := search(d, query)
	defer q.Close()
	for q.Next() {
		if err := q.Scan(&count); err != nil {
			panic(err)
		}
	}
	return count
}

//func (d *DB) Insert(data map[int]map[string]interface{}) bool {
//
//	d.connection.Exec()
//}

// Clear return the clean db handle
// .eg DB.Clear()
func (d *DB) Clear() *DB {
	d.limit = 0
	d.offset = 0
	d.fields = DefalutFields
	d.tables = []Table{}
	d.orders = []Order{}
	d.groups = []Order{}
	d.wheres = []Condition{}
	d.havings = []Condition{}
	return d
}

// GetLastSQL get last sql
func (d *DB) GetLastSQL() string {
	return d.lastSQL
}

// Instance return a DB handle in driver pool
// dsn [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
func Instance(dsn string) *DB {
	if drivers == nil {
		drivers = make(map[string]*DB)
	}
	if _, ok := drivers[dsn]; ok {
		return drivers[dsn]
	}
	drivers[dsn] = connet(&DB{
		key:    dsn,
		fields: "*",
	})
	return drivers[dsn]
}

func parseWhere(d *DB, isHaving bool) string {
	var (
		query      string
		conditions []Condition
	)
	if isHaving {
		conditions = d.havings
		if len(conditions) > 0 {
			query += " HAVING "
		}
	} else {
		conditions = d.wheres
		if len(conditions) > 0 {
			query += " WHERE "
		}
	}
	i := 0
	for _, c := range conditions {
		if i != 0 {
			query += " " + c.connector
		}
		if c.query == "" {
			field := strings.Split(c.field, ".")
			query += fmt.Sprintf(" `%s` %s \"%s\"", strings.Join(field, "`.`"), c.match, c.value)
		} else {
			query += c.query
		}
		i++
	}
	return query
}

func parseTable(d *DB) string {
	var query string
	for _, t := range d.tables {
		if t.join == "" {
			query += fmt.Sprintf("`%s` `%s`", t.name, t.alias)
		} else {
			query += fmt.Sprintf("%s JOIN `%s` `%s` ON %s", strings.ToUpper(t.joinType), t.name, t.alias, t.join)
		}
	}
	return query
}

func parseLimit(d *DB) string {
	var query string
	if d.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", d.limit)
	}
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
		if len(orders) > 0 {
			query += " GROUP BY "
			if d.rollup {
				extra = " WITH ROLLUP"
			}
		}
	} else {
		if len(orders) > 0 {
			query += " ORDER BY "
		}
	}
	i := 0
	for _, c := range orders {
		if i != 0 {
			query += ", "
		}
		field := strings.Split(c.field, ".")
		query += fmt.Sprintf(" `%s` %s ", strings.Join(field, "`.`"), c.sort)
		i++
	}
	return query + extra
}

func connet(d *DB) *DB {
	db, err := sql.Open(DriverName, d.key)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	d.connection = db
	setExpired(d)
	return d
}

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

func search(d *DB, query string) *sql.Rows {
	d.lastSQL = query
	if d.expired.Before(time.Now()) {
		connet(d)
	} else {
		setExpired(d)
	}
	q, err := d.connection.Query(query)
	if err != nil {
		panic(err)
	}
	d.Clear()
	return q
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

func condition(d *DB, isHaving bool, field, match, value string, params ...string) {
	cc := Condition{
		field:     field,
		match:     match,
		value:     value,
		connector: "and",
	}
	if len(params) > 0 {
		cc.connector = params[0]
	}
	d.wheres = append(d.wheres, cc)
}
