# FINBIN

find stuff in binary content or find keychain files.

## Install

```sh
go get github.com/threez/finbin/cmd/finbin
go get github.com/threez/finbin/cmd/finkc
```

## Usage finkc

Finds mac os keychain files.

All data will be stored with finding location into the provided
dir. The name of the file is `file-<offset>.keychain` offset is the byte
offset of the match.

```sh
$ finkc -file /dev/sda
Found file-1231231223.keychain (55018 to 55038 size 20.00B)
```

## Usage finbin

The pattern will be used to search for the content. Size is
the size of data to extract before **and** after the location.
All data will be stored with finding location into the provided
dir. The name of the file is `file-<offset>` offset is the byte
offset of the match.

```sh
$ finbin -file /dev/sda -pattern "kych\x00|SQLite format 3\x00|<key>ACKeychainItemVersion</key>" -size 10B -dir found
Found file-1231231223 (55018 to 55038 size 20.00B)
```
