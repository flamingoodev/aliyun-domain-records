# 实现阿里云动态域名解析DDNS

## 前提条件

1. 需要运行本程序的主机有公网IP
2. 需要阿里云账号和访问密钥（AccessKey）。 请在阿里云控制台中的AccessKey管理页面上创建和查看AccessKey。
3. go开发环境
```shell
go version
# output
# go version go1.18.1 darwin/amd64
```

## 使用说明

1. 修改```main```函数中的各项参数
2. 修改```Initialization```函数中的```AccessKeyId```和```AccessKeySecret```
3. 使用crontab定时运行，将编译后的程序放置到```/usr/local/bin/```或者其他目录，然后新增一条crontab记录
    ```shell
    # 编译程序
    go build
    # 编辑crontab
    crontab -e
    # 每一分钟执行一次 /usr/local/bin/adr为本程序的安装路径
    * * * * * /usr/local/bin/adr
    
    ```
4. 直接运行
    ```shell
    go run main.go
    ```
5. 若需要开机启动，则需设置Linux系统服务

## 注意事项
**重要！！！**  
个人公网IP的话，一般情况下80、443端口会被禁用

为了安全起见，建议部署时使用环境变量获取```AccessKeyId```和```AccessKeySecret```

```shell
export ADR_ALIYUN_ACCESS_KEY_ID="xxx"
export ADR_ALIYUN_ACCESS_KEY_SECRET="xxx"
export ADR_DOMAIN_NAME="xxx.com"
```

## 直接运行

```shell
cd aliyun-domain-records
export ADR_ALIYUN_ACCESS_KEY_ID="xxx"
export ADR_ALIYUN_ACCESS_KEY_SECRET="xxx"
export ADR_DOMAIN_NAME="xxx.com"
go mod download
go run main.go
```

## 2022-05-28更新
在程序中增加定时器，不用添加到Linux定时任务中即可定时执行

```shell
# 直接运行
./adr
# 输出日志文件
./adr > /var/log/adr.log
# 后台执行并输出日志
nohup /usr/local/bin/adr > /var/log/adr.log 2>&1 &
```