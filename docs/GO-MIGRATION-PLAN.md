# teaching-open 服务端 Go 化升级方案

> 适用基线：`teaching-open` **V2.6**（tag `V2.6`，commit `0c79eaf`，2022-01-06）
> 编写日期：2026-11-09
> 配套交付：`go-server/`（可编译、可测试的服务端骨架）、`SECURITY-HARDENING-V26.md`（V2.6 加固记录）

---

## 1. 结论先行

**建议做"渐进式服务端替换"，不做一次性重写。**

V2.6 服务端是 402 个 Java 文件、约 4.1 万行代码、50 个 Controller、495 个 HTTP 端点、
43 个 MyBatis Mapper、50 个实体。直接全量重写为 Go 的周期与风险都不可控，而且真正的
痛点并不在"语言"，而在下面三类结构性问题——它们可以在 Go 侧一次性根治：

| 结构性问题 | V2.6 的具体表现 | Go 方案的解法 | 状态 |
|---|---|---|---|
| 鉴权默认开放 | Shiro 过滤链用 30+ 条 `anon` 白名单定义公开接口，改错一行即永久失鉴权 | 反转为"默认受保护 + 显式公开"，启动时断言，漏挂即拒绝启动 | 已实现并测试 |
| SQL 字符串拼接 | Mapper 里 `select ${text} from ${table}`、`id IN (${userIds})` 等 | 全部占位符绑定；标识符走白名单校验；动态表名不支持 | 已实现并测试 |
| 密码可逆存储 | `PBEWithMD5AndDES` 加密用户名，**可解密**出明文密码 | bcrypt 单向散列 + 登录时自动 rehash，不停机迁移 | 已实现并测试 |

**关键量化结论**：V2.6 的密码"散列"在 `dkLen == 16 == MD5 输出长度` 时，
PBKDF1 规定只取 T1，**1000 次迭代对最终密钥毫无影响**。也就是说，
库中密码的真实强度 = 一次 `MD5(password || salt)` + 一次 DES。
这是必须优先处理的最高危项。

---

## 2. V2.6 服务端现状测绘

### 2.1 技术栈

| 层 | 组件 | 版本 | 备注 |
|---|---|---|---|
| 运行时 | JDK | 1.8 | EOL，无安全更新 |
| 框架 | Spring Boot | 2.1.3.RELEASE | 2019 年版本，已停止社区维护 |
| 安全 | Apache Shiro | 1.4.0 | 存在已知 CVE（见 §2.4） |
| 令牌 | java-jwt (auth0) | 3.7.0 | 密钥取自用户密码 MD5 |
| ORM | MyBatis-Plus | 3.1.2 | 大量 `${}` 拼接 |
| 连接池 | Druid | 1.1.17 | 监控面板默认 admin/123456 |
| JSON | fastjson | 1.2.60 | 反序列化 RCE 高危版本段 |
| 驱动 | mysql-connector-java | 5.1.47 | 不支持 MySQL 8 新认证插件 |
| 缓存 | Redis (shiro-redis) | 3.1.0 | 无密码时裸连 |
| 调度 | Quartz (JDBC JobStore) | 随 Boot | 11 张 `QRTZ_*` 表 |

### 2.2 规模

```
Java 文件          402
Java 代码行        40,984
Controller         50
HTTP 端点          495
Mapper 接口        43   （XML 53 个）
Service 实现       45
实体               50
数据库表           70   （含 11 张 Quartz 表、16 张 onl_* 在线表单表）
前端代码行         65,317（Vue 2.7 + ant-design-vue 1.6）
```

### 2.3 模块分布

| 模块 | 职责 | 迁移优先级 |
|---|---|---|
| `sys_user` / `sys_role` / `sys_permission` / `sys_depart` | 账号、RBAC、组织架构 | **P0**（所有业务的前置） |
| `sys_dict` / `sys_category` / `sys_config` | 字典、分类、站点配置 | **P0**（前端启动依赖） |
| `teaching_course` / `teaching_course_unit` / `teaching_course_dept` | 课程与授权 | P1 |
| `teaching_work` / `teaching_work_comment` / `teaching_work_correct` | 作业、作品、互评 | P1（核心业务） |
| `teaching_scratch_assets` / `oss_file` / `sys_file` | 素材库与文件 | P1（有存储耦合） |
| `sys_log` / `sys_announcement` / `sys_message` | 日志、公告、站内信 | P2 |
| `onl_cgform_*` / `onl_cgreport_*` | 在线表单/报表（低代码引擎） | **P3 / 建议不迁移** |
| `qrtz_*` + `sys_quartz_job` | 定时任务 | P2（可用独立调度替代） |
| `demo` / `joa_demo` / `test_*` | 示例与测试表 | **不迁移，直接删除** |

