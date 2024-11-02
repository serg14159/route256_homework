#!/bin/bash

GRPC_HOST="localhost:50051"
GRPC_METHOD="loms.Loms/OrderList"

payload=$(
  cat <<EOF
{

}
EOF
)

grpcurl -plaintext -emit-defaults \
  -rpc-header 'x-app-name:dev' \
  -rpc-header 'x-app-version:1' \
  -d "${payload}" ${GRPC_HOST} ${GRPC_METHOD}
