# plgo
plgo is an "library" for easily creating PostgreSQL stored procedures extension in golang.

For now not all type conversions (from PostgreSQL to Go types) are supported.

Creating new stored procedures in plgo is easy:

1. Copy the `pl.go` to your new extension directory

2. Edit the section in pl.go marked as: //PG_FUNCTION declarations
    Add all your new procedure names
    eg.
    ```c
    PG_FUNCTION_INFO_V1(my_awesome_procedure)
    ```

3. Create a file where your procedures will be declared
    eg. my_procedures.go:
    ```go
    package main

    /*
    #include "postgres.h"
    #include "fmgr.h"
    */
    import "C"

    //all stored procedures must be of type func(fcinfo *FuncInfo) Datum
    //before the procedure must be an comment: //export procedure_name

    //export my_awesome_procedure
    func my_awesome_procedure(fcinfo *FuncInfo) Datum {
	    //getting the function parameters
	    t := fcinfo.Text(0)
	    x := fcinfo.Int(1)

	    //preparing query statement
	    plan, err := Prepare("select * from test where id=$1", []string{"integer"})
	    if err != nil {
		    return ToDatum(err)
	    }
	    defer plan.Close()

	    //running statement
	    row, err := plan.QueryRow(1)
	    if err != nil {
		    return ToDatum(err)
	    }

	    //scanning result row
	    var id int
	    var txt string
	    err = row.Scan(&id, &txt)
	    if err != nil {
		    return ToDatum(err)
	    }

	    //some magic with return value :)
	    var ret string
	    for i := 0; i < x; i++ {
		    ret += t + txt
	    }
        //return type must be converted to Datum
	    return ToDatum(ret)
    }
    ```

4. Build an shared object file with `$ go build -v -buildmode=c-shared -o my_procedures.so my_procedures.go pl.go`

5. Copy the shared object file to the PostgreSQL libs directory `sudo cp my_procedures.so $(pg_config --pkglibdir)`

6. Create the procedure in PostgreSQL
    ```sql
    CREATE OR REPLACE FUNCTION public.my_awesome_procedure(text, integer)
      RETURNS text AS
    '$libdir/my_procedures', 'my_awesome_procedure'
      LANGUAGE c IMMUTABLE STRICT;
    ```

7. Happily run the function in your queries `select my_awesome_procedure('foo', 10)`
    output:
    ```
                             plgo_example                         
    --------------------------------------------------------------
     foomehfoomehfoomehfoomehfoomehfoomehfoomehfoomehfoomehfoomeh
    (1 row)
    ```
