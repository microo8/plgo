# plgo
trying to make postgresql stored procedures extension in golang

currently just experimenting, but some types conversion Go <-> C <-> Postgres run ok and the example also :)

just run:

```
$ ./run.sh
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
--------------------------------
 foofoofoofoofoofoofoofoofoofoo
(1 row)
```

also SPI prepare and query is running!
