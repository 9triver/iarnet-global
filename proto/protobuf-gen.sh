#!/bin/bash

# Simple protobuf generation script for iarnet-global
# Usage: ./proto/protobuf-gen.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="${ROOT_DIR}/proto"
GO_OUT_DIR="${ROOT_DIR}/internal/proto"

export PATH="$PATH:$HOME/go/bin"

PROTOC_CMD="${PROTOC_CMD:-python -m grpc_tools.protoc}"

echo "=========================================="
echo "iarnet-global Protobuf Generation"
echo "Project root : ${ROOT_DIR}"
echo "Proto root   : ${PROTO_DIR}"
echo "Go output    : ${GO_OUT_DIR}"
echo "Using protoc : ${PROTOC_CMD}"
echo "=========================================="

run_protoc() {
  eval "${PROTOC_CMD} $*"
}

if ! run_protoc --version >/dev/null 2>&1; then
  cat >&2 <<EOF
Error: Unable to execute '${PROTOC_CMD}'.
Please ensure python 环境已安装 grpc_tools.protoc，
或通过设置 PROTOC_CMD 指向可用的 protoc 命令。
EOF
  exit 1
fi

if ! command -v protoc-gen-go >/dev/null 2>&1 || ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  cat >&2 <<'EOF'
Error: protoc-gen-go and/or protoc-gen-go-grpc not found in PATH.
Install them with:
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
EOF
  exit 1
fi

mkdir -p "${GO_OUT_DIR}"

generate_go() {
  local rel_dir="$1"
  local out_dir="$2"
  shift 2
  local files=("$@")

  mkdir -p "${out_dir}"
  rm -f "${out_dir}"/*.pb.go

  pushd "${PROTO_DIR}/${rel_dir}" >/dev/null
  echo ""
  echo ">>> Generating ${rel_dir} -> ${out_dir}"
  run_protoc \
    -I "${PROTO_DIR}" -I . \
    --go_out="${out_dir}" --go_opt=paths=source_relative \
    --go-grpc_out="${out_dir}" --go-grpc_opt=paths=source_relative \
    "${files[@]}"
  popd >/dev/null
}

# 清理旧的 resource/scheduler 目录（历史迁移遗留）
rm -rf "${GO_OUT_DIR}/resource/scheduler"

generate_go "resource" "${GO_OUT_DIR}/resource" "resource.proto"
generate_go "resource/scheduler" "${GO_OUT_DIR}/scheduler" "scheduler.proto"


echo ""
echo "Done."
echo "=========================================="
