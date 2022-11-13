# icmp-ufw
## 通过icmp包来控制防火墙的打开(ufw / iptables)

# 实现功能
- 热加载
- 远程热更新
- 通过icmp包的size控制防火墙策略
- 支持ufw
- 缓存功能，防止异常退出后策略遗留

# 暂未实现（即将）
- webhook调用
- iptables支持
- 策略时间限制
- 更加完善的日志系统
- 快速部署

# 使用方法
## 1. 安装
```shell
go mod download

go build main.go

./main -h
```
![image](img.png)

## 2. 配置
```yaml
# 配置文件
listen_interface:  #监听接口
  - lo0
firewall_program: ufw  #调用防火墙程序 ufw/iptables
time_out: 3600  #过期时间
webhook_url: ""  #webhook地址
webhook_method: ""  #webhook调用方式 get/post/...
webhook_data: ""  #webhook数据 占位符：{srcip} {allow_port}
webhook_headers:  #webhook请求头
  - ""
hot_update: ""  #热更新地址，返回为本配置文件
auto_reload: false  #自动重载当前配置文件
icmp_ufw_rules:  #防火墙策略
  - size: 56  #icmp包大小
    time_out: 3600  #过期时间
    allow_port: 1-65535  #允许端口
```

## 3. 运行
### 命令行参数
- -h 获取帮助
- -c 配置文件路径(config.yaml)
- -hotUpdate 远程热更新地址
- -autoReload 自动重载当前配置文件
- -timeOut 过期时间

## 4.开发初衷
- 通过icmp包来控制防火墙的打开，可以在不开放端口的情况下，通过icmp包来控制防火墙的打开，从而实现远程控制防火墙的功能
- 咱的数据库服务器设置了白名单 但是作为在欧洲的一台服务器，不开梯子速度慢的要死
- 而梯子又由于一些原因 可能隔一段时间ip就变了，或者切换节点，这时候就无法访问服务器，就很...不舒服
- 所以用了几个小时写了这个demo（目前） 之后会逐步完善
- 就酱～