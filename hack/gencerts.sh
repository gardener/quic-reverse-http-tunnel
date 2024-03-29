#!/usr/bin/env bash

# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Modifications Copyright 2024 SAP SE or an SAP affiliate company and Gardener contributors

set -e

# gencerts.sh generates the certificates for the webhook authz plugin tests.

REPO_ROOT="$(git rev-parse --show-toplevel)"
CERTS_DIR=${1:-$REPO_ROOT/dev/certs}

mkdir -p "${CERTS_DIR}"
pushd "${CERTS_DIR}"
trap popd ERR EXIT

cat > server.conf << EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = quic-tunnel-server
IP.1 = 127.0.0.1
EOF

cat > client.conf << EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

# Create a certificate authority
openssl genrsa -out ca.key 3072
openssl req -x509 -new -nodes -key ca.key -days 1 -out ca.crt -subj "/CN=quic-tunnel-ca"

# Create a server certiticate
openssl genrsa -out tls.key 3072
openssl req -new -key tls.key -out server.csr -subj "/CN=quic-tunnel-server" -config server.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt -days 1 -extensions v3_req -extfile server.conf

# Create a client certiticate
openssl genrsa -out client.key 3072
openssl req -new -key client.key -out client.csr -subj "/CN=quic-tunnel-client" -config client.conf
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 1 -extensions v3_req -extfile client.conf

# Clean up after we're done.
rm ./*.csr
rm ./*.srl
rm ./*.conf
