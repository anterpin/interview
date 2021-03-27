# Introduction
This project is an interview exercise at Teleport.  
It is a simple job scheduling service provided by and HTTPs server.  
It is divided in major parts:
- Client CLI
- HTTP RESTful API
- Process scheduling library

# Build
This respository is based on go modules, make sure go can use them.

    go env -w GO111MODULE=auto

Then install all dependencies

    go mod tidy

# Testing
## Process scheduling library

From the project directory

    cd ./server/manager

And then run the tests

    go test -v

