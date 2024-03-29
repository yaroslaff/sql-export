# SQL Export
Export SQL tables or queries to file(s) in JSON/Markdown format. Mainly to use with static site generators like [Hugo](https://gohugo.io/).

`sql-export` can generate thousands .md files with YAML frontmatter based on content from mysql database in few seconds.

MySQL/MariaDB (default), PostgreSQL and SQLite3 databases are supported. For sqlite3 use filename as DBNAME.

## Usage

### Export SQL query to one JSON list or file
For brevity, we omit [database credentials](#database-credentials) in examples, assume it comes from environment or .env file

~~~
$ ./sql-export -q 'SELECT title, price FROM libro LIMIT 2'  
[
    {
        "price": 170,
        "title": "GLI ARMAROLI MILANESI - I MISSAGLIA E LA LORO CASA. Notizie, documenti, ricordi. - Gelli J., Moretti G. - Hoepli, - 1903"
    },
    {
        "price": 38,
        "title": "LES FUSILS D\u0026#039;INFANTERIE EUROPEENS A LA FIN DU XIX SIECLE. - Sor Daniel. - Crepin-Leblond, - 1972"
    }
]
~~~

Surely, you can redirect to file:
~~~
$ ./sql-export -q 'SELECT title, price FROM libro LIMIT 2' > /tmp/books.json
~~~

### Export SQL to many (one file per record) JSON files 
Use `-f json` and provide template to output filename `-o '/tmp/libro/{{.id}}.json'`.

~~~
$ ./sql-export -q 'SELECT id, title, price FROM libro' -f json -o '/tmp/libro/{{.id}}.json'
$ cat /tmp/libro/123.json 
{
    "id": 123,
    "price": 45,
    "title": "ALLGEMEINE GESCHICHTE DER HANDFEUERWAFFEN. Eine Übersicht ihrer Entwickelung. Mit 123 Abbildungen und 4 Ubersichtstafeln. - Günther Reinhold. - Reprint Verlag, - 2001"
}
~~~

### Export SQL to many markdown files with YAML frontmatter 
Use `-f md` and provide template to output filename `-o '/tmp/libro/{{.id}}.md'`.

~~~
$ ./sql-export -q 'SELECT id, title, price FROM libro' -f md -o '/tmp/libro/{{.id}}.md'
$ cat /tmp/libro/123.md 
---
id: 123
price: 45
title: ALLGEMEINE GESCHICHTE DER HANDFEUERWAFFEN. Eine Übersicht ihrer Entwickelung.
    Mit 123 Abbildungen und 4 Ubersichtstafeln. - Günther Reinhold. - Reprint Verlag,
    - 2001

---
~~~

### Export SQL to many files with any custom template 
Provide output filename template (`-o`) and template (`--tpl`)

~~~
$ cat out-template.html 
id: {{.id}}
title: {{.title}}
price: {{.price}}
$ ./sql-export -q 'SELECT id, title, price FROM libro' -o '/tmp/libro/{{.id}}.txt' --tpl out-template.html 
$ cat /tmp/libro/123.txt 
id: 123
title: ALLGEMEINE GESCHICHTE DER HANDFEUERWAFFEN. Eine Übersicht ihrer Entwickelung. Mit 123 Abbildungen und 4 Ubersichtstafeln. - Günther Reinhold. - Reprint Verlag, - 2001
price: 45
~~~

### Use JSON as input source (generate markdown files based on json)

To read dataset from JSON, just pass filename as `-q` parameter.

~~~
./sql-export -q libro.json -o '/tmp/libro/{{.id}}.md' -f md
~~~


## Options
~~~
$ ./sql-export --help
Usage of ./sql-export:
  -d string
    	$DBTYPE (default "mysql")
  -f string
    	Format: json or md (markdown with frontmatter) or template (default "template")
  -h string
    	$DBHOST (default "localhost")
  -n string
    	$DBNAME (default "XXXXXX")
  -o string
    	Output filename
  -p string
    	$DBPASS (default "XXXXXX")
  -port int
    	$DBPORT (default 3306)
  -q string
    	SQL query or table name
  -tpl string
    	Template input file
  -u string
    	$DBUSER (default "XXXXXX")
  -v	verbose mode
~~~

You may use table name as value to `-q`. `-q tableName` equals to `-q SELECT * FROM tableName`

### Database credentials

|value         |key      |environment |default value|
|---           |---      |---         |---|
|database type | `-d`    |`$DBTYPE`   | `mysql`|
|database host | `-h`    |`$DBHOST`   | `localhost`|
|database port | `-port` |`$DBPORT`   | 3306 or 5432 |
|database user | `-u`    |`$DBUSER`   | current system user name |
|database name | `-p`    |`$DBPASS`   |   |

You can use `.env` file:
~~~
DBUSER=xenon
DBPASS=YouWillNotSeeMyRealPasswordHere
DBHOST=localhost
DBNAME=books
~~~

## Install/build

Three different methods.

### Releases
download precompiled binary from https://github.com/yaroslaff/sql-export from [latest release](https://github.com/yaroslaff/sql-export/releases/latest) (if your arch is x86_64).

### go install
If you have modern golang installed:
~~~
go install github.com/yaroslaff/sql-export@latest
~~~

Installed (default) to ~/go/bin/sql-export


### clone and build from sources
~~~
git clone https://github.com/yaroslaff/sql-export
cd sql-export
go build
cp sql-export /usr/local/bin
~~~

### Post-install test
Optional quick post-install test with sqlite3 db from sample [Chinook database](https://github.com/lerocha/chinook-database):

~~~
wget https://github.com/lerocha/chinook-database/raw/master/ChinookDatabase/DataSources/Chinook_Sqlite.sqlite
sql-export -n Chinook_Sqlite.sqlite -d sqlite3 -q Customer
~~~

## Benchmarking
sql-export is written in Go, so it's very fast. I test on database with 57000+ records.

|Test                                                            |time     |
|---                                                             |---      |
| Export 3 fields of 57k+ records to one (11Mb) json list        | 0.336s  |
| Export all (40) fields of 57k+ records to one (92Mb) json list | 3.102s  |
| Export 3 fields to 57k+ JSON files                             | 5.573s  |
| Export 3 fields to 57k+ md/yaml files                          | 10.869s |
| Export 3 fields to 57k+ custom template files                  | 4.321s  |



