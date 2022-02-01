# spleen

轻量级内网穿透工具, 使用 `spleen` 可以支持通过外网访问不具备公网 IP 的家庭服务器/内网主机.

## 介绍

通过在一台具有公网 IP 的小型服务器(阿里云轻量)部署 `spleen`, 可以随时随地访问你的家庭服务器/内网主机(闲置笔记本)的 `TCP` 服务, 例如 `SSH`, `HTTP/S` 等. **`spleen`支持针对内网服务器指定服务进行流速限制, 避免打爆公网小水管**.

例如 `SSH` 服务, 在顺利部署 `spleen` 的客户端和服务端后, 通过公网服务器(假定 IP 为`1.1.1.1`), 可以直接通过端口映射来连接你的家庭服务器/内网主机:

```shell
执行 ssh -p 5000 leviathan@1.1.1.1 # 即可直接连接到家庭服务器/内网主机
```

## 如何使用

### 家庭服务器/内网主机部署 spleen-client

* 通过 [release](https://github.com/Leviathan1995/spleen/releases) 下载对应架构的 spleen 包:
```shell
# wget 下载 (请自行替换最新版本)
> wget https://github.com/Leviathan1995/spleen/releases/download/v0.0.5/spleen_0.0.5_Linux_64-bit.tar.gz

# 解压
> tar -zxvf spleen_0.0.5_Linux_64-bit.tar.gz
> cd spleen_0.0.5_Linux_64-bit/

# 配置公网服务器地址
> vim .spleen-client.json
{
  "ClientID"  : 1, # 内网服务器 ID, 全局唯一 1, 2, ...
  "ServerIP"  : "127.0.0.1", # 公网服务器 IP
  "ServerPort": 1234, # 公网服务器监听端口
  "LimitRate" : [
    "5001:512" # 指定内网服务的流速限制, 例如指定内网 5001 端口的流速不要超过 512 KB/s, 单位是 KB, 默认不限制, 0 值即为不限制.
  ]
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

### 公网服务器部署 spleen-server

* 通过 [release](https://github.com/Leviathan1995/spleen/releases) 下载对应架构的 spleen 包:
```shell
# wget 下载 (请自行替换最新版本)
> wget https://github.com/Leviathan1995/spleen/releases/download/v0.0.5/spleen_0.0.5_Linux_64-bit.tar.gz

# 解压
> tar -zxvf spleen_0.0.5_Linux_64-bit.tar.gz
> cd spleen_0.0.5_Linux_64-bit/

# 配置端口转发规则
> vim .spleen-server.json

{
  "ServerIP"   : "0.0.0.0", # 公网服务器监听地址
  "ServerPort"   : 1234, # 公网服务器监听端口, 该端口用来与家庭服务器/内网主机建立通信隧道
  "Rules" : [ # 端口映射规则
    {
      # 即访问公网服务器的 5000 端口等于直接访问 ID 为 1 的家庭服务器/内网主机的 22 端口
      "ClientID" : 1,     # Client ID, 在 .spleen-client.json 中配置
      "LocalPort" : 5000, # 公网端口
      "MappingPort" : 22  # 内网转发端口
    },
    {
      # 即访问公网服务器的 5001 端口等于直接访问 ID 为 2 的家庭服务器/内网主机的 443 端口
      "ClientID" : 2,     # Client ID, 在 .spleen-client.json 中配置
      "LocalPort" : 5001, # 公网端口
      "MappingPort" : 443 # 内网转发端口
    }
  ]
}

# 启动
> ./spleen-server -c .server.json
2022/01/12 19:39:39 The server listening for the intranet server at 0.0.0.0:1234 successful.
2022/01/12 19:39:39 The server listening at 0.0.0.0:5001 successful.
2022/01/12 19:39:39 The server listening at 0.0.0.0:5000 successful.
```

## SSH 服务样例
当我们分别按照上述步骤在公网服务器部署了 `spleen-server`、家庭服务器/内网主机部署了 `spleen-client`后, 通过设定的转发规则 `5000:22`即访问公网服务器的 5000 端口就等于访问家庭服务器/内网主机的 22 端口.
我们可以直接使用 `SSH` 连接家庭服务器/内网主机, 假如公网 IP 为 `1.1.1.1`:
```shell
ssh -p 5000 leviathan@1.1.1.1 # 即可直接连接到家庭服务器/内网主机
```

## TODO

* 增加安全性配置, 鉴权
* 支持 UDP 转发
* 支持 QUIC 传输协议提升访问速度

###
# License
[GNU General Public License v3.0](https://github.com/Leviathan1995/spleen/blob/master/LICENSE)
