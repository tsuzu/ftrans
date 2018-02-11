#! /bin/bash

set -eu

GOOS=linux go build -o ftrans *.go
GOOS=darwin go build -o ftrans_mac *.go
GOOS=windows go build -o ftrans.exe *.go

FTRANS=$(openssl md5 ftrans)
FTRANS_MAC=$(openssl md5 ftrans_mac)
FTRANS_WIN=$(openssl md5 ftrans.exe)

cat << EOF
# Version $1

All the following binaries are ones for x86_64.

# Linux
[Download binary]()
$FTRANS

# Mac
[Download binary]()
$FTRANS_MAC

# Windows
[Download binary]()
$FTRANS_WIN

EOF

