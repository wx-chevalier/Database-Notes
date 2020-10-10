package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Percona-Lab/mysql_random_data_load/internal/getters"
	"github.com/Percona-Lab/mysql_random_data_load/tableparser"
	"github.com/go-ini/ini"
	"github.com/go-sql-driver/mysql"
	"github.com/gosuri/uiprogress"
	"github.com/kr/pretty"

	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type cliOptions struct {
	app *kingpin.Application

	// Arguments
	Schema    *string
	TableName *string
	Rows      *int
	// Flags
	BulkSize   *int
	ConfigFile *string
	Debug      *bool
	Factor     *float64
	Host       *string
	MaxRetries *int
	MaxThreads *int
	NoProgress *bool
	Pass       *string
	Port       *int
	Print      *bool
	Samples    *int64
	User       *string
	Version    *bool
}

type mysqlOptions struct {
	Host     string
	Password string
	Port     int
	Sock     string
	User     string
}

var (
	opts *cliOptions

	validFunctions = []string{"int", "string", "date", "date_in_range"}
	maxValues      = map[string]int64{
		"tinyint":   0XF,
		"smallint":  0xFF,
		"mediumint": 0x7FFFF,
		"int":       0x7FFFFFFF,
		"integer":   0x7FFFFFFF,
		"float":     0x7FFFFFFF,
		"decimal":   0x7FFFFFFF,
		"double":    0x7FFFFFFF,
		"bigint":    0x7FFFFFFFFFFFFFFF,
	}

	Version   = "0.0.0."
	Commit    = "<sha1>"
	Branch    = "branch-name"
	Build     = "2017-01-01"
	GoVersion = "1.9.2"
)

type getter interface {
	Value() interface{}
	Quote() string
	String() string
}
type insertValues []getter
type insertFunction func(*sql.DB, string, chan int, chan bool, *sync.WaitGroup)

const (
	defaultMySQLConfigSection = "client"
	defaultConfigFile         = "~/.my.cnf"
)

