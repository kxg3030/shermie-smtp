#### 介绍
使用go实现的smtp原生协议

#### 启动
```go
go run server.go
```

#### 测试
```shell
smtp 127.0.0.1:9091 -u example@yeah.net -p 12345 -s "邮件主题"  -n "发件人昵称"   -to "rcpt01@163.com, 收件人昵称 <rcpt2@qq.com>"    -cc "cc01@163.com, 抄送人昵称 <cc2@qq.com>"  -bcc "bcc01@163.com, 密送人昵称 <bcc2@qq.com>" -a "附件.xlsx" -ssl
```