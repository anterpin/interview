# Introduction
This project is an application exercise for Teleport.  
It is a simple job scheduling service provided by a HTTPs server.  
It is divided in three major parts:
- Client CLI
- HTTP RESTful API
- Process scheduling library

# Install
This respository is based on go modules, make sure go can use them.

    go env -w GO111MODULE=auto

Go will automatically install all the dependencies on `go build` or `go test`

# Testing
In order to test the scheduling process library, access the library folder (from the project root directory).

    cd ./server/manager

And then run the tests

    go test -v

