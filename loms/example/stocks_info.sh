#!/bin/bash

GRPC_HOST="localhost:50051"
GRPC_METHOD="loms.Loms/StocksInfo"

payload=$(
  cat <<EOF
{
  "sku": 1003
}
EOF
)

grpcurl -plaintext -emit-defaults \
  -rpc-header 'x-app-name:dev' \
  -rpc-header 'x-app-version:1' \
  -d "${payload}" ${GRPC_HOST} ${GRPC_METHOD}