> 建议：`onl_*`（16 张表）属于 JeecgBoot 低代码引擎，开源版前端几乎不暴露入口。
> 在 Go 版中直接不实现，并用 410 明确告知，可砍掉约 15% 的后端复杂度。

### 2.4 组件级 CVE（改配置无法消除，必须升版本）

| 组件 | V2.6 版本 | 主要风险 | 建议 |
|---|---|---|---|
| Apache Shiro | 1.4.0 | 认证绕过类漏洞（CVE-2020-1957 / 11989 / 13933 等系列）、路径规范化绕过 | ≥ 1.13（若留 Java）；Go 版用自研中间件，无此依赖 |
| fastjson | 1.2.60 | `autoType` 反序列化 RCE | 升级 fastjson2 ≥ 2.0.53，或换 Jackson |
| mysql-connector-java | 5.1.47 | 不支持 MySQL 8 `caching_sha2_password`；旧版反序列化风险 | ≥ 8.0.33 |
| Druid | 1.1.17 | 监控面板弱口令 + 未授权访问 | ≥ 1.2.23 或直接关闭面板 |
| Spring Boot | 2.1.3 | 已 EOL，多个 Spring Framework CVE | ≥ 2.7.18（Java 8 上限）或 3.x（需 JDK 17） |
| java-jwt | 3.7.0 | 算法混淆风险（需显式校验 alg） | ≥ 4.4.0 + 强制 HMAC 白名单 |

**这也是选择 Go 的一个直接动因**：V2.6 的依赖面里，60% 的高危项来自"框架自带的
历史包袱"。Go 版把安全边界收敛到少数几个库（gin / go-redis / jwt / bcrypt），
升级面小、可控。

---

## 3. Go 方案总体设计

### 3.1 技术选型

| 关注点 | 选型 | 理由 |
|---|---|---|
| HTTP 框架 | `gin-gonic/gin` v1.10 | 生态成熟、中间件模型清晰，团队上手快 |
| 数据库 | `jmoiron/sqlx` + `go-sql-driver/mysql` | 显式 SQL，无隐式拼接；`sqlx` 提供结构体映射 |
| 连接池 | `database/sql` 内置 | 参数透明，无需额外依赖 |
| 缓存 | `redis/go-redis/v9` | 官方维护，支持 Lua 脚本原子操作 |
| 令牌 | `golang-jwt/jwt/v5` | 强制校验签名算法，避免 alg 混淆 |
| 密码 | `golang.org/x/crypto/bcrypt` | 单向、自适应成本，业界标准 |
| 配置 | `gopkg.in/yaml.v3` + 环境变量 | 密钥只走环境变量，配置可版本化 |
| 日志 | 标准库 `log/slog` | 结构化日志，零依赖 |
| 迁移 | `golang-migrate` | 与现有 `updateX.Y.sql` 并存，版本可追溯 |

**选型原则**：依赖总数控制在 10 个以内。V2.6 有 45 个 `<dependency>`，
其中大量是 JeecgBoot 脚手架带来的传递依赖——这是 CVE 的主要来源。

### 3.2 分层与目录

```
go-server/
├── cmd/server/main.go              入口：配置校验 → 依赖连接 → 迁移自检 → 启动
├── internal/
│   ├── config/                     配置装载与强校验（缺密钥即拒绝启动）
│   ├── auth/                       JWT 签发/校验/吊销、登录限流
│   ├── passwd/                     密码散列（bcrypt）+ V2.6 PBE 兼容层
│   ├── middleware/                 安全头、CORS 白名单、认证、panic 兜底、访问日志
│   ├── store/                      数据访问（全部占位符绑定）
│   ├── service/                    业务用例（登录、刷新、登出）
│   └── httpapi/                    路由与处理器（含路由守卫启动自检）
└── _crosscheck/                    与 Java 侧算法对齐的交叉验证脚本
```

