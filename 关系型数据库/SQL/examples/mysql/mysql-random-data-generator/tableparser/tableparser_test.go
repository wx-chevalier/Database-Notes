package tableparser

import (
	"testing"
	"time"

	tu "github.com/Percona-Lab/mysql_random_data_load/testutils"
	_ "github.com/go-sql-driver/mysql"
	version "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

func TestParse(t *testing.T) {
	db := tu.GetMySQLConnection(t)
	v := tu.GetMinorVersion(t, db)
	var want *Table

	// Patch part of version is stripped by GetMinorVersion so for these test
	// it is .0
	sampleFiles := map[string]string{
		"5.6.0": "table001.json",
		"5.7.0": "table002.json",
		"8.0.0": "table003.json",
	}
	sampleFile, ok := sampleFiles[v.String()]
	if !ok {
		t.Fatalf("Unknown MySQL version %s", v.String())
	}

	table, err := NewTable(db, "sakila", "film")
	if err != nil {
		t.Error(err)
	}
	if tu.UpdateSamples() {
		tu.WriteJson(t, sampleFile, table)
	}
	tu.LoadJson(t, sampleFile, &want)

	tu.Equals(t, table, want)
}

func TestGetIndexes(t *testing.T) {
	db := tu.GetMySQLConnection(t)
	want := make(map[string]Index)
	tu.LoadJson(t, "indexes.json", &want)

	idx, err := getIndexes(db, "sakila", "film_actor")
	if tu.UpdateSamples() {
		tu.WriteJson(t, "indexes.json", idx)
	}
	tu.Ok(t, err)
	tu.Equals(t, idx, want)
}

func TestGetTriggers(t *testing.T) {
	db := tu.GetMySQLConnection(t)
	want := []Trigger{}
	v572, _ := version.NewVersion("5.7.2")
	v800, _ := version.NewVersion("8.0.0")

	sampleFile := "trigers-8.0.0.json"
	if tu.GetVersion(t, db).LessThan(v800) {
		sampleFile = "trigers-5.7.2.json"
	}
	if tu.GetVersion(t, db).LessThan(v572) {
		sampleFile = "trigers-5.7.1.json"
	}

	tu.LoadJson(t, sampleFile, &want)

	triggers, err := getTriggers(db, "sakila", "rental")
	// fake timestamp to make it constant/testeable
	triggers[0].Created.Time = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	if tu.UpdateSamples() {
		log.Info("Updating sample file: " + sampleFile)
		tu.WriteJson(t, sampleFile, triggers)
	}
	tu.Ok(t, err)
	tu.Equals(t, triggers, want)
}
