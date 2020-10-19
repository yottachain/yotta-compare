# yotta-compare
对账数据生成服务，该服务会以固定时间间隔从SN获取改时间段内全部分片信息并按矿机分类后上传至腾讯COS存储服务，矿机可直接从COS上下载数据用于对账

## 1. 部署与配置：
在项目的main目录下编译：
```
$ go build -o yotta-compare
```
配置文件为`yotta-compare.yaml`（项目的main目录下有示例），默认可以放在`home`目录或重建程序同目录下，各配置项说明如下：
```
#服务所连接的数据库URL
mongodb-url: "mongodb://127.0.0.1:27017/?connect=direct"
#数据库名
db-name: "compare"
#全部SN的同步服务地址
all-sync-urls:
  - "http://192.168.36.132:8051"
  - "http://192.168.36.132:8052"
  - "http://192.168.36.132:8053"
#起始时间，从该时间开始生成对账数据，格式为UNIX时间戳
start-time: 1601387400
#以该时间间隔生成对账文件，单位为秒
time-range: 600
#程序出错或没有数据可获取时的等待时间，单位为秒
wait-time: 30
#与当前时间相差该值的时间段内数据不用于生成对账文件，防止数据一致性问题，单位为秒
skip-time: 300
#COS相关配置
cos:
  #COS连接协议，默认为https
  schema: "https"
  #COS连接域名
  domain: "cos.ap-beijing.myqcloud.com"
  #存储桶名
  bucket-name: "compare-1258989317"
  #密钥ID
  secret-id: "xxx"
  #密钥xxx
  secret-key: "xxx"
#日志设置
logger:
  #日志输出类型：stdout为输出到标准输出流，file为输出到文件，默认为stdout，此时只有level属性起作用，其他属性会被忽略
  output: "file"
  #日志路径，默认值为./rebuilder.log，仅在output=file时有效
  file-path: "./compare.log"
  #日志拆分间隔时间，默认为24（小时），仅在output=file时有效
  rotation-time: 24
  #日志最大保留时间，默认为240（10天），仅在output=file时有效
  max-age: 240
  #日志输出等级，默认为Info
  level: "Info"
```

# 2. 启动服务
配置文件设置完毕后执行以下命令启动：
```
$ nohup ./yotta-compare &
```

# 3. 对账数据下载方式
对账文件都以`<矿机ID>_<时间戳>`的格式存放在COS上，时间戳表示该文件所对应对账数据从何时间点开始，比如某矿机ID为12，则第一个对账文件为`12_0`（所有矿机的第一个对账文件时间戳均为0），可通过该文件的标签获取下一个文件的文件名，标签的key为`next`，例如`12_0`的下一个文件为`12_1601388600`，则文件`12_0`存在key为`next`值为`12_1601388600`的标签，可根据该值找到后续的对账文件，如果标签不存在，说明暂时没有后续的对账文件被生成，程序应该等待一段时间后重新获取标签。