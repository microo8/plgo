go build -buildmode=c-shared -o plgo_test.so plgotest.go pl.go && sudo cp plgo_test.so $(pg_config --pkglibdir) &&
psql -U root -d meh -c "select plgo_test()"
rm plgo_test.so
rm plgo_test.h
