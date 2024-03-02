#!/bin/xsh 
echo "Starting BenXAdmin"
pwd
echo "Starting BenXAdmin Client" 
echo "Starting BenXAdmin Server"
nohup go run ./client/webAppmain.go & disown
go run ./server/main.go
