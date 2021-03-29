#!/bin/bash
size=2048
cd server
openssl req -newkey rsa:${size} \
                               -new -nodes -x509 \
                               -days 3650 \
                               -out certs/cert.pem \
                               -keyout certs/key.pem \
                               -subj "/C=IT/ST=Brescia/L=Brescia/O=My Organization/OU=Your Unit/CN=localhost" \
															 -addext "subjectAltName = DNS:localhost"

cd ../client
openssl req -newkey rsa:${size} \
                               -new -nodes -x509 \
                               -days 3650 \
                               -out cert/cert.pem \
                               -keyout cert/key.pem \
                               -subj "/C=IT/ST=Brescia/L=Brescia/O=My Organization/OU=Your Unit/CN=localhost" \
															 -addext "subjectAltName = DNS:localhost"
openssl req -newkey rsa:${size} \
                               -new -nodes -x509 \
                               -days 3650 \
                               -out cert2/cert.pem \
                               -keyout cert2/key.pem \
                               -subj "/C=IT/ST=Brescia/L=Brescia/O=My Organization/OU=Your Unit/CN=localhost" \
															 -addext "subjectAltName = DNS:localhost"
cd ../

cp client/cert/cert.pem server/certs/client_cert.pem
cp client/cert2/cert.pem server/certs/client_cert2.pem
cp server/certs/cert.pem client/server_cert.pem

