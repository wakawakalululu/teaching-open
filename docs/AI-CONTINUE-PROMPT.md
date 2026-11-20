# AI 推进 Prompt：teaching-open 后端 GoFrame v2 重构

> 用途：把本文件整体作为 prompt 交给任何 AI 会话，即可在无上下文的情况下
> 按既定方向继续推进重构。配合仓库内 `goframe-v2` Skill 使用。
> 最后更新：2026-11-21（进度：登录链路 + P0-1 + P0-2 完成并本地实测通过，下一任务 P0-3）

---

## 角色与目标

你是 teaching-open 项目的后端重构工程师。**唯一方向**：把 Java（JeecgBoot V2.8）
服务端逐步迁移为 `goframe-server/`（GoFrame v2.10.x），前端 `web/` 零改动对接，
最终以 Go 服务端承载 100% 流量并下线 Java。

- 仓库根目录：`C:\Users\Admin\Desktop\code\teaching-open-v26`（分支 `master`）
- 目标服务端：`goframe-server/`（禁止再动 `go-server/` Gin 版，已废弃）
- 契约蓝本：`api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/`
- 数据库 Schema 唯一依据：`api/db/teachingopen2.8.sql`
- 前端行为依据：`web/src/`（以实际调用为准，不要臆测）

## 总路线（docs/ROADMAP.md 为准）

1. ~~P0-1 字典/分类/站点配置~~ ✅ 完成（36 端点，真实库实测通过）
2. ~~P0-2 当前用户信息/菜单/权限~~ ✅ 完成（登录响应 sysAllDictItems、
   userInfo 白名单扩展、getUserPermissionByToken 菜单树/auth/allAuth，
   本地 MariaDB 实测 menu=23/auth=3/allAuth=10）
3. **P0-3 用户管理 CRUD（当前任务）**：~20 端点，含部门关联、角色分配；
   蓝本 SysUserController.java，注意 sys_user_department/sys_user_role 联表
4. P0-4 角色/权限/部门 CRUD（~30 端点）
5. 阶段 2：课程/作业/作品/文件（P1，~65 端点）
6. 阶段 3：日志/公告/站内信/定时任务（P2）
7. 阶段 4：Go 承载 100% 流量 ≥2 周无故障后下线 Java

## 硬性规范（违反即返工）

1. **GoFrame v2.10.x**，遵循 `.agents/skills/goframe-v2/SKILL.md` 全部约定：
   - 业务逻辑直接写在 `internal/service/<域>/`，**禁止** `logic/` 目录；
   - 数据库操作一律 **DO 对象**（`internal/model/do/`），禁止 `g.Map` 充当 Data；
   - dao/do/entity 按生成风格维护（本仓库无活库可跑 `gf gen dao`，沿用
     `sys_user` 先例手工按 GoFrame CLI 风格书写，列以 `teachingopen2.8.sql` 为准）；
   - 多变量声明用 `var` 块。
2. **契约对齐 Java V2.8 Controller**：路径/方法/参数/响应形状逐字段对齐；
   响应统一 `{success, code, message, result, timestamp}`（用
   `internal/controller/common` 的 OK/OKMsg/Fail；业务失败 HTTP 200 + success:false）。
3. **安全默认拒绝**：新路由必须登记 `verifyRouteGuards`；公开接口必须有 Java
   Shiro anon 名单依据（见 `api/.../config/ShiroConfig.java`），否则全部走鉴权。
4. **防注入**：占位符绑定；动态表名/列名必须过 `^[a-zA-Z0-9_]+$` 白名单或列集合校验；
   Java 版原始 SQL 拼接的能力（如 4 段式过滤表字典）明确拒绝并文档化，不要照搬。
5. **JeecgBoot 特殊语义**：`del_flag` 是 0/1 标志位（非时间戳），手动过滤，**不要**
   启用 GoFrame 软删特性；`create_time`/`update_time` 由服务端写入（DO 传 gtime.Now()）。
6. **错误处理**：与现有包保持一致（errors.New/fmt.Errorf + %w 包装），日志脱敏。
7. **验证门槛（每次交付必过）**：
   ```bash
   cd goframe-server && go build ./... && go vet ./... && go test ./... -count=1 && gofmt -l .
   ```
8. **交付动作**：每完成一个域：更新 `docs/ROADMAP.md` 进度快照 +
   `goframe-server/README.md` 端点表 → conventional commit → push 到
   `origin/master`（全局约定「固化成果并推送到远程仓库」）。

## 当前状态快照（2026-11-21）

- 已迁移端点（~50）：登录/刷新/登出/改密/重置、`/sys/user/info`（简化版）、
  P0-1 全部 36 个、`/sys/permission/getUserPermissionByToken`。
- 登录响应已含 Java 契约的 `{token, userInfo, sysAllDictItems}`（refreshToken
  为 Go 版扩展字段）；`userInfo` 输出白名单超集。
- `tools/e2e` 已扩展到 14 步并全过（本地 MariaDB + V2.8 种子数据）。
- 事实核对结论（勿重复分析）：
  - `web/src/api/login.js` 的 `getInfo()`（/api/user/info）是死代码，无调用方；
  - `sys_role` 表**无** `del_flag` 列，查询禁止加该过滤；
  - Java 菜单树中 menuType=2（按钮）不进路由树（getPermissionJsonObject 返回
    null 被跳过），按钮权限全部走 auth/allAuth；
  - **GoFrame 响应铁律**：设置 HTTP 状态码用 `WriteHeader`，**禁用**
    `WriteStatus`（无内容时会写状态文本污染 JSON）——已全局修复；
  - dump 导入需去掉 `RENAME TABLE qrtz_*` 行（与 lower_case_table_names=1 冲突）。

## 本机环境事实

- Windows 10，Git Bash；有 Go 1.27、无 Docker、无 gcc（cgo 不可用）。
- **本地完整运行环境已建成**（Temp 目录，一次性验证用）：
  - 便携版 MariaDB 11.4.3：`C:\Users\Admin\AppData\Local\Temp\goframe-mysql\mariadb-11.4.3-winx64`，
    数据目录 `...\goframe-mysql\data`，端口 3307，root 无密码；
    启动：`mysqld.exe --datadir=... --port=3307 --console`
  - 已导入 `api/db/teachingopen2.8.sql`（去掉 RENAME TABLE 行）：
    69 表种子数据（7 用户/138 菜单/51 字典/4 角色/23 配置）
  - 服务端启动：`JWT_SECRET=<≥32字节> MYSQL_LINK="mysql:root:@tcp(127.0.0.1:3307)/teachingopen?parseTime=true&loc=UTC" REDIS_ENABLED=false /tmp/goframe-server.exe`
  - e2e：`MYSQL_LINK=... go run ./tools/e2e`（14 步）
  - 注意：`RENAME TABLE` 冲突、端口占用旧进程、控制台中文乱码（仅显示问题）
- sqlpub 远程库凭据不在仓库中，需要用户提供；无法连库时以编译+单测+本地
  MariaDB 为底线，不要谎报"已验证"。

## 推进循环（每个工作单元）

1. 读 Java Controller/Service/Mapper XML，把契约写进实现注释；
2. 对照 `web/src` 确认前端真实消费的字段，白名单输出；
3. 实现 service + controller + 路由 + 守卫登记；
4. 过第 7 条验证门槛；
5. 更新 ROADMAP/README → commit → push；
6. 在下一单元开始前重新确认 ROADMAP 进度，避免重复劳动。
