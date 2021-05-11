#!/usr/bin/env bash

export CGO_ENABLED=0
export GOARCH=amd64
export GOOS=linux
export GO111MODULE=on

TARGET_CMD=${1:-server}
TARGET_OS=${2:-linux}
TARGET_ARCH=${3:-amd64}

case "${TARGET_OS}" in
"")
  if [[ $(uname) == "Darwin" ]]; then
    export GOOS=darwin
  fi
  ;;
"darwin")
  export GOOS=darwin
  ;;
"windows")
  export GOOS=windows
  ;;
esac

case "${TARGET_ARCH}" in
"arm64")
  export GOARCH=arm64
  ;;
"arm")
  export GOARCH=arm
  ;;
esac

GO_VER=$(go version)
BUILD_VER="${GO_VER: 11}"
if [[ -d ".git" ]]; then
  which git
  if [ $? -eq 0 ]; then
    BUILD_REV="branch $(git symbolic-ref --short -q HEAD) rev $(git rev-parse --short HEAD)"
  fi
fi
BUILD_DATE="$(date "+%Y-%m-%d %H:%M:%S")"
BUILD_INFO=$(
cat<<EOF
${BUILD_VER}
${BUILD_REV}
${BUILD_DATE}
EOF
)
CMD_FILE="vision_${TARGET_CMD}_${GOOS}_${GOARCH}"
go build -ldflags "-extldflags -static -s -w -X 'main.buildInfo=${BUILD_INFO}'" -a -v -o ./bin/${CMD_FILE} ./cmd/

which upx
if [ $? -eq 0 ]; then
  upx -9 bin/${CMD_FILE}
fi