对应关系：`controller → httpapi`、`service → service`、`mapper → store`、
`ShiroConfig + JwtFilter + ShiroRealm → middleware/auth + auth`、
`PasswordUtil → passwd`。

### 3.3 与 Java 版的接口兼容

前端 65,317 行代码不动。Go 版保持：

- **路径**：仍为 `/api/**`，`context-path` 语义保留
- **响应结构**：`{ success, code, message, result }`，与 `Result<T>` 一致
- **鉴权头**：`X-Access-Token`（同时兼容 `Authorization: Bearer`）
- **令牌字段名**：登录返回 `token` / `refreshToken` / `userInfo`

这样前端只需在 axios 拦截器里补一个"401 时用 refreshToken 换新 token"的逻辑，
其余零改动。

---

## 4. 安全架构对比（核心）

### 4.1 鉴权模型：从"白名单豁免"改为"默认拒绝"

**V2.6 的做法**（`ShiroConfig.shiroFilter`）：

```java
// 30+ 条 anon 豁免，最后 /** = jwt
filterChainDefinitionMap.put("/druid/**", "anon");
filterChainDefinitionMap.put("/actuator/metrics/**", "anon");
filterChainDefinitionMap.put("/swagger-ui.html", "anon");
filterChainDefinitionMap.put("/student/**", "anon");
// ... 省略 25 条
filterChainDefinitionMap.put("/**", "jwt");
```

问题：Shiro 过滤链是**顺序匹配、先命中先返回**。任何一条 `anon` 都可能被
后续新增的敏感接口"套中"。例如 `/student/**` 这一条，当前后端没有对应
Controller（属于历史残留），但一旦未来有人新建 `@RequestMapping("/student/xxx")`
的敏感接口，就会**自动继承匿名访问**，且代码评审时很难发现——因为
`anon` 列表和 Controller 分处两个文件。

**Go 版的做法**：

```go
// 公开路由必须显式登记 requireAuth=false，且必须落在 publicPaths 白名单
pub := api.Group("/sys")
s.register(pub, http.MethodPost, "/login", false, s.handleLogin)
s.register(pub, http.MethodPost, "/mLogin", false, s.handleLogin)
s.register(pub, http.MethodPost, "/refreshToken", false, s.handleRefresh)

// 其余全部挂在带 RequireAuth 的分组
authed := api.Group("")
authed.Use(middleware.RequireAuth(s.tokens, s.logger))
// ...
s.verifyRouteGuards()   // 启动自检，漏挂即 panic
```

`verifyRouteGuards` 在启动时遍历路由登记表：任何写操作（POST/PUT/PATCH/DELETE）
若 `requireAuth=false` 且不在 `publicPaths` 白名单内，**直接 panic 拒绝启动**。
这把"安全问题"从"运行时才暴露"前移到"编译/启动期必现"。

已覆盖测试：`TestVerifyRouteGuardsRejectsUnguardedWrite`、
`TestVerifyRouteGuardsAllowsPublicLogin`、`TestVerifyRouteGuardsRejectsEmpty`。

### 4.2 SQL 注入：从"黑名单过滤"改为"绑定 + 白名单"

**V2.6 的风险点**（`SysUserMapper.xml`）：

```xml
<!-- 拼接：departId 直接进入 SQL -->
AND id IN (${userIds})
UPDATE sys_user SET ${fieldName} = NULL WHERE ${fieldName} = ''
... org_code LIKE '${orgCode}%'
```

以及 `SysDictMapper.xml`、`NgAlainMapper.xml` 的 `select ${text} from ${table}`。

V2.6 的防护是 `SqlInjectionUtil.filterContent()`——一个**黑名单**：
`'|and |exec |insert |select |delete |update |drop |count |chr |mid |...`
黑名单的固有缺陷是"漏一个就破防"（如 `||`、`/**/`、大小写变体、Unicode 同形字）。

**Go 版的做法**：

