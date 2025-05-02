#!/bin/bash

CERT_FILE="cert.pem"
KEY_FILE="key.pem"
DAYS_VALID=365
COMMON_NAME="localhost"

openssl genpkey -algorithm RSA -out "$KEY_FILE"

openssl req -x509 -new -nodes -key "$KEY_FILE" -sha256 -days "$DAYS_VALID" -out "$CERT_FILE" -subj "/CN=$COMMON_NAME"

echo "Completed"