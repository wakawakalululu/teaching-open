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
| `MYSQL_LINK` | MySQL 连接串（`mysql:user:pass@tcp(host:port)/db?...`） | 见 config.yaml |
| `REDIS_ENABLED` | 是否启用 Redis（限流 + 令牌吊销）；关闭时两者 fail-open | false |
| `REDIS_ADDR` | Redis 地址（见 config.yaml `redis.default.address`） | 127.0.0.1:6379 |
| `CORS_ALLOWED_ORIGINS` | 跨域白名单（逗号分隔，禁止 `*`） | 空 |
| `FEISHU_WEBHOOK_URL` | 飞书机器人告警地址（账号锁定/DB 故障推送） | 空（禁用） |
| `APP_ENV` | 运行环境标识 | production |

> **Redis 开关说明**：`REDIS_ENABLED=false` 时，登录限流与令牌吊销自动降级为
> fail-open（可用性优先，日志告警）。**生产环境必须设为 `true`**，否则登出/改密
> 无法真正吊销已签发令牌。

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
| GET | `/api/sys/user/info` | 当前用户信息（简化版） | 是 |
| GET | `/api/sys/permission/getUserPermissionByToken` | 路由菜单树 + auth/allAuth 按钮权限（P0-2） | 是 |

> 登录响应（login/mLogin）已含 Java 契约的 `sysAllDictItems`（全量字典，
> 前端缓存后免二次请求）；`userInfo` 输出白名单超集（id/username/realname/
> avatar/orgCode/birthday/sex/userIdentity/status + 脱敏 phone/email）。
| POST | `/api/sys/logout` | 登出（吊销令牌） | 是 |
| PUT | `/api/sys/user/updatePassword` | 自助改密（校验旧密码，仅本人） | 是 |
| PUT | `/api/sys/user/changePassword` | 管理员重置他人密码（需 admin 角色） | 是 |
| GET | `/api/sys/dict/queryAllDictItems` | 全量字典（登录后前端初始化强依赖） | 是 |
| GET | `/api/sys/dict/getDictItems/{dictCode}` | 按字典 Code 取字典项（普通/表字典二态） | 否（对齐 Shiro anon） |
| GET | `/api/sys/dict/getDictText/{dictCode}/{key}` | 按键值取字典文本 | 是 |
| GET | `/api/sys/dict/loadDict/{dictCode}?keyword=` | 表字典异步搜索 | 是 |
| GET | `/api/sys/dict/loadDictItem/{dictCode}?key=` | 表字典回显 | 是 |
| GET | `/api/sys/dict/list` / `treeList` / `getByCode` / `deleteList` | 字典管理页查询 | 是 |
| POST/PUT/DELETE | `/api/sys/dict/add` / `edit` / `delete` / `deleteBatch` | 字典增改删（del_flag 软删 + 回收站） | 是 |
| DELETE | `/api/sys/dict/deletePhysic/{id}` | 物理删除字典并级联清理字典项 | 是 |
| PUT | `/api/sys/dict/back/{id}` | 回收站恢复 | 是 |
| GET | `/api/sys/dict/refleshCache` | 刷新字典缓存（Go 版直查库，兼容占位） | 是 |
| GET | `/api/sys/dictItem/list` | 字典项分页（sort_order 升序） | 是 |
| POST/PUT/DELETE | `/api/sys/dictItem/add` / `edit` / `delete` / `deleteBatch` | 字典项管理 | 是 |
| GET | `/api/sys/category/rootList` / `childList` / `queryById` | 分类字典查询 | 是 |
| GET | `/api/sys/category/loadOne` / `loadTreeChildren` / `loadTreeRoot` / `loadDictItem` / `loadAllData` / `loadTreeData` / `checkCode` | 分类树加载与回显（列名白名单校验） | 是 |
| POST/PUT/DELETE | `/api/sys/category/add` / `edit` / `delete` / `deleteBatch` | 分类增改删（服务端生成 YouBian 流水码） | 是 |
| GET | `/api/sys/config/getConfig?key=` | 按键取配置 | 否（对齐 Shiro anon） |
| GET | `/api/sys/config/getCurrentConfig` | 对外公开配置（登录页加载） | 否（对齐 Shiro anon） |
| GET | `/api/sys/config/getAllConfigList` / `getAllConfigMap` | 配置管理页查询 | 是 |
| POST | `/api/sys/config/saveTenantConfig` | 保存租户配置（需 admin 角色，逐键 upsert） | 是 |

> P0-1 域（字典/分类/配置）契约对齐 Java V2.8 Controller。暂未迁移：Excel
> 导入导出（`exportXls`/`importExcel`）、带过滤条件的表字典（4 段式 dictCode，
> Java 版为原始 SQL 片段拼接，Go 版明确拒绝）。

### 运维工具（`tools/`）

| 工具 | 用途 |
|---|---|
| `probe` | 探测数据库连接、列出表 |
| `importer` | 导入 `teachingopen2.6.sql`（自动关闭外键检查） |
| `verifyseed` | 统计密码方案分布（bcrypt / legacy） |
| `e2e` | 端到端验证（14 步）：登录/rehash/改密/越权拦截/登录响应字典/菜单权限树/字典与分类抽查 |

## 安全要点

1. **默认拒绝鉴权**：所有写操作路由必须显式声明认证，否则启动自检 `panic`。
2. **占位符绑定防注入**：全部 gdb 占位符；表字典的动态表名/列名经
   `^[a-zA-Z0-9_]+$` 白名单校验后才允许拼接（比 Java 版黑名单过滤更严）。
3. **密码迁移双路径**：
   - 被动：存量 PBE 密文登录成功后自动 rehash 为 bcrypt（cost=12）；
   - 主动：任何改密/重置动作直接写 bcrypt（salt 清空），并推送飞书告警。
4. **双令牌 + 吊销**：Access 2h / Refresh 7d，刷新令牌一次性使用。
5. **登录限流**：用户名 + IP 维度，5 次 / 15 分钟，Lua 原子计数。
6. **越权防护**：自助改密只允许本人；管理员重置显式校验 admin 角色
   （Java 版该注解被注释掉，此处补齐）。

## 待办

- P0-1（字典/分类/配置）已迁移，待 sqlpub 真实库回归（登录 → 首页字典/分类下拉）。
- 未实现全部 495 个业务端点（下一批：P0-2 当前用户信息/菜单/权限）。
- 密码兼容向量已通过 Go 侧 `Encrypt`/`Decrypt` 双向验证（与 Java 4 组向量逐字节一致），
  仍建议在真实 JDK 环境跑一次 `PasswordUtil.encrypt` 作为最终确认。
- 文件存储、对象存储、Quartz 调度、在线表单等模块未实现。
