rm -f proxy
go build -o proxy .
./proxy -l "0.0.0.0:8899" -r "192.168.4.13:8899"
