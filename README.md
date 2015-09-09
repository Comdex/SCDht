# Golang语言实现的DHT网络爬虫整站程序

作者博客：<http://shuang.ca>

# 程序说明

本程序使用Golang开发，未开发完全，程序有BUG，运行一段时间后会出现崩溃现象，现仅供测试，请谨慎使用。

# 安装说明

* 安装Golang
* 安装Git
* 安装MongoDB

## 获取所使用的第三方库

```bash
go get github.com/astaxie/beego
go get github.com/beego/i18n
go get github.com/zeebo/bencode
go get gopkg.in/mgo.v2
go get github.com/wangbin/jiebago
```

## 配置conf/app.conf

```
appname = SCDht # 项目名称, 无需理会
httpport = 80 # 运行端口
runmode = dev # 开发模式

showmsg = true # 是否输出信息

dbhost = 127.0.0.1 # MongoDB连接地址
dbport = 27017 # MongoDB连接端口
dbname = SCDht # MongoDB数据库名
dbuser = # MongoDB连接用户名
dbpass = # MongoDB连接密码

cnhotlist = # 简体中文版首页推荐列表, 以 | 分割
```

## 编译运行

进入SCDht源码目录，运行如下命令

```bash
go build
./SCDht
```

# 授权方式

本程序遵循MIT授权

# 程序反馈

<http://shuang.ca>