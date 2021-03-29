package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/anterpin/interview/server/apiobj"
	"github.com/anterpin/interview/server/manager"
)

func setupCert(file string, t *testing.T) *x509.Certificate {

	caCertPem, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(caCertPem)
	if block == nil {
		t.Fatal("Failing parsing pem block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

func TestAuthorization(t *testing.T) {

	// setup manager
	_manager = manager.NewManager()
	_manager.AddUser(1)
	cert1 := setupCertAndManager("certs/client_cert.pem", 1) // add to user_table and to manager
	cert2 := setupCert("certs/client_cert2.pem", t)          // get only the certificate

	tt := []struct {
		name   string
		certs  []*x509.Certificate
		permit bool
	}{
		{"registered certificate", []*x509.Certificate{cert1}, true},
		{"not registered certificate", []*x509.Certificate{cert2}, false},
		{"no certificate", []*x509.Certificate{}, false},
		{"list certificate", []*x509.Certificate{cert1, cert2}, true},
		{"reverse list certificate", []*x509.Certificate{cert2, cert1}, false},
		{"no TLS", []*x509.Certificate{}, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// setup buffer request
			buffer := bytes.Buffer{}
			json.NewEncoder(&buffer).Encode(struct {
				Command string `json:"command"`
			}{
				Command: "ls",
			})

			req, err := http.NewRequest("GET", "localhost:8000/start", &buffer)
			if err != nil {
				t.Fatal("could not create a request")
			}
			if tc.name != "no TLS" {
				req.TLS = new(tls.ConnectionState)
				req.TLS.PeerCertificates = tc.certs
			}

			rec := httptest.NewRecorder()
			// test
			start(rec, req)

			// check
			res := rec.Result()
			defer res.Body.Close()

			io.Copy(os.Stdout, res.Body)
			println()
			permitted := res.StatusCode != http.StatusForbidden
			if permitted != tc.permit {
				t.Fatalf("test not pass %s", tc.name)
			}
		})
	}

}

func makeRequest(testServer *httptest.Server, certs []*x509.Certificate, endpoint string, obj interface{}, key string, values []string) *http.Request {
	buffer := bytes.Buffer{}
	if obj != nil {
		json.NewEncoder(&buffer).Encode(obj)
	}
	// the request method does not matter
	req, err := http.NewRequest("GET", testServer.URL+"/"+endpoint, &buffer)
	if err != nil {
		log.Fatal("could not create a request")
	}
	req.TLS = new(tls.ConnectionState)
	req.TLS.PeerCertificates = certs

	if key != "" {
		q := req.URL.Query()
		for _, s := range values {
			q.Add(key, s)
		}
		req.URL.RawQuery = q.Encode()
	}
	return req
}
func makeRequest2(testServer *httptest.Server, certs []*x509.Certificate, endpoint string, body string, uri string) *http.Request {
	buffer := bytes.Buffer{}
	buffer.WriteString(body)
	// the request method does not matter
	req, err := http.NewRequest("GET", testServer.URL+"/"+endpoint+"?"+uri, &buffer)
	if err != nil {
		log.Fatal("could not create a request")
	}
	req.TLS = new(tls.ConnectionState)
	req.TLS.PeerCertificates = certs

	return req
}
func TestHandlers(t *testing.T) {
	_manager = manager.NewManager()
	_manager.AddUser(1)
	cert1 := setupCertAndManager("certs/client_cert.pem", 1) // add to user_table and to manager
	cert2 := setupCert("certs/client_cert2.pem", t)          // get only the certificate

	uuid, err := _manager.Start("watch date", 1)
	if err != nil {
		t.Fatal("cannot start the test")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/stop", stop)
	mux.HandleFunc("/list", list)
	mux.HandleFunc("/status", status)
	mux.HandleFunc("/log", _log)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	// endpoint test
	type ET struct {
		req        *http.Request
		statusCode int
	}
	tt := []struct {
		name     string
		subtests []ET
	}{
		{
			"unknown client",
			[]ET{
				{makeRequest(srv, []*x509.Certificate{cert2}, "list", nil, "", nil), http.StatusForbidden},
				{makeRequest(srv, []*x509.Certificate{cert2}, "log", nil, "", nil), http.StatusForbidden},
				{makeRequest(srv, []*x509.Certificate{cert2}, "status", nil, "", nil), http.StatusForbidden},
				{makeRequest(srv, []*x509.Certificate{cert2}, "stop", nil, "", nil), http.StatusForbidden},
			},
		},
		{
			"normal request",
			[]ET{
				{makeRequest(srv, []*x509.Certificate{cert1}, "list", nil, "", nil), http.StatusOK},
				{makeRequest(srv, []*x509.Certificate{cert1}, "log", nil, "id", []string{uuid}), http.StatusOK},
				{makeRequest(srv, []*x509.Certificate{cert1}, "status", nil, "id", []string{uuid}), http.StatusOK},
				{makeRequest(srv, []*x509.Certificate{cert1}, "stop", apiobj.UUID{UUID: uuid}, "", nil), http.StatusOK},
			},
		},
		{
			"empty request",
			[]ET{
				{makeRequest(srv, []*x509.Certificate{cert1}, "list", nil, "", nil), http.StatusOK},
				{makeRequest(srv, []*x509.Certificate{cert1}, "log", nil, "", nil), http.StatusBadRequest},
				{makeRequest(srv, []*x509.Certificate{cert1}, "status", nil, "", nil), http.StatusBadRequest},
				{makeRequest(srv, []*x509.Certificate{cert1}, "stop", nil, "", nil), http.StatusBadRequest},
			},
		},
		{
			"wrong type",
			[]ET{
				{makeRequest2(srv, []*x509.Certificate{cert1}, "list", "", ""), http.StatusOK},
				{makeRequest2(srv, []*x509.Certificate{cert1}, "log", "", "aldk"), http.StatusBadRequest},
				{makeRequest2(srv, []*x509.Certificate{cert1}, "status", "", "aldk"), http.StatusBadRequest},
				{makeRequest2(srv, []*x509.Certificate{cert1}, "stop", "something", ""), http.StatusBadRequest},
			},
		},
		{
			"bad format",
			[]ET{
				{makeRequest(srv, []*x509.Certificate{cert1}, "list", nil, "", nil), http.StatusOK},
				{makeRequest(srv, []*x509.Certificate{cert1}, "log", nil, "id", []string{"hello", uuid}), http.StatusBadRequest},
				{makeRequest(srv, []*x509.Certificate{cert1}, "status", nil, "id", []string{"hello", uuid}), http.StatusBadRequest},
				{makeRequest(srv, []*x509.Certificate{cert1}, "stop", struct{ What string }{uuid}, "", nil), http.StatusBadRequest},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// list
			// sequential tests
			for _, st := range tc.subtests {
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, st.req)
				res := rec.Result()

				if res.StatusCode != st.statusCode {
					t.Errorf("error on test %s %d", st.req.URL, res.StatusCode)
					io.Copy(os.Stdout, res.Body)
				}
				res.Body.Close()
			}

		})
	}
}