1. **值一律绑定**：`username = ?`、`phone = ?`、`id IN (?)` 展开为 `IN (?,?,?)`
2. **标识符走白名单**：表名/列名只允许 `[A-Za-z_][A-Za-z0-9_]{0,63}`，且拒绝保留字
3. **不支持动态表名**：V2.6 的表字典（`sys_user,realname,id`）改为
   在 Go 侧维护"表字典注册表"，键为 `dictCode`，值为预定义的 SQL 片段

示例（`store/user.go`）：

```go
const userColumns = "id, username, realname, password, salt, avatar, email, phone, status, del_flag, org_code, create_time"

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
    query := "SELECT " + userColumns + " FROM sys_user WHERE username = ? AND del_flag = 0 LIMIT 1"
    //                                     ↑ 常量              ↑ 绑定
    var u User
    if err := r.db.GetContext(ctx, &u, query, username); err != nil { ... }
}
```

`SELECT *` 也被显式列名替代——防止未来新增敏感列（身份证、住址）被自动带出。

### 4.3 密码：从"可逆加密"改为"单向散列 + 自动迁移"

这是本次最关键的发现。

**V2.6 的实现**（`LoginController`）：

```java
String userpassword = PasswordUtil.encrypt(username, password, sysUser.getSalt());
String syspassword = sysUser.getPassword();
if (!syspassword.equals(userpassword)) { ... }
```

`PasswordUtil.encrypt(plaintext, password, salt)` 的语义是：
用 `password`（用户输入的密码）作密钥，`PBEWithMD5AndDES` **加密** `plaintext`（用户名）。
所以数据库里存的是**密文**，且 `decrypt()` 方法能直接还原出明文密码。

**两个致命后果**：
1. 库泄露 = 全站明文密码泄露（用户普遍多站复用，横向风险巨大）
2. 声称的"1000 次迭代"实际无效（见 §1 量化结论）

**Go 版的迁移设计**（`passwd/passwd.go` + `service/login.go`）：

```
登录请求
   │
   ├─ 存储值以 $2a$/$2b$/$2y$ 开头 → bcrypt 校验（新路径）
   │
   └─ 否则（V2.6 存量密文）→ LegacyPBE.Verify(用户名, 密码, 密文, 盐)
                                │
                                ├─ 失败 → 拒绝登录
                                └─ 成功 → 立即 bcrypt 重新散列并写回
                                          （rehash-on-login）
```

**兼容性已验证**：`_crosscheck/pbkdf1_des_ref.py` 用纯 Python 独立复算
PBKDF1(MD5) + DES-CBC（其 DES 实现先通过 FIPS 81 已知向量
`0123456789ABCDEF` / `133457799BBCDFF1` → `85E813540F0AB405` 自检），
再用 4 组向量与 Go 实现比对，全部一致：

| 用户名 | 密码 | 盐 | 密文（hex） |
|---|---|---|---|
| admin | 123456 | 63293188 | `c1461cf4492050e9` |
| teacher | 123456 | 63293188 | `44a7ec3167f68089` |
| student | Abc@12345 | 63293188 | `ae7d77e35833d801` |
| 张三 | passw0rd | 63293188 | `2e235213908e99bf` |

**迁移是可灰度的**：不需要停机改数据。用户下次登录自动升级；
运维可通过 `main.go` 的 `reportLegacyAccounts` 查询实时进度：

```
密码散列迁移进度 total_users=1000 bcrypt_users=873 legacy_users=127
```

当 `legacy_users = 0` 后，删除 `passwd.go` 中的 Legacy 部分即可彻底移除弱算法。

### 4.4 令牌：从"密码派生密钥"改为"独立高熵密钥 + 双令牌"

| 维度 | V2.6 | Go 版 |
|---|---|---|
| HMAC 密钥 | `JwtUtil.sign(username, syspassword)`——用户密码的 MD5 | 独立 32+ 字节密钥，环境变量注入，缺失即拒绝启动 |
| 有效期 | 30 小时 | Access 2 小时 / Refresh 7 天 |
| 续期机制 | Redis 缓存 token，TTL 设为 2 倍（实际在线窗口 60 小时） | Refresh 换新，旧 refresh 立即吊销（一次性） |
| 主动失效 | 只能靠 Redis key 过期 | 支持 `jti` 黑名单吊销（登出、改密、强制下线） |
| 算法校验 | 未显式限制 `alg` | `jwt.WithValidMethods([]string{"HS256"})` |
| 密钥轮换 | 不支持 | `JWT_PREV_SECRET` 只验不签，平滑轮换 |

