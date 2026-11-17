# teaching-open GoFrame 服务端

基于 **GoFrame v2.10.x** 重写的服务端，替代原 `go-server/`（Gin 版）与 Java V2.6 服务端。

## 架构

采用 GoFrame 官方最佳实践分层（参考 GoFrame Skills 的 `user-http-service` 示例）：

```
goframe-server/
├── main.go                         入口：装载 MySQL 驱动 + 启动命令
├── manifest/config/config.yaml     框架配置（HTTP/日志/数据库）
├── internal/
│   ├── cmd/cmd.go                  启动编排 + 路由守卫自检（fail-fast）
│   ├── config/config.go            业务配置强校验（缺密钥即拒绝启动）
│   ├── controller/login/           登录/刷新/登出/当前用户处理器
│   ├── service/login/              登录/刷新/登出用例 + rehash-on-login + Redis 吊销
│   ├── dao/                        sys_user 数据访问（gdb ORM）
│   ├── model/do/、model/entity/    DO 对象与实体
│   ├── auth/                       JWT 签发/校验/吊销 + 登录限流
│   ├── passwd/                     bcrypt + V2.6 PBE 兼容层
│   └── middleware/                 安全头/CORS 白名单/认证/panic 兜底
```

## 与 Gin 版的差异

| 维度 | Gin 版（go-server） | GoFrame 版（本目录） |
|---|---|---|
| HTTP 框架 | gin | **ghttp** |
| 配置 | 纯环境变量 | **gcfg**（config.yaml）+ 环境变量 |
| 日志 | log/slog | **glog** |
| ORM | jmoiron/sqlx | **gdb**（dao/do/entity） |
| Redis | go-redis/v9 | **gredis**（`g.Redis()`） |
| 密码/令牌/限流 | 复用 | 复用（框架无关逻辑平移） |

安全模型与业务语义完全保持一致：默认拒绝的鉴权、占位符绑定防注入、
bcrypt rehash-on-login、双令牌 + 吊销、登录限流、CORS 白名单、安全响应头。

## 编译与测试

```bash
# 若网络受限，切换国内代理
export GOPROXY=https://goproxy.cn,direct

cd goframe-server
go build ./...      # 编译
go vet ./...        # 静态检查
go test ./... -count=1   # 测试（含 4 组 Java 密码兼容向量）
```

## 运行

```bash
# 必须注入 ≥32 字节的 JWT 密钥（缺失即拒绝启动）
export JWT_SECRET=$(openssl rand -hex 32)
export MYSQL_DSN="mysql:root:password@tcp(127.0.0.1:3306)/teachingopen?parseTime=true&loc=UTC"
export REDIS_ADDR="127.0.0.1:6379"
export CORS_ALLOWED_ORIGINS="https://your.domain"

go run .
```

### 环境变量

| 变量 | 说明 | 默认 |
|---|---|---|
| `JWT_SECRET` | JWT HMAC 密钥（≥32 字节，必填） | 无（缺省拒绝启动） |
| `JWT_PREV_SECRET` | 密钥轮换期旧密钥（只验不签） | 空 |
| `MYSQL_DSN` | MySQL 连接串 | 见 config.yaml |
| `REDIS_ADDR` | Redis 地址 | 127.0.0.1:6379 |
| `CORS_ALLOWED_ORIGINS` | 跨域白名单（逗号分隔，禁止 `*`） | 空 |
| `APP_ENV` | 运行环境标识 | production |

## 接口契约（与前端零改动对接）

- 路径：`/api/**`
- 响应：`{ success, code, message, result }`
- 鉴权头：`X-Access-Token`（兼容 `Authorization: Bearer`）
- 登录返回：`token` / `refreshToken` / `userInfo`

### 已实现端点

| 方法 | 路径 | 说明 | 认证 |
|---|---|---|---|
| GET | `/api/actuator/health` | 健康检查（仅返回 `{"status":"UP"}`） | 否 |
| POST | `/api/sys/login` | 账号密码登录 | 否 |
| POST | `/api/sys/mLogin` | App 登录（同 login） | 否 |
| POST | `/api/sys/refreshToken` | 刷新令牌换新 | 否 |
| GET | `/api/sys/user/info` | 当前用户信息 | 是 |
| POST | `/api/sys/logout` | 登出（吊销令牌） | 是 |

## 安全要点

1. **默认拒绝鉴权**：所有写操作路由必须显式声明认证，否则启动自检 `panic`。
2. **占位符绑定防注入**：全部 gdb 占位符，表名/列名为常量。
3. **密码迁移**：存量 V2.6 PBE 密文登录成功后自动 rehash 为 bcrypt（cost=12）。
4. **双令牌 + 吊销**：Access 2h / Refresh 7d，刷新令牌一次性使用。
5. **登录限流**：用户名 + IP 维度，5 次 / 15 分钟，Lua 原子计数。

## 待办（与旧 go-server 相同的诚实边界）

- 未实现全部 495 个业务端点（当前为可运行的安全骨架 + 登录链路）。
- 密码兼容向量仍需在真实 JDK 环境跑一次 `PasswordUtil.encrypt` 做最终确认。
- 文件存储、对象存储、Quartz 调度、在线表单等模块未实现。
