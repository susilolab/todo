@echo off

rem go build -ldflags "-s" -o server_bolt.exe serverb.go
rem server_bolt.exe
gin --bin server_bolt.exe -i --port "3000" --appPort "3001" run