# 管理面板「后端能力缺失」根因分析与升级方向

> 日期：2026-11-22　　基线：Java V2.8（411 个方法级端点）× GoFrame v2.10.4
> 方法：Java 注解静态盘点 → 运行态路由表转储（`dumpRouterMap`）→ 前端源码调用点
> 提取（316 条路径 / 133 条 verb 对）→ Puppeteer 逐页实测（102 个菜单 URL）。

---

## 1. 结论摘要

管理面板「大量功能像是坏的」并非单一原因，实测归为 **四类根因**，
其中只有一类是真正的后端能力缺失：

| 类别 | 占比（主观感受） | 实证 | 是否真缺能力 |
| --- | --- | --- | --- |
| A. 系统表被清空 → 页面「暂无数据」 | ~55% | `sys_position/sys_data_source/sys_sms_template/sys_depart_role/...` 行数 = 0 | ❌ 数据问题 |
| B. Excel 导入导出域未注册 → 按钮 404 | ~20% | 11 个域 `/exportXls`、`/importExcel` 无路由 | ✅ 已修复 |
| C. HTTP 方法不一致 → 前端 405/404 | ~10% | Java `GET /sys/permission/edit`，Go 只注册 PUT/POST | ✅ 已修复 |
| D. 演示/未适配模块 | ~15% | MockController 19、Test 7、Actuator 4、ThirdLogin 3、OSS 4、CAS 1 | ❌ 有意排除 |

**端点覆盖变化：80.7% → 89.5%**（缺 76 → 43，方法不一致 4 → 0）。
剩余 43 条中 39 条属 demo/test/Spring 专属/未配置凭据，真实业务缺口仅约 4 条（OSS 域）。

---

## 2. 根因逐条分析

### A. 系统表被清空（最主要原因）

「全量替换 Mock 数据」时按 `DELETE FROM` 清空了 V2.8 演示库，但只回填了业务表
（用户/课程/作品），**系统字典类表未回填**：

```
sys_position 0 · sys_data_source 0 · sys_sms_template 0
sys_depart_role 0 · sys_depart_permission 0 · sys_log 0 · sys_data_log 0
```

于是页面渲染正常、接口返回 200，但表格是「暂无数据」——用户感知为「功能没做」。
**处置**：`chseed` 新增 `seedAdminBase()`，`chunhui_seed.sql` 新增第 14 节：
职务 5 / 数据源 1 / 消息模板 3 / 部门角色 3。

### B. Excel 域未注册（真实缺口，影响「导出/导入」按钮）

`internal/cmd/register_p2_6.go` 的 `excelxBases` 只列了 15 个域，
而前端管理页对另外 11 个域也提供了「导出/导入」按钮：

```
sys/dict · sys/dictItem · sys/category · sys/role · sys/sysDepart
sys/sysDepartPermission · sys/sysDepartRole · system/sysFile
sys/sysUserAgent · message/sysMessage · message/sysMessageTemplate
teaching/teachingDepartDayLog
```

**处置**：`excelx` 注册表补 12 份列规格（按 `@Excel` 表头语义），
`specTable`/`delFlagCol` 补映射，`excelxBases` 补 12 个 Base → **新增 36 条路由**。
实测 `GET /api/sys/category/exportXls` 返回合法 xlsx（PK 头）。

### C. HTTP 方法不一致（真实缺口，隐蔽且难排查）

Java 侧一批 `edit` 端点声明为 `GET`（历史遗留，依赖 query 参数），
Go 侧按 REST 语义只注册了 `PUT`/`POST`：

| Java 契约 | Go 原实现 | 影响 |
| --- | --- | --- |
| `GET /sys/permission/edit` | 仅 PUT/POST | 菜单「编辑」回显 404 |
| `GET /sys/permission/editPermissionRule` | 仅 PUT/POST | 数据规则编辑失败 |
| `GET /teaching/menu/edit` | 仅 PUT/POST | 前台菜单编辑失败 |
| `GET /sys/logout` | 仅 POST | 部分页面退出 404 |
| `PUT /sys/selectDepart` | **仅声明守卫，无路由** | 多部门用户登录后选择部门失败 |
| `GET /sys/loginfo` | **仅声明守卫，无路由** | 首页访问统计 404 |

