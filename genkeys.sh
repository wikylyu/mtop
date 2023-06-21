#!/bin/bash
#
# This script helps to generate a self-signed certificate
# 
# If you want to use self-signed certificate for your server,
# modify IP:127.0.0.1 to your domain, such as DNS:example.com.
# 

read -p "Enter your server domain or ip (IP:111.111.111.11 or DNS:example.com): " domain

mkdir -p keys

openssl genrsa -out keys/ca.key 2048
openssl req -new -x509 -days 365 -key keys/ca.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=Acme Root CA" -out keys/ca.crt

openssl req -newkey rsa:2048 -nodes -keyout keys/server.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=localhost" -out keys/server.csr
openssl x509 -req -extfile <(printf "subjectAltName=$domain") -days 365 -in keys/server.csr -CA keys/ca.crt -CAkey keys/ca.key -CAcreateserial -out keys/server.crt
