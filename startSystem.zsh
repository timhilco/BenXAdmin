#!/bin/xsh 
echo "------------------"
echo "Starting BenXAdmin"
pwd
echo "-------------------------"
echo "Starting BenXAdmin Client" 
cd ./client/webApp
pwd
nohup go run main.go > clientWebapp.log & disown 
echo "--------------------------"
echo "Starting BenXAdmin Server"
cd ..
cd ..
cd ./server
pwd
go run main.go
