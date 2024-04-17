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
[betaBot](https://t.me/unclebetabot)

# 项目结构
```angular2html
.internal       # 后端代码
├───exts──────  # 组件
├───utils─────  # 通用工具
├───router────  # 路由
├───models────  # 数据库模型
├───handlers──  # handlers
└───services──  # services
.env        # 配置文件
templates   # telegram回复模板
cmd         # 程序入口
```

# 配置文件 
# 文件目录

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

# 安全说明：
    1 .env路径下的文件包含敏感信息，请不要暴漏给外部（不要把该文件夹放到nginx网站目录下）
    2 请对自己的信息负责，保护好自己的tg账号，不要使用第三方客户端；登录管理页面后及时退出；如数据库中有钱包密钥，谨慎使用导出钱包功能

# Docker运行方式(推荐)
参考 [umfaka/tgfaka_release](https://github.com/umfaka/tgfaka_release)

# 使用方法
- 配置`conf`下的配置文件`config.ini`
- 把的程序放到与`conf`、`templates`等同一目录下，直接运行程序
- 正常启动程序后，使用管理员账号回复机器人`/login`即可生成一次性登录地址（配置文件填写域名后该登录链接便会附带域名）

## 程序运行参数
    --port 端口号 默认 8082

## Nginx反向代理配置(端口写自己的)
    location / {
        proxy_pass http://127.0.0.1:8082;
    }
    
## TG回复模板在templates中修改

# 有问题反馈
在使用中有任何问题，欢迎反馈
* TG开发频道: [@DawenDev](https://t.me/DawenDev)

## 打赏
如果该项目对您有所帮助，希望可以请我喝一杯咖啡☕️
Usdt(trc20)打赏地址: 
```
TQ17mbGbkjx3sdfsBR1SqmpUsRTyD8XHW3
```

## 灵感来自以下的项目

* [epusdt](https://github.com/assimon/epusdt)
