# plgo
trying to make postgresql stored procedures extension in golang

currently just experimenting, but some types conversion Go <-> C <-> Postgres run ok and the example also :)

also SPI prepare and query is running!

just run:

```
$ ./run.sh
```

create a table:

```sql
CREATE TABLE test
(
  id serial,
  txt text NOT NULL,
  CONSTRAINT test_pk PRIMARY KEY (id)
);
insert into test (txt) values ('meh');
```

and create the function in postgresql:

```
psql -U root -d database -f example.sql
```

then you can call the function:

```
psql -U root -d database -c "select plgo_example('foo', 10)"

```
output:

```
                         plgo_example                         
--------------------------------------------------------------
 foomehfoomehfoomehfoomehfoomehfoomehfoomehfoomehfoomehfoomeh
(1 row)
```
