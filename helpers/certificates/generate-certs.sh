#!/bin/bash

set -e

# Create certs directory if it doesn't exist
mkdir -p certs

# Generate CA for Site A
openssl genrsa -out certs/site-a-ca.key 2048
openssl req -new -x509 -days 365 -key certs/site-a-ca.key -out certs/site-a-ca.crt \
    -subj "/C=US/ST=Test/L=Test/O=Site A/CN=Site A CA"

# Generate CA for Site B
openssl genrsa -out certs/site-b-ca.key 2048
openssl req -new -x509 -days 365 -key certs/site-b-ca.key -out certs/site-b-ca.crt \
    -subj "/C=US/ST=Test/L=Test/O=Site B/CN=Site B CA"

# Generate server certificates for Site A
openssl genrsa -out certs/site-a-server.key 2048
openssl req -new -key certs/site-a-server.key -out certs/site-a-server.csr \
    -subj "/C=US/ST=Test/L=Test/O=Site A/CN=site-availability-a"
openssl x509 -req -days 365 -in certs/site-a-server.csr \
    -CA certs/site-a-ca.crt -CAkey certs/site-a-ca.key -CAcreateserial \
    -out certs/site-a-server.crt

# Generate server certificates for Site B
openssl genrsa -out certs/site-b-server.key 2048
openssl req -new -key certs/site-b-server.key -out certs/site-b-server.csr \
    -subj "/C=US/ST=Test/L=Test/O=Site B/CN=site-availability-b"
openssl x509 -req -days 365 -in certs/site-b-server.csr \
    -CA certs/site-b-ca.crt -CAkey certs/site-b-ca.key -CAcreateserial \
    -out certs/site-b-server.crt

# Set appropriate permissions
chmod 644 certs/*.crt
chmod 600 certs/*.key

echo "Certificates generated successfully in the certs directory"
