cp ../pl.go .
cp ../funcdec.h .
go build -buildmode=c-shared -o plgo_test.so plgotest.go pl.go && sudo cp plgo_test.so $(pg_config --pkglibdir) &&
psql -U root -d postgres -c "select plgo_test()"
rm pl.go
rm funcdec.h
rm plgo_test.so
rm plgo_test.h
