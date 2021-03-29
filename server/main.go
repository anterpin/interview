package main

import (
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/anterpin/interview/server/manager"
)

// process scheduling manager
var _manager manager.Manager

type Client struct {
	id int
}

// client table that associates a request certificate to the calling client
var user_table map[[16]byte]Client

// return the certificate fingerprint using md5
// ? in the future can be implemented setting a certicate field to an encrypted client id
// ? so the server can decrypt on request to extract the client id
// ? it's similar to how jwt cookies works
func useCertificateAsKey(cert *x509.Certificate) [16]byte {
	return md5.Sum(cert.Raw)
}

// retrieve the user id from a request
// extract the certificate from the request
// hash it using the function useCertificateAsKey
// get the client id through the user_table
func getUserId(r *http.Request) (int, error) {
	if r.TLS == nil {
		return -1, errors.New("the request was made without a TLS connection")
	}
	if len(r.TLS.PeerCertificates) == 0 {
		return -1, errors.New("there is no peer certifcate")
	}
	id := useCertificateAsKey(r.TLS.PeerCertificates[0])

	client, exists := user_table[id]
	if !exists {
		return -1, errors.New("unknown user")
	}
	return client.id, nil
}

// create the certificate from the file
// create and entry in the user_table with the certificate and the userid
// call the method AddUser on the global object manager
func setupCertAndManager(file string, userid int) *x509.Certificate {
	// Setup global variable user_table
	// it not suppose for concurrent access
	if user_table == nil {
		user_table = make(map[[16]byte]Client)
	}
	// Setup client cert
	caCertPem, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	block, _ := pem.Decode(caCertPem)
	if block == nil {
		log.Fatal("Failing parsing pem block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	// Setting the user list
	client := Client{id: userid}
	// Setup user table
	user_table[useCertificateAsKey(cert)] = client

	// Setup manager table
	_manager.AddUser(userid)
	return cert
}

// Init global manager
// Setup default server multiplexer
// Setup port
// Setup certificate pool to authenticate clients
// Setup TLS config
// Setup server
// Run the server
func main() {
	// Init global manager
	_manager = manager.NewManager()

	// TODO: use a server multiplexer library https://github.com/gorilla/mux
	// TODO: to limit an endpoint to a specific HTTP method
	// Setup default server multiplexer
	mux := http.DefaultServeMux
	mux.HandleFunc("/start", start)   // POST
	mux.HandleFunc("/stop", stop)     // POST
	mux.HandleFunc("/list", list)     // GET
	mux.HandleFunc("/status", status) // GET
	mux.HandleFunc("/log", _log)      // GET

	// Setup port
	PORT, err := strconv.ParseUint(os.Getenv("PORT"), 10, 64)
	if err != nil || PORT > 65535 {
		//default port
		PORT = 8443
	}

	// Setup certificate pool to authenticate clients
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(setupCertAndManager("certs/client_cert.pem", 1))
	caCertPool.AddCert(setupCertAndManager("certs/client_cert2.pem", 2))

	// Setup TLS config
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
		}, // the cipher suites are not editable in go 1.16 using TLS 1.3
		MinVersion:               tls.VersionTLS13,
		MaxVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
	}

	// Setup server
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", PORT),
		TLSConfig: tlsConfig,
		Handler:   mux,
	}
	server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

	// Run the server
	fmt.Println("Start Server")
	log.Fatal(server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
}
