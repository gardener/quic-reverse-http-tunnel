# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

REGISTRY              ?= ghcr.io/gardener/quic-reverse-http-tunnel
CLIENT_IMAGE_NAME     := $(REGISTRY)/quic-client
CLIENT_TCP_IMAGE_NAME := $(REGISTRY)/quic-client-tcp
SERVER_IMAGE_NAME     := $(REGISTRY)/quic-server
IMAGE_TAG             ?= local-dev

LOCAL_CERTS_DIR := dev/certs
LOCAL_CERTS     := $(LOCAL_CERTS_DIR)/ca.crt $(LOCAL_CERTS_DIR)/ca.key $(LOCAL_CERTS_DIR)/client.crt $(LOCAL_CERTS_DIR)/client.key $(LOCAL_CERTS_DIR)/tls.crt $(LOCAL_CERTS_DIR)/tls.key

#########################################
# Rules for local development scenarios #
#########################################

$(LOCAL_CERTS):
	@bash hack/gencerts.sh $(LOCAL_CERTS_DIR)

.PHONY: start-server
start-server: $(LOCAL_CERTS)
	@go run cmd/server/main.go \
		--listen-tcp 0.0.0.0:10443 \
		--listen-quic 0.0.0.0:8888 \
		--cert-file $(LOCAL_CERTS_DIR)/tls.crt \
		--cert-key $(LOCAL_CERTS_DIR)/tls.key \
		--client-ca-file $(LOCAL_CERTS_DIR)/ca.crt \
		--v=2

.PHONY: start-client
start-client: $(LOCAL_CERTS)
	@go run cmd/client/main.go \
		--server=localhost:8888 \
		--ca-file $(LOCAL_CERTS_DIR)/ca.crt \
		--cert-file $(LOCAL_CERTS_DIR)/client.crt \
		--cert-key $(LOCAL_CERTS_DIR)/client.key \
		--v=2

.PHONY: start-client-tcp
start-client-tcp: $(LOCAL_CERTS)
	@go run cmd/client-tcp/main.go \
		--server=localhost:8888 \
		--ca-file $(LOCAL_CERTS_DIR)/ca.crt \
		--cert-file $(LOCAL_CERTS_DIR)/client.crt \
		--cert-key $(LOCAL_CERTS_DIR)/client.key \
		--upstream=www.example.com:80 \
		--v=2

########################################################
# Rules related to binary build and Docker image build #
########################################################

.PHONY: docker-images
docker-images:
	@docker build --platform linux/amd64,linux/arm64 -t $(CLIENT_IMAGE_NAME):$(IMAGE_TAG) -t $(CLIENT_IMAGE_NAME):latest -f Dockerfile --target quic-client .
	@docker build --platform linux/amd64,linux/arm64 -t $(CLIENT_TCP_IMAGE_NAME):$(IMAGE_TAG) -t $(CLIENT_TCP_IMAGE_NAME):latest -f Dockerfile --target quic-client-tcp .
	@docker build --platform linux/amd64,linux/arm64 -t $(SERVER_IMAGE_NAME):$(IMAGE_TAG) -t $(SERVER_IMAGE_NAME):latest -f Dockerfile --target quic-server .

.PHONY: install
install:
	@CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) GO111MODULE=on go install ./...
