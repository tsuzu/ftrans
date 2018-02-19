#! /bin/bash

set -eu

export GOARCH=amd64
echo "Build for Mac..."
make GOOS=darwin BINARY_NAME=ftrans_mac
echo

echo "Build for Linux..."
make GOOS=linux BINARY_NAME=ftrans
echo

echo "Build for Windows..."
make GOOS=windows BINARY_NAME=ftrans.exe
echo

FTRANS=$(openssl md5 ftrans)
FTRANS_MAC=$(openssl md5 ftrans_mac)
FTRANS_WIN=$(openssl md5 ftrans.exe)

cat << EOF
# Version $1

All the following binaries are ones for x86_64.

## Linux
$FTRANS

## Mac
$FTRANS_MAC

## Windows
$FTRANS_WIN

EOF

