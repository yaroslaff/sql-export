# SQL Export
Export SQL tables or queries to file(s) in JSON/Markdown format.

## Usage

### export SQL query to one JSON list or file
~~~
$ ./sql-export -u dbuser -p dbpass -d dbname -q 'SELECT title, price FROM libro LIMIT 2'  
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
$ ./sql-export -u dbuser -p dbpass -d dbname -q 'SELECT title, price FROM libro LIMIT 2' > /tmp/books.json
~~~

### export SQL to many (one file per record) JSON files 
Use `-f json` and provide template to output filename `-o '/tmp/libro/{{.id}}.json'`.

~~~
$ ./sql-export -u dbuser -p dbpass -d dbname -q 'SELECT id, title, price FROM libro' -f json -o '/tmp/libro/{{.id}}.json'
$ cat /tmp/libro/123.json 
{
    "id": 123,
    "price": 45,
    "title": "ALLGEMEINE GESCHICHTE DER HANDFEUERWAFFEN. Eine Übersicht ihrer Entwickelung. Mit 123 Abbildungen und 4 Ubersichtstafeln. - Günther Reinhold. - Reprint Verlag, - 2001"
}
~~~

### export SQL to many markdown files with YAML frontmatter 
Use `-f md` and provide template to output filename `-o '/tmp/libro/{{.id}}.json'`.

~~~
$ ./sql-export -u dbuser -p dbpass -d dbname -q 'SELECT id, title, price FROM libro' -f md -o '/tmp/libro/{{.id}}.md'
$ cat /tmp/libro/123.md 
---
id: 123
price: 45
title: ALLGEMEINE GESCHICHTE DER HANDFEUERWAFFEN. Eine Übersicht ihrer Entwickelung.
    Mit 123 Abbildungen und 4 Ubersichtstafeln. - Günther Reinhold. - Reprint Verlag,
    - 2001

---
~~~

### export SQL to many files with any custom template 
Provide output filename template (`-o`) and template (`--tpl`)

~~~
$ cat out-template.html 
id: {{.id}}
title: {{.title}}
price: {{.price}}
$ ./sql-export -u dbuser -p dbpass -d dbname -q 'SELECT id, title, price FROM libro' -o '/tmp/libro/{{.id}}.txt' --tpl out-template.html 
xenon@mir:~/repo/sql-export$ cat /tmp/libro/123.txt 
id: 123
title: ALLGEMEINE GESCHICHTE DER HANDFEUERWAFFEN. Eine Übersicht ihrer Entwickelung. Mit 123 Abbildungen und 4 Ubersichtstafeln. - Günther Reinhold. - Reprint Verlag, - 2001
price: 45
~~~

## Install/build
~~~
git clone https://github.com/yaroslaff/sql-export
cd sql-export
go build
cp sql-export /usr/local/bin
~~~

or download from https://github.com/yaroslaff/sql-export from **Releases** (if your arch is x86_64).
