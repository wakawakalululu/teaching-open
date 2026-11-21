# 后续推进方向（Roadmap）

> 更新时间：2026-11-21
> 当前基线：`master` = V2.8（上游最新）+ GoFrame 服务端骨架
> 分支状态：`security/v2.6-hardening` 已合并入 `master`，分支差异为 0

---

## 0. 当前进度快照

| 项 | 状态 | 说明 |
|---|---|---|
| Java V2.6 安全加固 | ⛔ 已放弃自研方案 | 上游 master 已有更完整加固（17de6b6 + 4aa5e31），统一采用上游 |
| GoFrame 服务端骨架 | ✅ 完成 | ghttp/gcfg/gdb/gredis/glog 全栈，分层符合官方最佳实践 |
| 登录链路 | ✅ 完成并实测 | sqlpub 真实库验证通过 |
| 密码迁移（DES→bcrypt） | ✅ 双路径完成 | 被动 rehash-on-login + 主动改密/重置 |
| 改密/重置密码 | ✅ 完成 | 自助改密（校验旧密码）+ 管理员重置（校验 admin 角色） |
| 密码迁移进度观测 | ✅ 完成 | 启动日志 total/bcrypt/legacy |
| 算法兼容性验证 | ✅ 双向自证 | Go Encrypt/Decrypt 与 Java 4 组向量逐字节一致 |
| 飞书安全告警 | ✅ 完成 | 账号锁定、DB 故障、改密提醒 |
| Redis 可选降级 | ✅ 完成 | `REDIS_ENABLED` 开关，无 Redis 时 fail-open |
| 端到端验证 | ✅ 14 步全通过 | `tools/e2e`（本地 MariaDB + V2.8 种子数据实测） |
| P0-1 字典/分类/配置 | ✅ 完成并实测 | 36 端点契约对齐 V2.8；字典/表字典/分类真实库抽查通过 |
| P0-2 用户信息/菜单/权限 | ✅ 完成并实测 | 登录响应 sysAllDictItems（40 字典）+ 菜单树 menu=23/auth=3/allAuth=10 |
| P0-3 用户管理 CRUD | ✅ 完成 | 32 端点：核心 CRUD + 角色分配 + 部门关联 + 回收站 + 注册/找回密码（见迁移说明） |
| 全部 495 个业务端点 | 🔄 进行中 | 已迁移 ~80 个（登录链路 + P0-1 + P0-2 + P0-3） |

---

## 1. 立即要做的两件事（本周）

### 1.1 完成存量密码迁移

当前实测：`total_users=3 bcrypt_users=0 legacy_users=3`

**动作**：让 admin / teacher / student 各登录一次，或由管理员重置密码。

**验收**：重启服务看到 `legacy_users=0`，WARN 告警消失。

### 1.2 生产环境 Redis 配置

`REDIS_ENABLED=false` 时令牌吊销失效（登出/改密后旧令牌仍可用）。

**动作**：生产部署设置 `REDIS_ENABLED=true` + `REDIS_ADDR`。

**验收**：登出后旧令牌访问受保护接口返回 401。

---

## 2. 阶段 1：P0 业务域迁移（核心，约 4 人周）

这是让**前端能完整启动**的最小集合，做完即可灰度替换 Java 登录+启动链路。

| 优先级 | 域 | 端点估算 | 关键点 |
|---|---|---|---|
| P0-1 | 字典 / 分类 / 站点配置 | ~15 | ✅ 已迁移（36 端点，2026-11-21）：`queryAllDictItems`/表字典/分类树/租户配置 |
| P0-2 | 当前用户信息 / 菜单 / 权限 | ~12 | `sys/user/info`、`sys/permission/getUserPermissionByToken` |
| P0-3 | 用户管理 CRUD | ~20 | ✅ 已迁移（32 端点，2026-11-21）：核心 CRUD / 角色分配 / 部门关联 / 回收站 / 注册找回密码 |
| P0-4 | 角色 / 权限 / 部门 CRUD | ~30 | RBAC 完整性 |

> 本地完整运行环境（2026-11-21 实测通过）：
> - 便携版 MariaDB 11.4.3（Temp 目录）+ `api/db/teachingopen2.8.sql` 全量导入
>   （69 表 + 种子数据：7 用户/138 菜单/51 字典/4 角色）；
> - `e2e` 工具扩展至 14 步：登录响应 sysAllDictItems、菜单权限树、
>   字典/表字典/分类真实数据抽查；顺带修复了 WriteStatus 把状态文本
>   （"OK"/"Unauthorized"）写入 JSON 响应体的缺陷（全局改用 WriteHeader）。

> P0-1 迁移说明（2026-11-21）：
> - 分类 code 由服务端 YouBian 流水规则生成（含 Z99 追加段边界，单测覆盖）；
> - 表字典动态标识符经 `^[a-zA-Z0-9_]+$` 白名单校验（修复 Java 版 `${}` 插值注入面）；
> - 4 段式带过滤表字典明确拒绝（Java 版为原始 SQL 拼接，不安全）；
> - Excel 导入导出未迁移；`refleshCache` 为兼容占位（Go 版字典直查库）；
> - 已通过本地 MariaDB + V2.8 种子数据回归（见上方完整运行环境）。

