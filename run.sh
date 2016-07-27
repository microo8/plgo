rm example.so
go build -v -buildmode=c-shared -o example.so example.go pl.go && sudo cp example.so /usr/lib/postgresql &&
sudo systemctl restart postgresql.service && psql -U root -d meh -c "select plgo_example('foo', 10)"
