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

# 安全说明：
    1 .env路径下的文件包含敏感信息，请不要暴漏给外部（不要把该文件夹放到nginx网站目录下）
    2 请对自己的信息负责，保护好自己的tg账号，不要使用第三方客户端；登录管理页面后及时退出；如数据库中有钱包密钥，谨慎使用导出钱包功能


# 使用方法
- 主程序可以自己clone下来编译，也可以直接下载编译好的版本（linux需要给权限运行`chmod +x ./crypto_tg_faka_linux`），请运行在开启SSL的网络环境下（不加SSL会登录失败，因为cookie是设置在https上的）
- 后台页面在[`build`](https://github.com/AnnonaOrg/tgfaka/releases/tag/release)压缩包的`wwwroot`文件夹里面，将`wwwroot`目录下的文件放到nginx网站根目录，并设置反向代理指向程序运行端口
- 配置.env下的配置文件`config.ini`
- 把的程序放到与`.env`、`templates`等同一目录下，直接运行程序
- 正常启动程序后，使用管理员账号回复机器人`/login`即可生成一次性登录地址（配置文件填写域名后该登录链接便会附带域名）

# 程序运行参数
    --port 端口号 默认8082

# Nginx反向代理配置(端口写自己的)
    location ~ ^/(api) {
        proxy_pass http://127.0.0.1:8082;
    }
    
## TG回复模板在templates中修改

# 有问题反馈
在使用中有任何问题，欢迎反馈给我，可以用以下联系方式跟我交流
* tg: [@DawenDev](https://t.me/DawenDev)
* 
## 灵感来自以下的项目

* [epusdt](https://github.com/assimon/epusdt)
