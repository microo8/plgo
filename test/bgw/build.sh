go build -buildmode=c-shared -o bgw.so && sudo cp bgw.so `pg_config --pkglibdir` && sudo chmod 755 `pg_config --pkglibdir`/* && sudo systemctl restart postgresql.service
