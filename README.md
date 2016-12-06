# jsql

jsql allows to make SQL query into JSON dataset.

## How it works

* read raw JSON dataset intput
* decode JSON into internal hash
* create table within in-memory SQLite database
* insert decoded items
* run user-provided SQL query
* read resulting database rows
* encode rows into JSON array of objects and print to stdout

## Usage

Pass JSON dataset as stdin or use `-f --file <path>` flag and specify SQL
query to run.

```

$ echo '[{"a":1,"f":3.14},{"a":2,"b":"3"}]' | jsql 'SELECT * FROM data'
[
     {
          "a": 1,
          "b": null,
          "f": 3.14
     },
     {
          "a": 2,
          "b": "3",
          "f": null
     }
]
```

## Installation

Arch Linux User Repository:

[https://aur.archlinux.org/packages/jsql-git/](https://aur.archlinux.org/packages/jsql-git/)

or manually:

```
go get github.com/kovetskiy/jsql
```

## License
MIT.