func main() {

	opts, err := processCliParams()
	if err != nil {
		log.Fatal(err.Error())
	}

	if *opts.Version {
		fmt.Printf("Version   : %s\n", Version)
		fmt.Printf("Commit    : %s\n", Commit)
		fmt.Printf("Branch    : %s\n", Branch)
		fmt.Printf("Build     : %s\n", Build)
		fmt.Printf("Go version: %s\n", GoVersion)
		return
	}

	address := *opts.Host
	net := "unix"
	if address != "localhost" {
		net = "tcp"
	}
	if *opts.Port != 0 {
		address = fmt.Sprintf("%s:%d", address, *opts.Port)
	}

	dsn := mysql.Config{
		User:                 *opts.User,
		Passwd:               *opts.Pass,
		Addr:                 address,
		Net:                  net,
		DBName:               "",
		ParseTime:            true,
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(100)

	// SET TimeZone to UTC to avoid errors due to random dates & daylight saving valid values
	if _, err = db.Exec(`SET @@session.time_zone = "+00:00"`); err != nil {
		log.Printf("Cannot set time zone to UTC: %s\n", err)
		db.Close()
		os.Exit(1)
	}

	table, err := tableparser.NewTable(db, *opts.Schema, *opts.TableName)
	if err != nil {
		log.Printf("cannot get table %s struct: %s", *opts.TableName, err)
		db.Close()
		os.Exit(1)
	}

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	if *opts.Debug {
		log.SetLevel(log.DebugLevel)
		*opts.NoProgress = true
	}
	log.Debug(pretty.Sprint(table))

	if len(table.Triggers) > 0 {
		log.Warnf("There are triggers on the %s table that might affect this process:", *opts.TableName)
		for _, t := range table.Triggers {
			log.Warnf("Trigger %q, %s %s", t.Trigger, t.Timing, t.Event)
			log.Warnf("Statement: %s", t.Statement)
		}
	}

	if *opts.BulkSize > *opts.Rows {
		*opts.BulkSize = *opts.Rows
	}

	if opts.MaxThreads == nil {
		*opts.MaxThreads = runtime.NumCPU() * 10
	}

	if *opts.MaxThreads < 1 {
		*opts.MaxThreads = 1
	}

	if !*opts.Print {
		log.Info("Starting")
	}

	// Example: want 11 rows with bulksize 4:
	// count = int(11 / 4) = 2 -> 2 bulk inserts having 4 rows each = 8 rows
	// We need to run this insert twice:
	// INSERT INTO table (f1, f2) VALUES (?, ?), (?, ?), (?, ?), (?, ?)
	// remainder = rows - count = 11 - 8 = 3
	// And then, we need to run this insert once to complete 11 rows
	// INSERT INTO table (f1, f2) VALUES (?, ?), (?, ?), (?, ?)
	newLineOnEachRow := false
	count := *opts.Rows / *opts.BulkSize
	remainder := *opts.Rows - count**opts.BulkSize
	semaphores := makeSemaphores(*opts.MaxThreads)
	rowValues := makeValueFuncs(db, table.Fields)
	log.Debugf("Must run %d bulk inserts having %d rows each", count, *opts.BulkSize)

	runInsertFunc := runInsert
	if *opts.Print {
		*opts.MaxThreads = 1
		*opts.NoProgress = true
		newLineOnEachRow = true
		runInsertFunc = func(db *sql.DB, insertQuery string, resultsChan chan int, sem chan bool, wg *sync.WaitGroup) {
			fmt.Println(insertQuery)
			resultsChan <- *opts.BulkSize
			sem <- true
			wg.Done()
		}
	}

	bar := uiprogress.AddBar(*opts.Rows).AppendCompleted().PrependElapsed()
	if !*opts.NoProgress {
		uiprogress.Start()
	}

	okCount, err := run(db, table, bar, semaphores, rowValues, count, *opts.BulkSize, runInsertFunc, newLineOnEachRow)
	if err != nil {
		log.Errorln(err)
	}
	var okrCount, okiCount int // remainder & individual inserts OK count
	if remainder > 0 {
		log.Debugf("Must run 1 extra bulk insert having %d rows, to complete %d rows", remainder, *opts.Rows)
		okrCount, err = run(db, table, bar, semaphores, rowValues, 1, remainder, runInsertFunc, newLineOnEachRow)
		if err != nil {
			log.Errorln(err)
		}
	}

	// If there were errors and at this point we have less rows than *rows,
	// retry adding individual rows (no bulk inserts)
	totalOkCount := okCount + okrCount
	retries := 0
	if totalOkCount < *opts.Rows {
		log.Debugf("Running extra %d individual inserts (duplicated keys?)", *opts.Rows-totalOkCount)
	}
	for totalOkCount < *opts.Rows && retries < *opts.MaxRetries {
		okiCount, err = run(db, table, bar, semaphores, rowValues, *opts.Rows-totalOkCount, 1, runInsertFunc, newLineOnEachRow)
		if err != nil {
			log.Errorf("Cannot run extra insert: %s", err)
		}

		retries++
		totalOkCount += okiCount
	}

	time.Sleep(500 * time.Millisecond) // Let the progress bar to update
	if !*opts.Print {
		log.Printf("%d rows inserted", totalOkCount)
	}
	db.Close()
}

func run(db *sql.DB, table *tableparser.Table, bar *uiprogress.Bar, sem chan bool,
	rowValues insertValues, count, bulkSize int, insertFunc insertFunction, newLineOnEachRow bool) (int, error) {
	if count == 0 {
		return 0, nil
	}
	var wg sync.WaitGroup
	insertQuery := generateInsertStmt(table)
	rowsChan := make(chan []getter, 1000)
	okRowsChan := countRowsOK(count, bar)

	go generateInsertData(count*bulkSize, rowValues, rowsChan)
	defaultSeparator1 := ""
	if newLineOnEachRow {
		defaultSeparator1 = "\n"
	}

	i := 0
	rowsCount := 0
	sep1, sep2 := defaultSeparator1, ""

	for i < count {
		rowData := <-rowsChan
		rowsCount++
		insertQuery += sep1 + " ("
		for _, field := range rowData {
			insertQuery += sep2 + field.Quote()
			sep2 = ", "
		}
		insertQuery += ")"
		sep1 = ", "
		if newLineOnEachRow {
			sep1 += "\n"
		}
		sep2 = ""
		if rowsCount < bulkSize {
			continue
		}

		insertQuery += ";\n"
		<-sem
		wg.Add(1)
		go insertFunc(db, insertQuery, okRowsChan, sem, &wg)

		insertQuery = generateInsertStmt(table)
		sep1, sep2 = defaultSeparator1, ""
		rowsCount = 0
		i++
	}

	wg.Wait()
	okCount := <-okRowsChan
	return okCount, nil
}

func makeSemaphores(count int) chan bool {
	sem := make(chan bool, count)
	for i := 0; i < count; i++ {
		sem <- true
	}
	return sem
}

// This go-routine keeps track of how many rows were actually inserted
// by the bulk inserts since one or more rows could generate duplicated
// keys so, not allways the number of inserted rows = number of rows in
// the bulk insert

func countRowsOK(count int, bar *uiprogress.Bar) chan int {
	var totalOk int
	resultsChan := make(chan int, 10000)
	go func() {
		for i := 0; i < count; i++ {
			okCount := <-resultsChan
			for j := 0; j < okCount; j++ {
				bar.Incr()
			}
			totalOk += okCount
		}
		resultsChan <- totalOk
	}()
	return resultsChan
}

// generateInsertData will generate 'rows' items, where each item in the channel has 'bulkSize' rows.
// For example:
// We need to load 6 rows using a bulk insert having 2 rows per insert, like this:
// INSERT INTO table (f1, f2, f3) VALUES (?, ?, ?), (?, ?, ?)
//
// This function will put into rowsChan 3 elements, each one having the values for 2 rows:
// rowsChan <- [ v1-1, v1-2, v1-3, v2-1, v2-2, v2-3 ]
// rowsChan <- [ v3-1, v3-2, v3-3, v4-1, v4-2, v4-3 ]
// rowsChan <- [ v1-5, v5-2, v5-3, v6-1, v6-2, v6-3 ]
//
func generateInsertData(count int, values insertValues, rowsChan chan []getter) {
	for i := 0; i < count; i++ {
		insertRow := make([]getter, 0, len(values))
		for _, val := range values {
			insertRow = append(insertRow, val)
		}
		rowsChan <- insertRow
	}
}

func generateInsertStmt(table *tableparser.Table) string {
	fields := getFieldNames(table.Fields)
	query := fmt.Sprintf("INSERT IGNORE INTO %s.%s (%s) VALUES ",
		backticks(table.Schema),
		backticks(table.Name),
		strings.Join(fields, ","),
	)
	return query
}

func runInsert(db *sql.DB, insertQuery string, resultsChan chan int, sem chan bool, wg *sync.WaitGroup) {
	result, err := db.Exec(insertQuery)
	if err != nil {
		log.Debugf("Cannot run insert: %s", err)
		resultsChan <- 0
		sem <- true
		wg.Done()
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Errorf("Cannot get rows affected after insert: %s", err)
	}
	resultsChan <- int(rowsAffected)
	sem <- true
	wg.Done()
}

// makeValueFuncs returns an array of functions to generate all the values needed for a single row
func makeValueFuncs(conn *sql.DB, fields []tableparser.Field) insertValues {
	var values []getter
	for _, field := range fields {
		if !field.IsNullable && field.ColumnKey == "PRI" && strings.Contains(field.Extra, "auto_increment") {
			continue
		}
		if field.Constraint != nil {
			samples, err := getSamples(conn, field.Constraint.ReferencedTableSchema,
				field.Constraint.ReferencedTableName,
				field.Constraint.ReferencedColumnName,
				100, field.DataType)
			if err != nil {
				log.Printf("cannot get samples for field %q: %s\n", field.ColumnName, err)
				continue
			}
			values = append(values, getters.NewRandomSample(field.ColumnName, samples, field.IsNullable))
			continue
		}
		maxValue := maxValues["bigint"]
		if m, ok := maxValues[field.DataType]; ok {
			maxValue = m
		}
		switch field.DataType {
		case "tinyint", "smallint", "mediumint", "int", "integer", "bigint":
			values = append(values, getters.NewRandomInt(field.ColumnName, maxValue, field.IsNullable))
		case "float", "decimal", "double":
			values = append(values, getters.NewRandomDecimal(field.ColumnName,
				field.NumericPrecision.Int64-field.NumericScale.Int64, field.IsNullable))
		case "char", "varchar":
			values = append(values, getters.NewRandomString(field.ColumnName,
				field.CharacterMaximumLength.Int64, field.IsNullable))
		case "date":
			values = append(values, getters.NewRandomDate(field.ColumnName, field.IsNullable))
		case "datetime", "timestamp":
			values = append(values, getters.NewRandomDateTime(field.ColumnName, field.IsNullable))
		case "tinyblob", "tinytext", "blob", "text", "mediumtext", "mediumblob", "longblob", "longtext":
			values = append(values, getters.NewRandomString(field.ColumnName,
				field.CharacterMaximumLength.Int64, field.IsNullable))
		case "time":
			values = append(values, getters.NewRandomTime(field.IsNullable))
		case "year":
			values = append(values, getters.NewRandomIntRange(field.ColumnName, int64(time.Now().Year()-1),
				int64(time.Now().Year()), field.IsNullable))
		case "enum", "set":
			values = append(values, getters.NewRandomEnum(field.SetEnumVals, field.IsNullable))
		case "binary", "varbinary":
			values = append(values, getters.NewRandomBinary(field.ColumnName, field.CharacterMaximumLength.Int64, field.IsNullable))
		default:
			log.Printf("cannot get field type: %s: %s\n", field.ColumnName, field.DataType)
		}
	}

	return values
}

func getFieldNames(fields []tableparser.Field) []string {
	var fieldNames []string
	for _, field := range fields {
		if !isSupportedType(field.DataType) {
			continue
		}
		if !field.IsNullable && field.ColumnKey == "PRI" &&
			strings.Contains(field.Extra, "auto_increment") {
			continue
		}
		fieldNames = append(fieldNames, backticks(field.ColumnName))
	}
	return fieldNames
}

func getSamples(conn *sql.DB, schema, table, field string, samples int64, dataType string) ([]interface{}, error) {
	var count int64
	var query string

	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", schema, table)
	if err := conn.QueryRow(queryCount).Scan(&count); err != nil {
		return nil, fmt.Errorf("cannot get count for table %q: %s", table, err)
	}

	if count < samples {
		query = fmt.Sprintf("SELECT `%s` FROM `%s`.`%s`", field, schema, table)
	} else {
		query = fmt.Sprintf("SELECT `%s` FROM `%s`.`%s` WHERE RAND() <= .3 LIMIT %d",
			field, schema, table, samples)
	}

	rows, err := conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("cannot get samples: %s, %s", query, err)
	}
	defer rows.Close()

	values := []interface{}{}

	for rows.Next() {
		var err error
		var val interface{}

		switch dataType {
		case "tinyint", "smallint", "mediumint", "int", "integer", "bigint", "year":
			var v int64
			err = rows.Scan(&v)
			val = v
		case "char", "varchar", "blob", "text", "mediumtext",
			"mediumblob", "longblob", "longtext":
			var v string
			err = rows.Scan(&v)
			val = v
		case "binary", "varbinary":
			var v []rune
			err = rows.Scan(&v)
			val = v
		case "float", "decimal", "double":
			var v float64
			err = rows.Scan(&v)
			val = v
		case "date", "time", "datetime", "timestamp":
			var v time.Time
			err = rows.Scan(&v)
			val = v
		}
		if err != nil {
			return nil, fmt.Errorf("cannot scan sample: %s", err)
		}
		values = append(values, val)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot get samples: %s", err)
	}
	return values, nil
}