> P0-3 迁移说明（2026-11-21，三波次并行 agent 实施 + 主会话集成）：
> - 32 端点分五个域：核心 CRUD（list/add/edit/delete/deleteBatch/frozenBatch）、
>   角色分配（queryUserRole/userRoleList/addSysUserRole/deleteUserRole[Batch]）、
>   部门关联（userDepartList/queryUserByDepId/departUserList/queryByOrgCode[ForAddressList]/
>   editSysDepartWithUser/deleteUserInDepart[Batch]/getCurrentUserDeparts）、
>   用户辅助（queryById/queryByIds/checkOnlyUser/generateUserId/querySysUser/回收站三件套）、
>   注册找回密码（sms/register/phoneVerification/passwordChange，Shiro anon 公开）；
> - add/edit 为 sys_user + sys_user_role + sys_user_depart 三表事务；delete 对齐
>   Java @TableLogic 语义为 del_flag 逻辑删（不级联）；putRecycleBin 实为还原
>   （revertLogicDeleted，1→0）；deleteRecycleBin 物理删并级联清理关联；
> - 有意偏差（均注释文档化）：新账号一律 bcrypt（Java 可逆 DES）；addSysUserRole
>   先全量校验再写库（Java 边校验边写会留部分状态）；排序列白名单 + 列集合校验
>   （Java 仅黑名单）；frozenBatch「异常吞掉仍返回成功」的 Java 缺陷未照搬；
> - 短信验证码 fail-closed：REDIS_ENABLED=true 走 Redis，否则进程内存储兜底，
>   短信通道（阿里云）未接入，验证码落日志仅限开发环境；
> - 暂未迁移：exportXls/importExcel/importStudent、/sys/phoneLogin、
>   /sys/randomImage、/sys/checkCaptcha（找回密码 Step1 依赖，属登录验证码域）；
> - 新增 6 个服务域单测（usercore/userrole/userdepart/useraux/sms/userreg，零 DB 依赖）。

**验收标准**：前端登录后首页、菜单、字典下拉、用户列表全部可用。

> 注意：合并后 `api/` 已是 V2.8，存在 `sys_config` 表与相关接口
> （V2.6 版没有），迁移时需以 V2.8 的 Controller 为契约蓝本。

---

## 3. 阶段 2：P1 核心业务（约 6 人周）

| 域 | 端点估算 | 说明 |
|---|---|---|
| 课程 / 单元 / 课程授权 | ~25 | `teaching_course*` |
| 作业 / 作品 / 互评 / 评分 | ~30 | `teaching_work*`，核心业务 |
| 文件上传 / 素材库 | ~10 | 涉及对象存储抽象 |

---

## 4. 阶段 3：P2 运营功能（约 3 人周）

日志、公告、站内信、定时任务（建议不再依赖 `qrtz_*`，改用独立调度）。

---

## 5. 阶段 4：下线 Java（约 2 人周）

前提：Go 承载 100% 流量 ≥ 2 周无 P0/P1 故障。

---

## 6. 技术债与优化项（穿插进行）

### 6.1 数据库（详见 `DB-OPTIMIZATION-ASSESSMENT.md`）

| 优先级 | 动作 | 说明 |
|---|---|---|
| P1 | 归档 22 张表 | `onl_*`(11) + `demo`/`joa_demo`/`test_*`(6) + `jeecg_*`(5) |
| P3 | 字符集 utf8mb4 | 目前 utf8 无法存 emoji（学生评论场景可能需要） |
| P3 | 修正 `sys_user.status` 注释 | 实际是账号状态，注释误写为"性别" |

### 6.2 代码

| 项 | 说明 |
|---|---|
| `sys_role` 无 `del_flag` | 已修复，后续所有 `sys_role` 查询都不要加该过滤 |
| JWT 角色载荷 | 目前每请求查 DB 校验 admin，可考虑在 claims 加 roles 降低开销 |
| 令牌吊销粒度 | 改密后应吊销该用户全部令牌，当前仅告警（需 Redis 支持 `user:revoked:{id}` 版本号） |
| 契约测试 | 建议用 V2.8 真实响应做 golden file，逐字段 diff |

### 6.3 工程化

| 项 | 说明 |
|---|---|
| CI | 增加 `go build/vet/test` + `gofmt -l` 门禁 |
| Docker | `goframe-server` 补 Dockerfile + docker-compose（MySQL/Redis） |
| 迁移脚本 | 把 `tools/importer` 演进为版本化 migration（golang-migrate 风格） |

---

## 7. 风险清单

| 风险 | 影响 | 对策 |
|---|---|---|
| V2.8 与 V2.6 响应结构差异 | 前端渲染异常 | 阶段 1 先建契约测试，以测试为准 |
| 密码迁移不彻底 | 安全红线 | 上线前必须 `legacy_users=0` |
| 无 Redis 上生产 | 令牌无法吊销 | 部署检查项强制 `REDIS_ENABLED=true` |
| 上游持续更新 | 合并冲突 | 定期（每 2 周）把 master rebase 进来，保持小步 |

---

## 8. 建议的下一步（选一个）

1. **A｜保守**：先完成存量密码迁移 + 生产 Redis 配置，补 CI/Docker，稳定骨架。
2. **B｜进取**：直接开始 P0-1（字典/分类/配置），让前端能登录后加载首页。
3. **C｜架构**：先建契约测试框架（V2.8 golden file），为后续大规模迁移兜底。

> 推荐顺序：**A → C → B**。A 是安全底线（1 天），C 是后续效率保障（3 天），
> B 是真正开始替换（4 周）。
