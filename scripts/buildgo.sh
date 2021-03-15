#!/bin/bash

set -e

cd src/bindings

# TMPGOPATH=$(mktemp -d /tmp/genji.XXXX)

# go mod vendor

# mv vendor $TMPGOPATH/src

# mkdir -p $TMPGOPATH/src/github.com/genjidb/genji.js/src/bindings
# cp -r . $TMPGOPATH/src/github.com/genjidb/genji.js/src/bindings/.

# ls -lh $TMPGOPATH/src/github.com/genjidb/genji.js/src/bindings

# docker run \
#     --rm \
#     -v $TMPGOPATH:/go \
#     -v "$(pwd)":/dist \
#     -e "GOPATH=/go" \
#     tinygo/tinygo:0.17.0 tinygo build -o /dist/genji.wasm -target wasm --no-debug github.com/genjidb/genji.js/src/bindings

# mv genji.wasm ../../dist/.

tinygo build -o ../../dist/genji.wasm -target wasm --no-debug .
