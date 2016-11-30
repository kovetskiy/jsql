package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kovetskiy/godocs"
	"github.com/reconquest/ser-go"
	"github.com/seletskiy/hierr"

	_ "github.com/mattn/go-sqlite3"
)

var (
	version = "[manual build]"
	usage   = "jsql " + version + `

Execute SQL query into JSON dataset.

Usage:
    jsql [options] <query>
    jsql -h | --help
    jsql --version

Options:
    -f --file <path>  Specify file with JSON dataset. [default: /dev/stdin]
    -h --help         Show this screen.
    --version         Show version.
`
)

func main() {
	args := godocs.MustParse(usage, version, godocs.UsePager)

	file, err := os.Open(args["--file"].(string))
	if err != nil {
		hierr.Fatalf(
			err, "can't open file: %s", args["--file"].(string),
		)
	}

	var value interface{}

	err = json.NewDecoder(file).Decode(&value)
	if err != nil {
		hierr.Fatalf(
			err, "invalid input data",
		)
	}

	records := []map[string]interface{}{}

	switch value := value.(type) {
	case map[string]interface{}:
		records = []map[string]interface{}{value}

	case []interface{}:
		for _, subvalue := range value {
			if subvalue, ok := subvalue.(map[string]interface{}); ok {
				records = append(records, subvalue)
			} else {
				hierr.Fatalf(
					errors.New("must be object or array of objects"),
					"invalid input data",
				)
			}
		}

	default:
		hierr.Fatalf(
			errors.New("must be object or array of objects"),
			"invalid input data",
		)
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		hierr.Fatalf(
			err, "can't allocate sqlite3 in-memory database",
		)
	}

	err = push(db, records)
	if err != nil {
		hierr.Fatalf(
			err, "can't push dataset into in-memory database",
		)
	}

	rows, err := db.Query(args["<query>"].(string))
	if err != nil {
		hierr.Fatalf(
			err, "query failed: %s", args["<query>"].(string),
		)
	}

	records = []map[string]interface{}{}
	for rows.Next() {
		record := map[string]interface{}{}

		columns, err := rows.Columns()
		if err != nil {
			hierr.Fatalf(
				err, "unable to obtain rows columns",
			)
		}

		pointers := []interface{}{}
		for _, column := range columns {
			var value interface{}
			pointers = append(pointers, &value)
			record[column] = &value
		}

		err = rows.Scan(pointers...)
		if err != nil {
			hierr.Fatalf(err, "can't read result records")
		}

		for key, value := range record {
			record[key] = *value.(*interface{})
		}

		records = append(records, record)
	}

	jsoned, err := json.MarshalIndent(records, "", "     ")
	if err != nil {
		hierr.Fatalf(err, "can't encode output data into JSON")
	}

	fmt.Println(string(jsoned))
}

func push(db *sql.DB, records []map[string]interface{}) error {
	hashKeys := map[string]struct{}{}

	for _, record := range records {
		for key, _ := range record {
			hashKeys[key] = struct{}{}
		}
	}

	keys := []string{}

	for key, _ := range hashKeys {
		keys = append(keys, key)
	}

	query := "CREATE TABLE data (" + strings.Join(keys, ",") + ")"

	_, err := db.Exec(query)
	if err != nil {
		return ser.Errorf(
			err, "can't create table",
		)
	}

	for _, record := range records {
		recordKeys := []string{}
		recordValues := []string{}
		recordArgs := []interface{}{}

		for key, value := range record {
			recordKeys = append(recordKeys, key)
			recordValues = append(recordValues, "?")
			recordArgs = append(recordArgs, value)
		}

		query := "INSERT INTO data (" + strings.Join(recordKeys, ",") +
			") VALUES (" + strings.Join(recordValues, ", ") + ")"

		statement, err := db.Prepare(query)
		if err != nil {
			return ser.Errorf(
				err, "can't prepare query: %s", query,
			)
		}

		_, err = statement.Exec(recordArgs...)
		if err != nil {
			return ser.Errorf(
				err, "can't insert record",
			)
		}
	}

	return nil
}
