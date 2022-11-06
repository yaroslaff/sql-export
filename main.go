package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

func toJsonMap(data []byte) (interface{}, error) {
	var newMap interface{}

	err := json.Unmarshal(data, &newMap)

	if err != nil {
		return nil, err
	}

	return newMap, nil
}

func main() {

	//def_server := "username:password@tcp(127.0.0.1:3306)/test"
	def_host := "localhost"
	def_user, err := user.Current()

	println(reflect.TypeOf(&def_user))

	def_port := 3306

	var (
		host     string
		port     int
		user     string
		password string
		dbname   string
		q        string
		verbose  bool
	)

	flag.StringVar(&host, "s", def_host, "server, def: "+def_host)
	flag.IntVar(&port, "P", def_port, fmt.Sprintf("port, def: %d", def_port))
	flag.StringVar(&user, "u", def_user.Username, "user, def: "+def_user.Username)
	flag.StringVar(&password, "p", "", "password")
	flag.StringVar(&dbname, "d", "", "database name")
	flag.StringVar(&q, "q", "", "SQL query or table name")
	flag.BoolVar(&verbose, "v", false, "verbose mode")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [filename]\n"+
				"  if no filename, then stdin", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("Debug")

	server := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbname)

	// Open up our database connection.
	// I've set up a database on my local machine using phpmyadmin.
	// The database is called testDb
	db, err := sql.Open("mysql", server)

	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	} else {
		log.Debug("Connected!")
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	// fix query

	if !strings.Contains(q, " ") {
		q = fmt.Sprintf("SELECT * FROM %s", q)
		log.Debug("Use query: ", q)
	}

	log.Debug("Run query: ", q)
	rows, err := db.Query(q)
	if err != nil {
		panic(err.Error())
	}

	cols, _ := rows.Columns()
	colTypes, _ := rows.ColumnTypes()
	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	mlist := make([]map[string]interface{}, 0)

	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {

		m := make(map[string]interface{})

		// Scan the result into the column pointers...
		if err := rows.Scan(dest...); err != nil {
			return
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		//m := make(map[string]interface{})
		for i, ct := range colTypes {
			name := ct.Name()

			if rawResult[i] == nil {
				m[name] = nil
				continue
			}

			v := string(rawResult[i])

			switch t := ct.DatabaseTypeName(); t {
			case "INT", "SMALLINT":
				m[name], _ = strconv.Atoi(v)
			case "CHAR", "VARCHAR", "DATETIME", "DATE", "TEXT":
				m[name] = v
			case "DECIMAL":
				m[name], err = strconv.ParseFloat(v, 64)
				if err != nil {
					fmt.Printf("error: %s (%s)", err, v)
					// return
				}
			default:
				fmt.Println("Do not know how to convert" + t)
			}

			//val := dest[i].(*interface{})
			//m[colName] = *val
		}

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		mlist = append(mlist, m)
	}

	j, err := json.MarshalIndent(mlist, "", "    ")
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		return
	}

	fmt.Printf("%s", j)
}