func backticks(val string) string {
	if strings.HasPrefix(val, "`") && strings.HasSuffix(val, "`") {
		return url.QueryEscape(val)
	}
	return "`" + url.QueryEscape(val) + "`"
}

func isSupportedType(fieldType string) bool {
	supportedTypes := map[string]bool{
		"tinyint":    true,
		"smallint":   true,
		"mediumint":  true,
		"int":        true,
		"integer":    true,
		"bigint":     true,
		"float":      true,
		"decimal":    true,
		"double":     true,
		"char":       true,
		"varchar":    true,
		"date":       true,
		"datetime":   true,
		"timestamp":  true,
		"time":       true,
		"year":       true,
		"tinyblob":   true,
		"tinytext":   true,
		"blob":       true,
		"text":       true,
		"mediumblob": true,
		"mediumtext": true,
		"longblob":   true,
		"longtext":   true,
		"binary":     true,
		"varbinary":  true,
		"enum":       true,
		"set":        true,
	}
	_, ok := supportedTypes[fieldType]
	return ok
}

func processCliParams() (*cliOptions, error) {
	app := kingpin.New("mysql_random_data_loader", "MySQL Random Data Loader")

	opts := &cliOptions{
		app:        app,
		BulkSize:   app.Flag("bulk-size", "Number of rows per insert statement").Default("1000").Int(),
		ConfigFile: app.Flag("config-file", "MySQL config file").Default(defaultConfigFile).String(),
		Debug:      app.Flag("debug", "Log debugging information").Bool(),
		Factor:     app.Flag("fk-samples-factor", "Percentage used to get random samples for foreign keys fields").Default("0.3").Float64(),
		Host:       app.Flag("host", "Host name/IP").Short('h').Default("127.0.0.1").String(),
		MaxRetries: app.Flag("max-retries", "Number of rows to insert").Default("100").Int(),
		MaxThreads: app.Flag("max-threads", "Maximum number of threads to run inserts").Default("1").Int(),
		NoProgress: app.Flag("no-progress", "Show progress bar").Default("false").Bool(),
		Pass:       app.Flag("password", "Password").Short('p').String(),
		Port:       app.Flag("port", "Port").Short('P').Default("3306").Int(),
		Print:      app.Flag("print", "Print queries to the standard output instead of inserting them into the db").Bool(),
		Samples:    app.Flag("max-fk-samples", "Maximum number of samples for foreign keys fields").Default("100").Int64(),
		User:       app.Flag("user", "User").Short('u').String(),
		Version:    app.Flag("version", "Show version and exit").Bool(),

		Schema:    app.Arg("database", "Database").Required().String(),
		TableName: app.Arg("table", "Table").Required().String(),
		Rows:      app.Arg("rows", "Number of rows to insert").Required().Int(),
	}
	_, err := app.Parse(os.Args[1:])

	if err != nil {
		return nil, err
	}

	if mysqlOpts, err := readMySQLConfigFile(*opts.ConfigFile); err == nil {
		checkMySQLParams(opts, mysqlOpts)
	}

	return opts, nil
}

