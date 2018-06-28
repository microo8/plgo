[![Report card](http://goreportcard.com/badge/microo8/plgo)](http://goreportcard.com/report/microo8/plgo)
[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.png?v=103)](https://opensource.org/licenses/mit-license.php)

# plgo
plgo is an tool for easily creating PostgreSQL extensions with stored procedures and triggers in golang. It creates wrapper code, PostgreSQL extension files and builds your package.

contribution of all kind welcome!


## installation

`go get -u github.com/microo8/plgo/plgo`

## write functions

Creating new stored procedures with plgo is easy:

Create a package where your procedures will be declared:

```go
//must be main package

package main

import (
	"log"
	"strings"

	"github.com/microo8/plgo"
)

//from every exported function will be generated a stored procedure
//functions can take (and return) any golang builtin type (like string, int, float64, []int, ...)

func Meh() {
    //NoticeLogger for printing notice messages to elog
    logger := plgo.NewNoticeLogger("", log.Ltime|log.Lshortfile)
    logger.Println("meh")
}

//ConcatAll concatenates all values of an column in a given table
func ConcatAll(tableName, colName string) string {
    //ErrorLogger for printing error messages to elog
    logger := plgo.NewErrorLogger("", log.Ltime|log.Lshortfile)
    db, err := plgo.Open() //open the connection to DB
    if err != nil {
        logger.Fatalf("Cannot open DB: %s", err)
    }
    defer db.Close() //db must be closed
    query := "select " + colName + " from " + tableName
    stmt, err := db.Prepare(query, nil) //prepare an statement
    if err != nil {
        logger.Fatalf("Cannot prepare query statement (%s): %s", query, err)
    }
    rows, err := stmt.Query() //execute statement
    if err != nil {
        logger.Fatalf("Query (%s) error: %s", query, err)
    }
    var ret string
    for rows.Next() { //iterate over the rows
        var val string
        rows.Scan(&val)
        ret += val
    }
    return ret
}

//CreatedTimeTrigger is an trigger function
//trigger function must have the first argument of type *plgo.TriggerData
//and must return *plgo.TriggerRow
func CreatedTimeTrigger(td *plgo.TriggerData) *plgo.TriggerRow {
    td.NewRow.Set(4, time.Now()) //set the 4th column to now()
    return td.NewRow //return the new modified row
}

//ConcatArray concatenates an array of strings
//function arguments (and return values) can be also array types of the golang builtin types
func ConcatArray(strs []string) string {
    return strings.Join(strs, "")
}
```

## create extension

build the PostgreSQL extension with `$ plgo [path/to/package]`

this will create an directory named `build`, where the compiled shared object will be and also all files needed for the extension installation (like `Makefile`, `extention.sql`, ...)

## install extension

go to the `build` directory and install your new extension:

```bash
$ cd build
$ sudo make install
```

this installs your extension to DB. You can then use this extension in db:

```sql
CREATE EXTENSION myextention;
```

Finally you can happily run your the functions in your queries `select concatarray(array['foo','bar'])`

output:

```
 concatarray
-------------
 foobar
(1 row)
```

### use of goroutines

Using goroutines is possible, but very tricky. The allocation of the stack for the goroutine is bigger than [max_stack_depth](https://www.postgresql.org/docs/current/static/runtime-config-resource.html). Running an procedure that spins-up some goroutines ends with crashing:

```
ERROR:  stack depth limit exceeded
HINT:  Increase the configuration parameter "max_stack_depth" (currently 7680kB), after ensuring the platform's stack depth limit is adequate.
```

Setting it to the kernel maximum (`ulimit -s`) doesn't help.
But the size of allocated stack is checked by the DB only when calling some statement. So you can probably play with that. You get all the data from the DB at the beginning of your procedure and then spin-up some goroutines, after that don't touch the DB. But I don't recommend doing it.

## todo

- Own type definition!
- Background Worker Processes!
- Functions returning `SETOF`
