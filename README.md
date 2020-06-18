# Spleen

Ths SOCKS5 over TLS

# How to use

Download the latest binary from [release](https://github.com/Leviathan1995/spleen/releases)

# How to generate pkcs12

```shell
openssl pkcs12 -export -clcerts -in client.pem -inkey client.key -out root.p12 -passout pass:abc
```

# License
[GNU General Public License v3.0](https://github.com/Leviathan1995/spleen/blob/master/LICENSE)
