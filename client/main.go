package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"io"
	"log"
	"net/http"

	"github.com/anterpin/interview/server/apiobj"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	certDir = kingpin.Flag("certDir", "cert.pem key.pem dir").Default("./cert/").String()

	PORT = kingpin.Flag("port", "port").Envar("PORT").Default("8443").Uint16()

	start         = kingpin.Command("start", "run command")
	startCommands = start.Arg("command", "specific command to run").Required().Strings()

	stop   = kingpin.Command("stop", "stop running process")
	stopId = stop.Arg("id", "process identifier").Required().String()

	_ = kingpin.Command("list", "list running processes")

	_log   = kingpin.Command("log", "get ouptut of running process")
	_logId = _log.Arg("id", "process identifier").Required().String()

	status   = kingpin.Command("status", "query status of running process")
	statusId = status.Arg("id", "process identifier").Required().String()
)

func main() {

	log.SetFlags(0)

	command := kingpin.Parse()
	var buffer bytes.Buffer
	id := ""
	method := "POST"
	switch command {
	case "start":
		startCommand := strings.Join(*startCommands, " ")
		json.NewEncoder(&buffer).Encode(apiobj.Command{Command: startCommand})
	case "stop":
		json.NewEncoder(&buffer).Encode(apiobj.UUID{UUID: *stopId})
	case "list":
		method = "GET"
	case "status":
		method = "GET"
		id = *statusId
	case "log":
		method = "GET"
		id = *_logId
	}

	URL := fmt.Sprintf("https://localhost:%d/%s", *PORT, command)

	req, err := http.NewRequest(method, URL, &buffer)
	if err != nil {
		log.Fatal("Cannot create a request")
	}

	// set the get parameter id
	if id != "" {
		q := req.URL.Query()
		q.Add("id", id)
		req.URL.RawQuery = q.Encode()
	}

	// setup certificate authority
	caCert, err := ioutil.ReadFile("server_cert.pem")
	if err != nil {
		log.Fatal(err)
	}

	// setup certificate pool
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatal("Cannot extract ca from cert.pem")
	}
	// we can also use separate cert and key for the client
	cert, err := tls.LoadX509KeyPair(*certDir+"/cert.pem", *certDir+"/key.pem")
	if err != nil {
		log.Fatal(err)
	}

	// setup client tls with ca pool
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,              // client ca
				Certificates: []tls.Certificate{cert}, // client certificate
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorObj := apiobj.Error{}

		getServerResponse(resp.Body, &errorObj)
		fmt.Println(errorObj.Err)
		return
	}
	// io.Copy(os.Stdout, resp.Body)
	switch command {
	case "start":
		uuidObj := apiobj.UUID{}

		getServerResponse(resp.Body, &uuidObj)
		fmt.Println(uuidObj.UUID)
	case "stop":
		statusObj := apiobj.Status{}

		getServerResponse(resp.Body, &statusObj)
		fmt.Println(statusObj.Status)
	case "list":
		listObj := apiobj.List{}

		getServerResponse(resp.Body, &listObj)
		// weird join behaviour
		// ? maybe print a message on empty list
		if len(listObj.List) != 0 {
			fmt.Println(strings.Join(listObj.List, "\n"))
		}
	case "status":
		statusObj := apiobj.State{}

		getServerResponse(resp.Body, &statusObj)
		if statusObj.State != nil {
			fmt.Println(statusObj.State)
		} else {
			fmt.Println("active")
		}
	case "log":
		logObj := apiobj.Log{}

		getServerResponse(resp.Body, &logObj)
		fmt.Print(logObj.Log)
	}
}

func getServerResponse(body io.Reader, obj interface{}) interface{} {
	err := json.NewDecoder(body).Decode(&obj)
	if err != nil {
		log.Fatal("Cannot decode the server response")
	}
	return obj
}
