# 仅供参考和学习，勿用于生产环境

# Crypto Telegram Faka
### 一个使用Go开发的加密货币USDT发卡机器人，带web后台


### 项目特点
* 支持实时汇率，固定汇率
* 数据库适用pgsql或sqlite（因数据库使用频繁，推荐适用pgsql）
* golang方便部署，免去环境配置
* 支持导入/导出钱包
* 每个Telegram用户只能开一个订单，重复开启订单将自动删除

### 收款钱包分为两种模式
    1.任意金额，每一个钱包只能处理一个一个订单，钱包直至订单结束前都处于解锁状态，可以识别多次或超额支付的情况
    2.小数点尾数，每一个钱包可以处理多个订单，钱包通过微小的金额增量步长与订单绑定，只有准确支付指定金额才能够被识别(如1USDT，会使用1.0001,1.0002...通过不同的金额识别不同的订单,参考epusdt)


# DEMO

# 使用方法

修改`conf/config.ini`中的配置
然后运行以下指令
```
sudo docker-compose up -d
```

## 文件目录

```angular2html
.
├── README.md
├── _init_env_sh   			# debian 系统 安装docker脚本
├── _start_sh   				# 启动脚本
├── conf
│   ├── config.ini   		# 配置文件 必填配置
│   ├── config.simple.ini 	# 配置模版 无需更改
│   └── db_config.ini 		# 数据库配置文件 无需更改
├── docker-compose.yml 		# docker配置文件 无需更改
└── nginx_site.conf 			# nginx配置文件 无需更改
```


## 有问题反馈
在使用中有任何问题，欢迎反馈
开发机器人频道: [@umfaka](https://t.me/umfaka)

# 打赏
如果该项目对您有所帮助，希望可以请我喝一杯咖啡☕️
Usdt(trc20)打赏地址: 
```
TYsBL3pvwzS6PPi9rYCR4n3WixTbWUya8J
```



## 灵感来自以下的项目

* [epusdt](https://github.com/assimon/epusdt)
