rm func.so
go build -v -buildmode=c-shared -o func.so func.go pl.go && sudo cp func.so /usr/lib/postgresql &&
sudo systemctl restart postgresql.service && psql -U root -d meh -c "select plgo_func('foo', 10)"
