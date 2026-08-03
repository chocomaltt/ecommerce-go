#!/bin/sh
set -eu

cd "$(dirname "$0")"
rm -f ca.crt ca.key ca.srl *.csr *.ext edge-api.crt edge-api.key identity-service.crt identity-service.key order-service.crt order-service.key identity-actor.key identity-actor.pub

openssl req -x509 -newkey rsa:3072 -sha256 -days 3650 -nodes \
  -keyout ca.key -out ca.crt -subj "/CN=commerce-local-ca"

issue() {
  name="$1"
  dns="$2"
  openssl req -newkey rsa:3072 -sha256 -nodes \
    -keyout "$name.key" -out "$name.csr" -subj "/CN=$dns"
  printf 'subjectAltName=DNS:%s\nextendedKeyUsage=clientAuth,serverAuth\n' "$dns" > "$name.ext"
  openssl x509 -req -sha256 -days 365 \
    -in "$name.csr" -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out "$name.crt" -extfile "$name.ext"
  rm -f "$name.csr" "$name.ext"
}

issue edge-api edge-api.internal
issue identity-service identity-service.internal
issue order-service order-service.internal
openssl genpkey -algorithm Ed25519 -out identity-actor.key
openssl pkey -in identity-actor.key -pubout -out identity-actor.pub
chmod 600 ca.key edge-api.key identity-service.key order-service.key identity-actor.key
