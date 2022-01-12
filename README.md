# spleen

轻量级内网穿透工具, 使用 spleen 可以支持通过外网访问不具备公网 IP 的家庭服务器/内网主机.

## 介绍

通过在一台具有公网 IP 的小型服务器(阿里云轻量云)部署 spleen, 可以随时随地访问你的家庭服务器/内网主机(闲置笔记本)的 TCP 服务, 例如 SSH, HTTP/S 等.

## SSH 服务样例

## 如何使用

### 公网服务器部署 spleen-server

* 通过 [release]() 下载对应架构的 spleen 包:
```shell
# wget 下载
> wget 

# 解压
> tar

# 配置端口转发规则
> vim .server.json

{
  "ServerIP"   : "0.0.0.0",  # 公网服务器监听地址
  "ServerPort"   : 1234,  # 公网服务器监听端口, 该端口用来与家庭服务器/内网主机建立通信隧道
  "MappingPort" : [ # 端口映射规则
	"5000:22",  # 即访问公网服务器的 5000 端口就等于访问家庭服务器/内网主机的 22 端口
	"5001:3306"
	...
  ]
}

# 启动
> ./spleen-server -c .server.json
2022/01/12 19:39:39 The server listening for the intranet server at 0.0.0.0:1234 successful.
2022/01/12 19:39:39 The server listening at 0.0.0.0:5001 successful.
2022/01/12 19:39:39 The server listening at 0.0.0.0:5000 successful.
```

### 家庭服务器/内网主机部署 spleen-client

* 通过 [release]() 下载对应架构的 spleen 包:
```shell
# wget 下载
> wget 

# 解压
> tar

# 配置公网服务器地址
> vim .client.json

{
  "server_ip"  : "127.0.0.1", # 公网服务器 IP
  "server_port": 1234 # 公网服务器监听端口
}

# 启动
> ./spleen-client -c .client.json # 默认预留 10 个活跃连接
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
2022/01/12 18:55:19 Connect to the server 127.0.0.1:1234 successful.
```

## TODO

* 增加安全性配置, 鉴权
* 支持 UDP 转发
* 支持 QUIC 传输协议提升访问速度

###
# License
[GNU General Public License v3.0](https://github.com/Leviathan1995/spleen/blob/master/LICENSE)
