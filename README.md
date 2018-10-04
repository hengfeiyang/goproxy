# goproxy
a simple tcp/udp proxy

## Usage

```
Usage of ./goproxy:
  -l string
    	-l=0.0.0.0:8080 指定本地监听端口
  -p string
    	-p=tcp 指定协议类型 tcp/udp (default "tcp")
  -s string
    	-s=127.0.0.1:80,127.0.0.1:81 指定真实服务器地址,多个用','隔开
```