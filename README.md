<!---
 Copyright (c) 2018 Tsuzu
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# ftrans
- simple File TRANSfer program

# Installation
- Download from [releases](../../releases)

# For Developers
- $ go get -u -d github.com/cs3238-tsuzu/ftrans
- $ cd $GOPATH/src/github.com/cs3238-tsuzu/ftrans
- $ make # On windows, make BINARY_NAME=ftrans.exe
- Make sure that ftrans works by executing `./ftrans`

# Usage
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
- 1.0