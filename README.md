# 极简版抖音后端（douyin）

第六届字节跳动青训营后端项目，团队「起名起了3min」，获青训营二等奖。
项目文档：https://vish8y9znlg.feishu.cn/docx/XffIdI4sso6oGNx2yWEc4DV4nrh

本仓库为重构后的微服务版本：Hertz 网关 + Kitex RPC，存储使用 MySQL / Redis / MongoDB，消息队列仅使用 Kafka（视频发布异步流水线），服务注册发现使用 etcd。

## 技术栈

| 分类 | 组件 | 版本 |
| --- | --- | --- |
| 语言 | Go | 1.26（toolchain go1.26.8） |
| HTTP 框架 | CloudWeGo Hertz | v0.10.6 |
| RPC 框架 | CloudWeGo Kitex | v0.16.3 |
| 注册中心 | etcd（registry-etcd） | v3 / contrib v0.3.0 |
| 关系数据库 | MySQL 8.x + GORM | gorm v1.25.3 / driver v1.5.1 |
| 缓存 | Redis + go-redis | v9.0.5 |
| 文档数据库 | MongoDB（私信消息） | driver v1.12.1 |
| 消息队列 | Kafka（segmentio/kafka-go） | v0.4.42 |
| 可观测性 | OpenTelemetry + Prometheus + Zap | obs-opentelemetry v0.3.0 |
| 其他 | JWT、雪花 ID、布隆过滤器、敏感词过滤、Viper 配置 | — |

## 服务与端口

| 服务 | 端口 | 职责 |
| --- | --- | --- |
| `service/api` | **8080** | Hertz HTTP 网关：路由、JWT 鉴权、限流、静态资源、`/metrics` |
| `service/social` | **4002** | 注册/登录/用户信息、关注/粉丝/好友 |
| `service/video` | **4003** | 视频流/发布、作品列表、评论、点赞（含 Kafka 视频发布消费者） |
| `service/message` | **4005** | 好友私信（MongoDB） |

基础设施默认端口：etcd `2379`、MySQL `3306`、Redis `6379`、Kafka `9092`、MongoDB `27017`。

跨域 RPC 仅保留三条必要调用：video → social（作者信息）、social → video（作品/点赞计数）、message → social（好友校验）。

## 目录结构

```
.
├── idl/                # Thrift IDL（结构体定义 + social/videoapp/message 三个 service）
├── kitex_gen/          # kitex 生成代码（make gen 重新生成）
├── service/
│   ├── api/            # HTTP 网关（biz 各业务 handler、mw 鉴权/限流、rpc 下游 client）
│   ├── social/         # 用户 + 关系链 RPC 服务
│   ├── video/          # 视频 + 评论 + 点赞 RPC 服务
│   └── message/        # 私信 RPC 服务
├── dal/                # 数据访问层：mysql / mongo / model
├── mw/                 # 中间件：redis（缓存+限流+token）、kafka、localcache
├── common/             # 公共能力：JWT、雪花ID、布隆、敏感词、OSS/本地存储
├── observability/      # tracing、Prometheus 指标、trace_id 日志
├── config/             # app.yaml（本地）/ app.docker.yaml（容器）
├── constant/           # 服务名端口等常量、业务常量
├── logger/             # Zap 日志初始化
├── scripts/dev.sh      # 宿主机一键启停脚本
├── build/bin/          # 编译产物（git 忽略，由 Makefile 生成）
├── Dockerfile          # 多阶段构建，--build-arg SERVICE 区分四个服务
├── docker-compose.yml  # 基础设施 + 四个服务一键编排
└── Makefile
```

## HTTP 接口

完整接口说明（含参数、响应格式、限流规则）见 [docs/http_api.md](docs/http_api.md)。

所有业务接口以 `/douyin` 为前缀；除标注外，写操作需登录（query 或表单携带 `token`）。

## 本地运行

### 1. 环境准备

- Linux
- Go 1.26+
- MySQL 8.0+、Redis 6.2+、Kafka 3.0+、MongoDB 4.4+、etcd 3.5+
- 本地视频封面截帧依赖 `ffmpeg`（仅在未配置 OSS 时需要）

### 2. 修改配置

编辑 [config/app.yaml](config/app.yaml)，把 MySQL / Redis / Kafka / MongoDB 连接信息改成本机环境。
etcd 地址默认 `127.0.0.1:2379`。

> 服务启动时会通过 GORM AutoMigrate 自动建表（user_login / user_info / video / comments / favorite / user_follow），MySQL 库需提前创建（compose 已自动创建 `doushen` 库）。

### 3. 编译与启动

```bash
# 编译四个服务到 build/bin/
make build-all

# 宿主机一键启停（需已自行启动基础设施）
./scripts/dev.sh start              # 编译并按 social → video → message → api 顺序启动
./scripts/dev.sh status             # 查看端口与运行状态
./scripts/dev.sh logs [服务名]       # 跟踪日志：social|video|message|api
./scripts/dev.sh stop
```

启动成功后访问：http://127.0.0.1:8080/douyin/feed/

## Docker 运行

`Dockerfile` 为四服务共用的多阶段构建文件，通过构建参数区分：

```bash
# 构建单个镜像
docker build --build-arg SERVICE=social -t douyin-social .

# 一键拉起基础设施 + 四个服务（首次会构建镜像）
docker compose up -d --build

docker compose logs -f api
docker compose down          # 停止；加 -v 同时清理数据卷
```

容器使用 [config/app.docker.yaml](config/app.docker.yaml)，基础设施地址为 compose 服务名；video 与 api 通过共享数据卷交换本地视频文件。

## IDL 代码生成

需要安装 kitex 工具后执行：

```bash
make gen     # 等价于按 social.thrift / videoapp.thrift / message.thrift 三次 kitex 生成
make tidy    # go mod tidy
make fmt     # gofmt -w .
make vet     # go vet ./...
```

## 配置与环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `DOUYIN_CONFIG` | `config/app.yaml` | 指定配置文件路径，容器内设为 `config/app.docker.yaml` |
| `DOUYIN_ETCD_ADDR` | `127.0.0.1:2379` | etcd 注册中心地址，容器内设为 `etcd:2379` |

`oss.enabled=false`（或密钥为空）时自动降级为本地文件存储：视频写入配置的 `local_fallback_path`（默认 `./tmp`），由 api 的 `/static/` 对外提供；配置 COS 后则走对象存储 + 数据万象截帧。

## 可观测性

- 每个请求注入 `X-Request-Id`，日志自动带上 request_id / trace_id
- Kitex 服务端与客户端均挂载 OpenTelemetry suite；`observability.collector_addr` 配置后导出链路追踪
- api 网关暴露 Prometheus 指标（HTTP 计数/时延、缓存命中）：`GET /metrics`
