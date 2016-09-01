# plgo
plgo is an "library" for easily creating PostgreSQL stored procedures and triggers in golang.

contribution of all kind welcome!

Creating new stored procedures with plgo is easy:

1. Copy the `pl.go` to your new extension directory (optionally edit the CFLAGS path if it is different, use `pg_config --includedir-server`)

2. Create `funcdec.h` header file

    Add all your new procedure names
    eg.
    ```c
    PG_FUNCTION_INFO_V1(my_awesome_procedure);
    PG_FUNCTION_INFO_V1(another_awesome_procedure);
    ```

3. Create a file where your procedures will be declared
    eg. my_procedures.go:
    ```go
    package main

    //this C block is required

    /*
    #include "postgres.h"
    #include "fmgr.h"
    */
    import "C"
    import "log"

    //all stored procedures must be of type func(*FuncInfo) Datum
    //before the procedure must be an comment: //export procedure_name

    //export my_awesome_procedure
    func my_awesome_procedure(fcinfo *FuncInfo) Datum {
	    //getting the function parameters
        var t string
        var x int
        fcinfo.Scan(&t, &x)

	    //Creating notice logger
	    logger := log.New(&elog{}, "", log.Ltime|log.Lshortfile)

    	//connect to DB
    	db, err := Open()
    	if err != nil {
    		logger.Fatal(err)
    	}
    	defer db.Close()

	    //preparing query statement
	    stmt, err := db.Prepare("select * from test where id=$1", []string{"integer"})
	    if err != nil {
    		logger.Fatal(err)
	    }

	    //running statement
	    row, err := stmt.QueryRow(1)
	    if err != nil {
    		logger.Fatal(err)
	    }

	    //scanning result row
	    var id int
	    var txt string
	    err = row.Scan(&id, &txt)
	    if err != nil {
    		logger.Fatal(err)
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

4. Build an shared objects file with `$ go build -v -buildmode=c-shared -o my_procedures.so my_procedures.go pl.go`

5. Copy the shared objects file to PostgreSQL libdir `$ sudo cp my_procedures.so $(pg_config --pkglibdir)`

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
                         my_awesome_procedure                         
    --------------------------------------------------------------
     foomehfoomehfoomehfoomehfoomehfoomehfoomehfoomehfoomehfoomeh
    (1 row)
    ```

##Triggers

Triggers are also easy:

```go
//export plgo_trigger
func plgo_trigger(fcinfo *FuncInfo) Datum {
    //logger
	t := log.New(&ELog{level: NOTICE}, "", log.Lshortfile|log.Ltime)

    //this must be true, else the function is not called as a trigger
	if !fcinfo.CalledAsTrigger() {
		t.Fatal("Not called as trigger")
	}

    //use TriggerData to manipulate the Old and New row
	triggerData := fcinfo.TriggerData()

    //test if the trigger is called before update event
	if !triggerData.FiredBefore() && !triggerData.FiredByUpdate(){
		t.Fatal("function not called BEFORE UPDATE :-O")
	}

    //setting an timestamp collumn to the yesterdays time (it's just an example)
	triggerData.NewRow.Set(4, time.Now().Add(-time.Hour*time.Duration(24)))

	//return ToDatum(nil) //nothing changed in the row
	//return ToDatum(triggerData.OldRow) //nothing changed in the row
	return ToDatum(triggerData.NewRow) //the new row will be changed
}
```

you also must create the trigger function in PostgreSQL and set the trigger to an table event:

```sql
CREATE OR REPLACE FUNCTION public.plgo_trigger()
  RETURNS trigger AS
'$libdir/plgo_test', 'plgo_trigger'
  LANGUAGE c IMMUTABLE STRICT
  COST 1;

CREATE TRIGGER my_awesome_trigger
  BEFORE UPDATE
  ON public.test
  FOR EACH ROW
  EXECUTE PROCEDURE public.plgo_trigger();
```


##TODO

- goroutines test
- range type support
- test test test!
- code generation tool?
- Background Worker Processes!
