package testutils

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	version "github.com/hashicorp/go-version"
)

var (
	basedir string
	db      *sql.DB
)

func BaseDir() string {
	if basedir != "" {
		return basedir
	}
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}

	basedir = strings.TrimSpace(string(out))
	return basedir
}

func GetMySQLConnection(tb testing.TB) *sql.DB {
	if db != nil {
		return db
	}

	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		fmt.Printf("%s TEST_DSN environment variable is empty", caller())
		tb.FailNow()
	}

	// Parse the DSN in the env var and ensure it has parseTime & multiStatements enabled
	// MultiStatements is required for LoadQueriesFromFile
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		fmt.Printf("%s cannot parse DSN %q: %s", caller(), dsn, err)
		tb.FailNow()
	}
	if cfg.Params == nil {
		cfg.Params = make(map[string]string)
	}
	cfg.ParseTime = true
	cfg.MultiStatements = true
	cfg.InterpolateParams = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		fmt.Printf("%s cannot connect to the db: DSN: %s\n%s", caller(), dsn, err)
		tb.FailNow()
	}
	return db
}

func UpdateSamples() bool {
	us, _ := strconv.ParseBool(os.Getenv("UPDATE_SAMPLES"))
	return us
}

func GetVersion(tb testing.TB, db *sql.DB) *version.Version {
	var vs string
	err := db.QueryRow("SELECT VERSION()").Scan(&vs)
	if err != nil {
		fmt.Printf("%s cannot get MySQL version: %s\n\n", caller(), err)
		tb.FailNow()
	}
	v, err := version.NewVersion(vs)
	if err != nil {
		fmt.Printf("%s cannot get MySQL version: %s\n\n", caller(), err)
		tb.FailNow()
	}
	return v
}

func GetMinorVersion(tb testing.TB, db *sql.DB) *version.Version {
	var vs string
	err := db.QueryRow("SELECT VERSION()").Scan(&vs)
	if err != nil {
		fmt.Printf("%s cannot get MySQL version: %s\n\n", caller(), err)
		tb.FailNow()
	}

	// Extract only major and minor version
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+).*`)
	m := re.FindAllStringSubmatch(vs, -1)
	vs = fmt.Sprintf("%s.%s", m[0][1], m[0][2])

	v, err := version.NewVersion(vs)
	if err != nil {
		fmt.Printf("%s cannot get MySQL version: %s\n\n", caller(), err)
		tb.FailNow()
	}
	return v
}

func LoadFile(tb testing.TB, filename string) []string {
	file := filepath.Join("testdata", filename)
	fh, err := os.Open(file)
	lines := []string{}
	reader := bufio.NewReader(fh)

	line, err := reader.ReadString('\n')
	for err == nil {
		lines = append(lines, strings.TrimRight(line, "\n"))
		line, err = reader.ReadString('\n')
	}
	return lines
}

func UpdateSampleFile(tb testing.TB, filename string, lines []string) {
	if us, _ := strconv.ParseBool(os.Getenv("UPDATE_SAMPLES")); !us {
		return
	}
	WriteFile(tb, filename, lines)
}

func UpdateSampleJSON(tb testing.TB, filename string, data interface{}) {
	if us, _ := strconv.ParseBool(os.Getenv("UPDATE_SAMPLES")); !us {
		return
	}
	WriteJson(tb, filename, data)
}

func WriteFile(tb testing.TB, filename string, lines []string) {
	file := filepath.Join("testdata", filename)
	ofh, err := os.Create(file)
	if err != nil {
		fmt.Printf("%s cannot load json file %q: %s\n\n", caller(), file, err)
		tb.FailNow()
	}
	for _, line := range lines {
		ofh.WriteString(line + "\n")
	}
	ofh.Close()
}

func LoadQueriesFromFile(tb testing.TB, filename string) {
	conn := GetMySQLConnection(tb)
	file := filepath.Join("testdata", filename)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("%s cannot load json file %q: %s\n\n", caller(), file, err)
		tb.FailNow()
	}
	_, err = conn.Exec(string(data))
	if err != nil {
		fmt.Printf("%s cannot load queries from %q: %s\n\n", caller(), file, err)
		tb.FailNow()
	}
}

func LoadJson(tb testing.TB, filename string, dest interface{}) {
	file := filepath.Join("testdata", filename)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("%s cannot load json file %q: %s\n\n", caller(), file, err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		fmt.Printf("%s cannot unmarshal the contents of %q into %T: %s\n\n", caller(), file, dest, err)
	}
}

func WriteJson(tb testing.TB, filename string, data interface{}) {
	file := filepath.Join("testdata", filename)
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("%s cannot marshal %T into %q: %s\n\n", caller(), data, file, err)
		tb.FailNow()
	}
	err = ioutil.WriteFile(file, buf, os.ModePerm)
	if err != nil {
		fmt.Printf("%s cannot write file %q: %s\n\n", caller(), file, err)
		tb.FailNow()
	}
}

// assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		fmt.Printf("%s "+msg+"\n\n", append([]interface{}{caller()}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf("%s: unexpected error: %s\n\n", caller(), err.Error())
		if len(args) > 0 {
			msg = fmt.Sprintf("%s: %s "+args[0].(string), append([]interface{}{caller(), err}, args[1:]...)) + "\n\n"
		}
		fmt.Println(msg)
		tb.FailNow()
	}
}

func NotOk(tb testing.TB, err error) {
	if err == nil {
		fmt.Printf("%s: expected error is nil\n\n", caller())
		tb.FailNow()
	}
}

func IsNil(tb testing.TB, i interface{}, args ...interface{}) {
	if i != nil {
		msg := fmt.Sprintf("%s: expected nil, got %#v\n\n", caller(), i)
		if len(args) > 0 {
			msg = fmt.Sprintf("%s: %s "+args[0].(string), append([]interface{}{caller(), i}, args[1:]...)) + "\n"
		}
		fmt.Println(msg)
		tb.FailNow()
	}
}

func NotNil(tb testing.TB, i interface{}) {
	if i == nil {
		fmt.Printf("%s: expected value, got nil\n", caller())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		fmt.Printf("%s\n\n\texp: %#v\n\n\tgot: %#v\n\n", caller(), exp, act)
		tb.FailNow()
	}
}

// Get the caller's function name and line to show a better error message
func caller() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
