test:
	go test -v ./...

ROOT ?= $(shell pwd)

go.image.build:
	@echo "-> build go image"
	@docker build -f Dockerfile . -t testy_proto_console_go

GOREPS=\
	userspb

.PHONY: $(GOREPS)
$(GOREPS): go.image.build
	@echo "-> build $@"
	@mkdir -p "$@"
	@docker run -v $(ROOT):/proto-build -e PROTO_NAME=$@ --rm testy_proto_console_go sh -c "\
		cd /proto-build && \
		protoc \
			--go_out=./\$$PROTO_NAME/ \
			--go_opt=module=testyexample \
			--go-grpc_out=./\$$PROTO_NAME/ \
			--go-grpc_opt=module=testyexample \
			--openapiv2_out=logtostderr=true:./\$$PROTO_NAME \
			--openapiv2_opt=output_format=json \
			--grpc-gateway_out=./\$$PROTO_NAME/ \
			--grpc-gateway_opt=logtostderr=true \
			--grpc-gateway_opt=paths=source_relative \
			--doc_out=./\$$PROTO_NAME/ \
			--doc_opt=markdown,\$$PROTO_NAME.md \
			--experimental_allow_proto3_optional \
			-I./src \
			./src/\$$PROTO_NAME/service.proto ./src/\$$PROTO_NAME/common.proto;"
	@if [ -d "$@/userspb" ]; then \
		mv $@/userspb/*.go $@/ 2>/dev/null || true; \
		mv $@/userspb/*.json $@/ 2>/dev/null || true; \
		rm -rf $@/userspb; \
	fi
