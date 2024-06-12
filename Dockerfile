# Use an official Go runtime as the parent image
FROM golang:1.22-alpine AS builder

# Fetch protoc and protobuf's go plugins
WORKDIR /usr/local/bin
RUN apk add --no-cache curl && \
    curl -LO "https://github.com/protocolbuffers/protobuf/releases/download/v26.1/protoc-26.1-linux-x86_64.zip" && \
    unzip protoc-26.1-linux-x86_64.zip && \
    mv bin/protoc /usr/local/bin && \
    rm -rf bin protoc-26.1-linux-x86_64.zip
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

FROM builder

# # Compile proto files to Go stubs
WORKDIR /app
COPY . .
ENV OUT=pkg/mp
ENV OPT=paths=source_relative
ENTRYPOINT protoc -I api --go_out $OUT --go-grpc_out $OUT \
    --grpc-gateway_out $OUT --openapiv2_out $OUT \
    --go_opt $OPT --go-grpc_opt $OPT --grpc-gateway_opt $OPT \
    --grpc-gateway_opt generate_unbound_methods=true \
    api/v1/*.proto
