package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func saveFile(pathTpl *template.Template, bodyTpl *template.Template, format string, m map[string]interface{}) {

	var path bytes.Buffer

	err := pathTpl.Execute(&path, m)
	check(err)

	log.Debugf("Write %s file: %s\n", format, path.String())

	f, err := os.Create(path.String())
	check(err)
	defer f.Close()

	switch format {
	case "template":
		err = bodyTpl.Execute(f, m)
	case "json":
		data, err := json.MarshalIndent(m, "", "    ")
		check(err)
		f.Write(data)
	case "md":
		fmt.Fprintf(f, "---\n")
		data, err := yaml.Marshal(m)
		check(err)
		fmt.Fprintf(f, "%s\n", data)
		fmt.Fprintf(f, "---\n")
	}

	check(err)
}

func getEnvDef(key, def string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return def
	}
	return value
}

func getDefaultPort(dbtype string, port int) int {

	if port != 0 {
		return port
	}

	defPorts := map[string]int{
		"mysql":    3306,
		"postgres": 5432,
	}

	return defPorts[dbtype]

}

var Usage = func() {

	version := "0.0.8"

	fmt.Fprintf(flag.CommandLine.Output(), "sql-export ( https://github.com/yaroslaff/sql-export ) version %s\nUsage:\n", version)
	flag.PrintDefaults()
}

func replace(input, from, to string) string {

	return strings.Replace(input, from, to, -1)
}

func read_any_json(path string, data *[]map[string]interface{}) {
	json_bytes, err := os.ReadFile(path)
	check(err)
	json.Unmarshal(json_bytes, data)
}

func read_from_db(dbtype, user, password, host, dbname, q string, port int) []map[string]interface{} {

	/* prepare port and conn_string */
	var conn_string string
	mlist := make([]map[string]interface{}, 0)

	port = getDefaultPort(dbtype, port)

	switch dbtype {
	case "mysql":
		conn_string = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbname)
	case "postgres":
		conn_string = fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", user, password, host, port, dbname)
	case "sqlite3":
		conn_string = dbname
	}

	// Open up our database connection.
	// I've set up a database on my local machine using phpmyadmin.
	// The database is called testDb
	log.Debug("connect to " + conn_string)
	db, err := sql.Open(dbtype, conn_string)
	check(err)
	log.Debugf("Connected to %s database %s at %s:%d !", dbtype, dbname, host, port)

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	if !strings.Contains(q, " ") {
		q = fmt.Sprintf("SELECT * FROM %s", q)
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

	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {

		m := make(map[string]interface{})

		// Scan the result into the column pointers...
		if err := rows.Scan(dest...); err != nil {
			return nil
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
			case "INT", "SMALLINT", "INTEGER", "INT4":
				m[name], _ = strconv.Atoi(v)
			case "CHAR", "VARCHAR", "DATETIME", "DATE", "TEXT", "TIMESTAMP":
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
		mlist = append(mlist, m)
	}
	return mlist
}

func main() {

	//def_server := "username:password@tcp(127.0.0.1:3306)/test"
	def_dbhost := getEnvDef("DBHOST", "localhost")
	def_user, err := user.Current()
	check(err)
	def_dbuser := getEnvDef("DBUSER", def_user.Username)
	def_dbpass := getEnvDef("DBPASS", "")

	def_port, _ := strconv.Atoi(getEnvDef("DBPORT", "0"))
	def_dbname := getEnvDef("DBNAME", "")
	def_dbtype := "mysql"

	var (
		host     string
		port     int
		user     string
		password string
		dbname   string
		q        string
		verbose  bool
		output   string
		format   string
		tpl      string
		dbtype   string
	)

	var pathTpl, contentTpl *template.Template

	mlist := make([]map[string]interface{}, 0)

	flag.StringVar(&host, "h", def_dbhost, "$DBHOST")
	flag.IntVar(&port, "port", def_port, "$DBPORT")
	flag.StringVar(&user, "u", def_dbuser, "$DBUSER")
	flag.StringVar(&password, "p", def_dbpass, "$DBPASS")
	flag.StringVar(&dbtype, "d", def_dbtype, "$DBTYPE")
	flag.StringVar(&dbname, "n", def_dbname, "$DBNAME")
	flag.StringVar(&q, "q", "", "SQL query or table name or filename")
	flag.StringVar(&output, "o", "", "Output filename")
	flag.StringVar(&format, "f", "template", "Format: json or md (markdown with frontmatter) or template")
	flag.StringVar(&tpl, "tpl", "", "Template input file")
	flag.BoolVar(&verbose, "v", false, "verbose mode")

	flag.Usage = Usage

	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	if q == "" {
		log.Error("Provide SQL query (e.g -q \"SELECT id, title FROM tablename\") or just -q tablename")
		os.Exit(1)
	}

	/* prepare templates */
	if output != "" {

		funcMap := template.FuncMap{
			"replace": replace,
		}

		pathTpl = template.Must(template.New("outfile").Funcs(funcMap).Parse(output))
		log.Debugf("Content: %#v", contentTpl)

		if format != "template" && format != "json" && format != "md" {
			panic("Format must be one of template/json/md")
		}

		if format == "template" {
			if tpl == "" {
				panic("When format (-f) is template, --tpl is required")
			}
			contentTpl = template.Must(template.New(path.Base(tpl)).Funcs(funcMap).ParseFiles(tpl))
		}

	}

	if strings.HasSuffix(q, ".json") {
		if _, err := os.Stat(q); err == nil {
			read_any_json(q, &mlist)
			log.Debugf("Read %d items", len(mlist))
		}
	} else {
		/* read from database */
		mlist = read_from_db(dbtype, user, password, host, dbname, q, port)
	}

	if pathTpl != nil {
		for _, mrun := range mlist {
			saveFile(pathTpl, contentTpl, format, mrun)
		}
	}

	if output == "" {
		j, err := json.MarshalIndent(mlist, "", "    ")
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			return
		}
		fmt.Printf("%s\n", j)
	}
}