func checkMySQLParams(opts *cliOptions, mysqlOpts *mysqlOptions) {
	if *opts.Host == "" && mysqlOpts.Host != "" {
		*opts.Host = mysqlOpts.Host
	}

	if *opts.Port == 0 && mysqlOpts.Port != 0 {
		*opts.Port = mysqlOpts.Port
	}

	if *opts.User == "" && mysqlOpts.User != "" {
		*opts.User = mysqlOpts.User
	}

	if *opts.Pass == "" && mysqlOpts.Password != "" {
		*opts.Pass = mysqlOpts.Password
	}
}

func readMySQLConfigFile(filename string) (*mysqlOptions, error) {
	cfg, err := ini.Load(expandHomeDir(filename))
	if err != nil {
		return nil, err
	}

	section := cfg.Section(defaultMySQLConfigSection)
	port, _ := section.Key("port").Int()

	mysqlOpts := &mysqlOptions{
		Host:     section.Key("host").String(),
		Port:     port,
		User:     section.Key("user").String(),
		Password: section.Key("password").String(),
	}

	return mysqlOpts, nil
}

func expandHomeDir(dir string) string {
	if !strings.HasPrefix(dir, "~") {
		return dir
	}
	u, err := user.Current()
	if err != nil {
		return dir
	}
	return u.HomeDir + strings.TrimPrefix(dir, "~")
}