**处置**：新增 `internal/cmd/register_gaps.go` 统一补齐别名；`selectDepart`/`loginfo`
补实际路由（守卫早已存在，属「声明了但没挂」的遗漏）。

### D. 演示 / 未适配模块（有意排除）

- `MockController` 19 条：前端 Vue 本地 mock 的配套接口（`/api/api/user` 等双前缀），
  仅工作台示例组件调用 → 工作台 6 个 widget 报 404 属预期，不迁移。
- `TestController`/`TestWebController` 7 条：示例/鉴权演示。
- `ActuatorRedisController` 4 条：Spring Actuator 专属，Go 无对应运行时。
- `ThirdLoginController` 3 条：GitHub/企业微信/钉钉 OAuth，未配置凭据。
- `CasClientController` 1 条：CAS SSO 未部署。
- `OSSFileController` 4 条：MinIO 对象存储域（当前 `UPLOAD_TYPE=local`）。

---

## 3. 本轮已交付

| 变更 | 文件 | 端点数 |
| --- | --- | --- |
| Excel 域补全（12 域 × 3） | `service/excelx/registry.go`、`controller/excelx/excelx.go`、`cmd/register_p2_6.go` | +36 |
| 方法别名与遗漏端点 | `cmd/register_gaps.go`、`cmd/cmd.go` | +18 |
| 字典通用表树（白名单） | `service/dict/dict.go`、`controller/dict/dict.go` | +1 |
| 管理基础数据回填 | `tools/chseed/main.go`、`api/db/chunhui_seed.sql` | 12 行数据 |

验证证据：
- 路由转储：**454 条**（原 403），Java 对账覆盖率 **89.5%**、方法差异 **0**。
- 实测：`exportXls` 返回 xlsx；`executeRuleByCode/org_num_role` → `A02`；
  `loadTreeData`、`appUserList`、`getUserSectionInfoByToken`、`getDictItemsByTable`、
  `loginfo`、`logout`、`selectDepart` 全部 200（500 仅为缺参业务校验，路由可达）。
- 管理页截图：职务管理 5 行、通讯录 10 人、数据源/模板/部门角色均非空。
- 门禁：`go vet / build / test / gofmt` 全绿。

---

## 4. 后续升级方向（按投入产出排序）

### P0 观察期（本轮之后立即可做）
1. **残留缺口清零**：`/sys/dict/queryTableData`、`/sys/common/transitRESTful`、
   `/sys/common/403`（3 条，半天）。
2. **OSS 域**：若确定不用 MinIO，删除前端 OSS 入口；若要用，按 `upload` 域抽象
   增加 `minio` 驱动（1~2 天）。

### P1 架构收敛（1~2 周）
3. **路由声明单一化**：现状存在「守卫已声明但路由未挂」（`selectDepart`/`loginfo`）
   这类漂移，根因是路由表与守卫表两处手写。建议改为**单一注册表**驱动：
   `type Route{Method, Path, Handler, Auth, Anon}` 一处定义，`verifyRouteGuards`
   自动校验，「声明-注册」不可能再分叉。
4. **Excel 规格自动生成**：12 份列规格是手写的，与 DDL 存在漂移风险。
   建议由 `information_schema` + 字典表生成初稿，人工确认后落盘，
   并在 CI 中做「规格列 vs 表列」一致性断言。
5. **契约测试**：把 `tools/probe_pairs` 的思路固化为 `go test`：解析 Java 注解
   → 断言 Go 路由存在（当前为一次性脚本）。防止后续回退。

### P2 数据治理（2~4 周）
6. **种子分层**：把 `chunhui_seed.sql` 拆成 `base`（系统表）与 `demo`（业务示例）
   两层，避免再次出现「换业务数据顺手清空系统表」。
7. **空表可观测**：启动时对管理页依赖的表做行数探针，空表打 WARN 日志，
   把「感觉功能坏了」变成「日志里的一行告警」。

