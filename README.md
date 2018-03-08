<!---
 Copyright (c) 2018 Tsuzu
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# ftrans
[![Travis](https://img.shields.io/travis/cs3238-tsuzu/ftrans.svg?style=flat-square)](https://travis-ci.org/cs3238-tsuzu/ftrans)
[![Go Report Card](https://goreportcard.com/badge/github.com/cs3238-tsuzu/ftrans?style=flat-square)](https://goreportcard.com/report/github.com/cs3238-tsuzu/ftrans)

# Description
- simple File TRANSfer program
- Quick and fast transferring with P2P(P2P library is [go-easyp2p](https://github.com/cs3238-tsuzu/go-easyp2p))
- Single & statically linked binary
- Pure Go
- Simple and easy installation
- Maybe useful to transfer files from host machines to VM, from PC to VPS, and so on.

# Installation
- Download from [releases](../../releases)

# For Developers
- $ go get -u -d github.com/cs3238-tsuzu/ftrans
- $ cd $GOPATH/src/github.com/cs3238-tsuzu/ftrans
- $ make # On windows, make BINARY_NAME=ftrans.exe
- Make sure that ftrans works by executing `./ftrans`

# Usage
- In Japanese, read [the wiki](https://github.com/cs3238-tsuzu/ftrans/wiki)
- On one computer

```
$ ftrans send .vimrc ~/.ssh/id_rsa.pub portrait.jpg
20XX/XX/XX XX:XX:XX Your password: password
```

- On the other computer

```
$ ftrans recv -p password
```

- You can set password for yourself using "-p" option

- Only these steps enable you to send and recieve files.
- If you want to use other options, run ftrans -h

# License
- Under the MIT License
- Copyright (c) 2018 Tsuzu

# Version
- 1.1