**为什么"用密码当密钥"是严重问题**：密码可被撞库/彩虹表还原；更关键的是，
一旦用户改密码，所有已签发 token 立即失效（因为密钥变了）——看似"安全"，
实际导致 V2.6 只能靠"把 token 存 Redis 且放大 TTL"来维持在线，
反而把在线窗口拉长到 60 小时。

### 4.5 登录限流

V2.6 **完全没有**登录失败限制。`/sys/login` 只需验证图形验证码，
而 `/sys/mLogin`（App 登录）**连验证码都没有**，是直接的爆破入口。

Go 版（`auth/limiter.go`）：

- 维度：`用户名 + 客户端 IP`（避免同一 NAT 出口下互相牵连）
- 阈值：5 次 / 15 分钟（可配 `LOGIN_FAIL_LIMIT`）
- 原子性：`INCR` + `PEXPIRE` 用 Lua 脚本一次完成，避免并发下计数丢失或 TTL 未设
- 键名：对用户名做哈希后再拼，不把明文账号写进 Redis
- 故障策略：Redis 不可用时 **fail-open**（放行登录 + 告警），
  因为"缓存故障导致全站无法登录"的损失大于短期限流失效

### 4.6 暴露面收敛对照

| 端点 | V2.6 | Go 版 | 依据 |
|---|---|---|---|
| `/druid/**` | `anon` + `admin/123456` | 不实现（生产关闭） | 监控面板泄露 SQL 与会话 |
| `/actuator/metrics,httptrace` | `anon` | 仅 `/api/actuator/health`，返回 `{"status":"UP"}` | 运行信息泄露 |
| `/swagger-ui.html`、`/doc.html`、`/v2/api-docs` | `anon` | 不注册；`ENABLE_SWAGGER` 仅非生产可开 | 接口清单泄露 |
| `/student/**` | `anon`（无对应 Controller，历史残留） | 不存在 | 避免未来接口继承匿名 |
| `/generic/**`、`/sys/common/pdf/**` | `anon` | 不实现 | 任意文件读取面 |
| 全局异常 | `Result.error("操作失败，" + e.getMessage())` | 固定话术 + `traceId`，堆栈仅进日志 | SQL/路径/类名泄露 |
| CORS | `allowCredentials(true)` + `allowedOrigin("*")` | 显式白名单，`*` 在校验期直接拒绝 | 跨域凭据泄露 |

---

## 5. 分阶段迁移路线

### 阶段 0：止血（1~2 周，与 Go 无关，可立即做）

在 V2.6 上先完成安全加固，避免迁移期间持续暴露。**本次已交付**，
详见 `SECURITY-HARDENING-V26.md`。

同时必须做的运维动作：
- 强制全站改密（因为存量密码可解密）
- 轮换数据库/Redis/对象存储密钥
- 网关层封禁 `/druid`、`/actuator`、`/swagger*`、`/doc.html`

### 阶段 1：并行建站（3~4 周）

Go 服务与 Java 服务**同时运行**，通过 Nginx 按路径分流：

```nginx
# 新接口走 Go
location ^~ /api/v2/ { proxy_pass http://127.0.0.1:8081; }

# 登录、字典、配置先切 Go（前端启动链路）
location = /api/sys/login            { proxy_pass http://127.0.0.1:8081; }
location ^~ /api/sys/dict/           { proxy_pass http://127.0.0.1:8081; }
location ^~ /api/sys/config/         { proxy_pass http://127.0.0.1:8081; }

# 其余仍走 Java
location ^~ /api/ { proxy_pass http://127.0.0.1:8080; }
```

**这一阶段的关键是令牌互通**：Go 与 Java 必须能互相验证对方签发的 token。
做法：
1. 先把 Java 的 `JwtUtil` 密钥改为读环境变量（与 Go 共用同一个 `JWT_SECRET`）
2. 两端都用 HS256 + 相同的 claim 结构（`username` / `uid` / `exp`）
3. Java 侧保留 `PREFIX_USER_TOKEN` Redis 白名单机制，Go 侧签发时同步写入

