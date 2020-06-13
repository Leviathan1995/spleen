# Spleen

Spleen is a SOCKS5 server written in Golang, the server can handles the TCP that support `CONNECT` method.

# How to use

Download the server binary from [release](https://github.com/Leviathan1995/spleen/releases)

```shell
./spleen-server -c .server.json
```

# How to generate pkcs12
```
openssl pkcs12 -export -clcerts -in client.pem -inkey client.key -out root.p12 -passout pass:abc
```

# TODO
* Optimize code
* Support others command
* Support UDP