### P3 未适配模块的显式决策
8. 工作台（Mock 19 条）：改造为真实数据（课程/作品/学员统计），
   或从前端菜单隐藏示例 widget。**建议改造**——工作台是管理员首屏。
9. 在线表单（`online/*`）、工作流（`bpm/extbpm`）、CAS/OAuth：
   在菜单层显式置灰并标注「本部署未启用」，避免点击后 404 造成「系统坏了」的错觉。

---

## 5. 复现命令

```bash
# 1) 路由转储对账
sed -i 's/dumpRouterMap: false/dumpRouterMap: true/' goframe-server/manifest/config/config.yaml
./goframe-server > /tmp/rm.log 2>&1 &
grep -E '^\| :18080' /tmp/rm.log | awk -F'|' '{...}'   # 见 tools 脚本

# 2) 前端调用点提取
grep -rhoE '(getAction|postAction|putAction|deleteAction)\("[^"]+' web/src | sort -u

# 3) 管理页逐页实测
node build/probe_admin.js   # 输出 /tmp/admin_probe.json
```

---

## 6. 第二轮推进结果（2026-11-22，按 §4 计划执行）

### 6.1 已完成项

| 计划项 | 交付 | 验证 |
| --- | --- | --- |
| P0-1 残留端点 | `/sys/dict/queryTableData`、`/sys/common/403`、`/sys/common/transitRESTful`（SSRF 防护：显式拒绝）、`/sys/upload/uploadMinio`（本地等价实现）、`/sys/visitInfo` | 全部 200 / 明确业务错误 |
| P0-2 OSS 域 | `oss_file` dao/do/entity + `service/oss` + `controller/oss`，4 端点（本地存储后端，MinIO 接入点隔离在 service 层） | `/oss/file/list` 200 |
| P1-1 路由一致性测试 | `internal/cmd/routes_test.go`：源码解析守卫与注册，动态域（excelxBases/p2_6）展开后比对 | **守卫 310 条 / 注册 472 条全部对应**；首跑即发现 `/sys/visitInfo` 漂移并修复 |
| P1-2 Excel 规格一致性测试 | `internal/service/excelx/registry_test.go`：Spec 列 vs DDL 列双向断言；`TableByBase` 收敛为唯一映射源 | **27 个 Excel 域全部一致**；首跑发现 CourseDept/CourseUnit 缺 Spec（导出报「未配置」）并补齐 |
| P1-3 契约覆盖测试 | `internal/cmd/contract_test.go`：解析 Java 注解（跳过注释方法），排除清单显式化，阈值 85% | **372/372 = 100%**（404 条 Java 端点中 demo/test/Spring/OAuth/CAS 共 32 条按白名单排除） |
| P3 工作台真实化 | 7 端点：teams/activity/radar/role/user/projects/task-process + service 健康；响应形状按 Workplace.vue 逐字段对齐（`result.data`、`user.nickname`、`个人/团队/部门`） | 工作台截图全量真实数据，**API 失败数 0** |

### 6.2 关键结论

- 端到端契约覆盖率由 **89.5% → 100%**（排除项均有文档化理由）。
- 新增 3 个回归测试把「一次性对账」变成持续门禁，其中两个测试**首次运行即各发现 1 个真实缺陷**。
- 工作台从「骨架屏 + 5 个 404」变为真实数据（6 门课程项目卡、8 条动态、3 个班级团队、雷达图有值）。

### 6.3 下一步（P1 剩余 + P2）

1. **路由单一注册表**（P1 架构项）：现状仍是「路由 + 守卫」两处手写，虽已有测试兜底，
   仍建议收敛为 `Route{Method,Path,Handler,Auth}` 结构体注册，测试改为运行时断言。
2. **种子分层**（P2）：`base`（系统表）/`demo`（业务示例）拆分，避免再次清空系统表。
3. **空表探针**（P2）：启动时对管理页依赖表做行数检查并打 WARN。
4. **未启用模块菜单置灰**（P3）：在线表单/工作流/CAS 在菜单层标注，避免 404 错觉。