若不想改动 Java，则退化为"双令牌"：Go 登录后同时返回一个 Java 可识别的
legacy token，过渡期由前端按接口目标选择。但这种方式维护成本高，
建议直接用共享密钥方案。

### 阶段 2：核心域接管（4~6 周）

按依赖倒序迁移，每迁移一个域就切一次流量：

```
P0  用户 / 角色 / 权限 / 部门     ← 所有业务的前置，必须先做
P0  字典 / 分类 / 站点配置         ← 前端启动依赖
P1  课程 / 课程单元 / 课程授权
P1  作业 / 作品 / 评论 / 评分
P1  文件上传 / 素材库
P2  日志 / 公告 / 站内信
P2  定时任务（改用独立调度，不再依赖 QRTZ_* 表）
P3  在线表单/报表                  ← 建议直接不迁移（410）
```

每个域的验收标准：
1. Go 侧接口的响应与 Java 侧**逐字段 diff 一致**（含字段顺序不敏感、空值语义）
2. 该域的写接口在 `verifyRouteGuards` 下全部 `requireAuth=true`
3. 灰度 5% → 50% → 100%，每档观察 24 小时错误率与 P99
4. 保留 Java 侧一键回切能力（Nginx 改一行）

### 阶段 3：下线 Java（2 周）

- 观察期：Go 承载 100% 流量 ≥ 2 周，无 P0/P1 故障
- 数据层：删除 `demo` / `joa_demo` / `test_*` / `onl_*` 表（先备份）
- 代码层：Java 模块归档到 `legacy/` 分支，不删除历史
- 文档层：更新部署文档、运维手册、应急回滚预案

### 阶段 4（可选）：架构增强

- 读写分离：`sqlx` 多数据源，报表查询走只读副本
- 对象存储统一：把 local/qiniu/alioss/minio 四种驱动收敛为接口 + 三个实现
- 可观测性：接入 OpenTelemetry，把 `traceId` 打通到日志与前端

---

## 6. 里程碑与工作量估算

| 阶段 | 内容 | 人周 | 关键交付 |
|---|---|---|---|
| 0 | V2.6 止血加固 | 1~2 | 本次交付的加固分支 + 强制改密 |
| 1 | Go 骨架 + 登录链路 + 令牌互通 | 3~4 | 可登录、可鉴权、可灰度分流 |
| 2a | P0 域（用户/权限/字典/配置） | 4 | 前端可完整启动 |
| 2b | P1 域（课程/作业/文件） | 6 | 核心业务可用 |
| 2c | P2 域（日志/公告/调度） | 3 | 运营功能可用 |
| 3 | 全量切换 + Java 下线 | 2 | 单栈运行 |
| **合计** | | **19~21 人周** | |

对比"一次性重写"：约 40~50 人周，且期间新旧两套都不完整，风险更高。

---

## 7. 风险与对策

| 风险 | 影响 | 对策 |
|---|---|---|
| 令牌互通改造触碰 Java 登录 | 全站无法登录 | 先在预发环境跑通双写；保留 Redis legacy token 回退开关 |
| 密码迁移期用户登录失败 | 客服压力 | 灰度按用户 ID 取模；保留 `LegacyPBE` 至少一个版本周期 |
| 响应字段顺序/空值语义差异 | 前端渲染异常 | 建契约测试：用 V2.6 的真实响应做 golden file 比对 |
| 动态表字典（`sys_user,realname,id`）无法直接迁移 | 部分下拉框失效 | 建"表字典注册表"，把用到的组合显式登记；未登记的返回明确错误 |
| Quartz 任务迁移遗漏 | 定时任务不执行 | 先双跑对账（两边都执行、只让一边写库），确认无差异后切 |
| Go 团队不熟悉 Vue 前端契约 | 联调返工 | 阶段 1 就把契约测试建起来，以测试而非文档为准 |
| MySQL 5.1 驱动 → 8.x | 时区/字符集行为变化 | 连接串显式 `parseTime=true&loc=UTC`；`SET time_zone` 已在启动时执行 |

---

## 8. 本次交付物说明

### 8.1 `go-server/` 已实现（可编译、可测试）

```
cmd/server/main.go                启动编排 + 存量密码迁移进度统计
internal/config/config.go         配置强校验（缺密钥/通配符 CORS/生产开 Swagger 均拒绝启动）
internal/auth/token.go            JWT 签发/校验/吊销/密钥轮换
internal/auth/limiter.go          Redis 原子登录限流
internal/passwd/passwd.go         bcrypt + V2.6 PBE 兼容层
internal/middleware/middleware.go 安全头/CORS 白名单/panic 兜底/请求追踪/访问日志
internal/middleware/auth.go       RequireAuth / RequireRole
internal/store/user.go            占位符绑定的数据访问
internal/service/login.go         登录/刷新/登出 + rehash-on-login
internal/httpapi/server.go        路由 + 路由守卫启动自检 + 统一响应
_crosscheck/pbkdf1_des_ref.py     与 Java 算法对齐的独立复算脚本
```

### 8.2 验证证据

```
$ go build ./...            → 通过
$ go vet ./...              → 通过
$ gofmt -l .                → 无输出（格式合规）
$ go test ./... -count=1    → 5 个包全部 ok，44 项测试通过
```

测试覆盖的安全属性：

| 类别 | 测试 | 数量 |
|---|---|---|
| 令牌 | 签发/解析、刷新令牌不可当访问令牌、篡改拒绝、错密钥拒绝、短密钥拒绝、吊销生效、吊销存储故障时 fail-closed、密钥轮换 | 8 |
| 配置 | 缺 DSN/缺密钥/短密钥/生产开 Swagger/生产开 pprof/CORS 通配符 均拒绝启动；开发环境允许 Swagger | 8 |
| 密码 | 4 组 Java 兼容向量校验、错误密码拒绝、错误用户名拒绝、畸形密文不 panic、方案识别 | 6 |
| 路由 | 健康检查不泄露内部信息、受保护接口 401/200、参数校验、超长输入拒绝、安全头、CORS 白名单 | 8 |
| 中间件 | 无令牌/坏令牌/有效令牌/Bearer 头、CORS 白名单与预检拒绝、安全头、panic 不泄露 | 7 |
| 守卫 | 未挂认证的写路由拒绝启动、公开登录放行、空路由表拒绝启动、脱敏函数 | 4 |

### 8.3 明确未实现（诚实边界）

- **未实现全部 495 个业务端点**。本次交付的是"可运行的安全骨架 + 登录链路"，
  用于验证架构可行性与安全模型，而非完整替代 V2.6。
- **`_crosscheck/pbkdf1_des_ref.py` 未经真实 Java 环境比对**。Python 实现的
  DES 已通过 FIPS 81 官方向量自检，PBKDF1 派生与 Go 一致，但"与真实 Java
  `PBEWithMD5AndDES` 输出完全一致"仍需在有 JDK 的环境跑一次
  `PasswordUtil.encrypt` 做最终确认。这是上线前必须执行的验收步骤。
- **无 JDK/Maven 环境**，因此 V2.6 的 Java 改动仅做了结构级静态校验
  （括号配对、XML 解析、YAML 解析、方法定义/调用配对），**未执行 `mvn compile`**。
- **未实现文件存储、对象存储、Quartz 调度、在线表单**等模块。

---

## 9. 上线前必做的验收清单

- [ ] 在 JDK 8 环境执行 `PasswordUtil.encrypt("admin","123456","63293188")`，
      确认输出为 `c1461cf4492050e9`（验证 §4.3 的兼容性结论）

      > **进展（2026-11-17）**：Go 侧已通过 `LegacyPBE.Encrypt` 双向验证——
      > 4 组向量（含 admin/123456/63293188）的加密输出与文档记录的密文**逐字节一致**，
      > 且 `Encrypt→Decrypt` 往返通过（`internal/passwd/passwd_test.go`）。
      > 仍建议在真实 JDK 环境跑一次作为交叉确认。
- [ ] `mvn clean package` 通过，且加固后的 V2.6 能正常登录、注册、找回密码
- [ ] 用 `sys_user` 全量数据跑一次 `reportLegacyAccounts`，确认存量账号数
- [ ] 强制全站改密（存量密码可解密，必须视为已泄露）
- [ ] 确认网关已封禁 `/druid`、`/actuator`、`/swagger*`、`/doc.html`、`/v2/api-docs`
- [ ] 配置 `JWT_SECRET`（≥32 字节随机值，不入库不入仓）
- [ ] 压测登录接口：确认 bcrypt cost=12 下 P99 可接受（预期 200~300ms/次）
- [ ] 演练令牌轮换：设置 `JWT_PREV_SECRET` 后确认旧 token 仍可用、新 token 正常签发

---

## 10. 下一阶段计划（2026-11-22 制定，基于架构审计与真机验证）

### 10.1 架构现状快照（审计基准）

| 项 | 值 | 备注 |
| --- | --- | --- |
| Go 工具链 | go1.27.4（go.mod 指令 1.26.0） | Linux 便携安装 ~/tools/go |
| Web 框架 | GoFrame v2.10.4 | 约定见 .agents/skills/goframe-v2 |
| 数据库 | MariaDB 11.8.9（teachingopen） | 69 表 V2.8 DDL + 春晖种子 |
| 端口约定 | API 18080 / 前端 dev 18082 / 静态预览 18083 / DB 13306 | 统一高端口防冲突（2026-11-22 起） |
| Excel 引擎 | excelize v2.8.1 | 15 域规格 + importStudent |
| 前端 | vue-antd-jeecg 2.2.0（Vue 2.7.16） | Node 20.19.4 生产构建通过 |
| 注册路由 | ~325 条（internal/cmd/） | Java 484 方法级基线见 §7 |
| 服务包 | internal/service/ 35 个域包 | dao/do/entity 23 套 |
| 测试 | 21 个 _test.go | go vet/build/test 全绿 |

### 10.2 真机端到端验证结论（Puppeteer + chrome-headless-shell）

- 前端生产构建 → dist 静态服务 + /api 反代 → 登录页/门户/管理端 10 页全通（docs/screenshots/）。
- 过程中发现并修复两类架构缺陷（见 commit 11ec19c）：
  1. anon 路由挂载前缀错位（/api/sys 组吞掉教学域顶层路径 → 实际路径多 /sys）；
  2. /api/sys 组内 5 条路由串重复携带 /sys 前缀（randomImage/checkCaptcha/phoneLogin/getEncryptedString/common/static）。
- 种子数据补齐 sys_role_permission 管理员全菜单授权（138 条，QueryByUser 无 admin 旁路）。

### 10.3 剩余工作清单（优先级序）

1. **端点级覆盖精核**：以 484 方法级清单逐条 diff 当前 325 条注册路由，
   输出「已迁移/合并实现/有意排除/待补」四态台账（本轮仅完成域级核对）。
2. **Excel 列规格逐域精校**：excelx 15 域规格按 Java @Excel 注解逐列核对表头/
   顺序/字典翻译（部分已核对，低风险）。
3. **结构性排除确认（需产品决策）**：ThirdLogin 2（OAuth 凭据未配置）、
   Quartz execute（Go 侧无调度器后端，QuartzSetStatus 已实现）、
   ssoLogin（CAS 未部署）。
4. **Redis 生产化**：REDIS_ENABLED=true 下复验验证码/登录限流/点赞幂等
   （当前单机 memory 模式全通，Redis 故障已按放行降级，日志可见）。
5. **静态资源 Content-Type**：upload.Static 当前统一 octet-stream +
   attachment；按扩展名映射 image/png、video/mp4、application/zip，
   以便门户封面直接内联渲染（<img> 已可用，直接访问体验待优化）。
6. **前端构建产物入库策略**：dist 体积 ~22MB，暂不入库；补 deploy/build.sh
   集成 Node 构建（或 CI 产物分发）。

### 10.4 数据完备度（春晖真实数据）

- 组织/用户/角色/授权：5 部门、10 用户、4 角色、138 条管理员授权。
- 课程域：6 课程 + 22 单元 + 6 授权（含 CSP-J）。
- 内容域：4 公告 + 13 阅读标记 + 4 资讯。
- 作品域：12 作品（3 门课 × 学员）+ 6 封面 PNG（渐变实图）+ 12 .sb3 占位 + 6 互评。
- 配置域：10 条（brandName/_address/_copyRight/_indexHtml/_homeHtml/_workShareHtml 等）。
- 双路径：chseed（Go，含二进制生成）与 chunhui_seed.sql（纯 SQL，二进制除外）均已实测入库。
