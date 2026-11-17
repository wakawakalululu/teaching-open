# 数据库结构优化评估

> 评估对象：`teachingopen2.6.sql`（70 张表）+ 已导入 sqlpub `system4ai` 库的实际状态
> 评估日期：2026-11-17

## 1. 结论先行

**数据库结构本身无需大改，但需要做三类动作：**

1. **必须做**：密码字段迁移（DES 可逆密文 → bcrypt）——这是安全红线，已通过 rehash-on-login 机制自动处理。
2. **建议做**：删除/归档 16 张低代码与示例表（`onl_*`、`demo`、`joa_demo`、`test_*`），减少 15% 复杂度。
3. **可做**：补充缺失索引、统一命名规范，但非紧急。

## 2. 表分类与处置建议

### 2.1 核心业务表（保留 + 继续使用，P0/P1）

| 前缀 | 数量 | 表 | 说明 |
|---|---|---|---|
| `sys_user`/`sys_role`/`sys_permission`/`sys_depart` | 16 | 账号/RBAC/组织 | 所有业务前置，已迁移 |
| `sys_dict`/`sys_dict_item`/`sys_category` | 3 | 字典/分类 | 前端启动依赖 |
| `teaching_*` | 6 | 课程/作业/作品/素材 | 核心业务 |

**这些表结构合理，无需改动。**

### 2.2 建议删除/归档（P3，可砍 15% 复杂度）

| 表 | 数量 | 理由 |
|---|---|---|
| `onl_cgform_*` / `onl_cgreport_*` | 11 | JeecgBoot 低代码引擎，开源版前端无入口，Go 版不实现 |
| `demo` / `joa_demo` / `test_demo` / `test_person` / `test_order_main` / `test_order_product` | 6 | 示例与测试表 |
| `jeecg_monthly_growth_analysis` / `jeecg_order_*` / `jeecg_project_nature_income` | 5 | JeecgBoot 示例报表 |

**合计 22 张表可归档**（先备份，不物理删除，移到 `_archive_` 前缀或独立库）。

### 2.3 保留但暂不迁移（P2）

| 前缀 | 数量 | 说明 |
|---|---|---|
| `qrtz_*` + `sys_quartz_job` | 12 | 定时任务，可用独立调度替代 |
| `sys_log`/`sys_announcement`/`sys_sms`/`sys_file`/`oss_file`/`sys_data_log`/`sys_data_source` | 7 | 日志/公告/文件，后置迁移 |

## 3. 关键发现：密码字段（最高优先级）

已确认（实测 `verifyseed` 工具）：

```
密码迁移进度: total=3 bcrypt=0 legacy=3
  admin    密码方案=legacy-DES
  teacher  密码方案=legacy-DES
  student  密码方案=legacy-DES
```

- 3 个种子用户全部是 **DES 可逆密文**（`PBEWithMD5AndDES` 加密用户名），可被解密还原明文。
- **这是必须优先处理的安全红线**：库泄露 = 全站明文密码泄露。
- GoFrame 版已实现 `rehash-on-login`：存量用户下次登录时自动升级为 bcrypt，无需停机。

**建议动作**：
1. 尽快让 3 个种子用户各登录一次（触发 rehash），或提供强制改密入口。
2. 上线前执行全量 `reportLegacyAccounts` 统计，确认 legacy=0 后再删 Legacy 代码。

## 4. 其他优化项（非紧急）

### 4.1 字符集
- 所有表 `DEFAULT CHARSET=utf8`（3 字节 UTF-8），无法存储 emoji 等 4 字节字符。
- 若需支持 emoji（如学生作品评论），建议改 `utf8mb4`。当前非阻塞。

### 4.2 索引
- `sys_user` 已有 `username`/`phone`/`email`/`work_no` 唯一索引 + `status`/`del_flag` 普通索引，够用。
- 未发现明显缺失的查询索引（登录/鉴权链路已覆盖）。

### 4.3 字段类型
- `sys_user.status` 注释写的是"性别"，实际语义是"账号状态"（1正常 2冻结）——注释错误，需修正（不影响功能）。
- `id` 全部是 `varchar(32)`（雪花ID），合理，无需改自增。

### 4.4 命名一致性
- 部分表字段大小写不一（`sys_user_depart.ID` 大写 vs 其他表 `id` 小写），历史遗留，不阻塞。

## 5. 优先级建议

| 优先级 | 动作 | 工作量 | 时机 |
|---|---|---|---|
| **P0** | 密码 rehash 迁移（3 个种子用户先登录触发） | 即时 | 现在 |
| **P1** | 归档 22 张 onl/demo/test/jeecg 表 | 1 天 | 迁移 P1 业务前 |
| ~~P2~~ | ~~补 `reportLegacyAccounts` 到 GoFrame 版~~ | ✅ 已完成 | 2026-11-17 |
| **P3** | 字符集 utf8mb4 + 注释修正 | 0.5 天 | 有空时 |

## 5.1 已完成的密码迁移能力（2026-11-17）

| 能力 | 实现 | 验证 |
|---|---|---|
| 被动迁移 | 登录时校验 legacy 密文成功后 rehash 为 bcrypt | E2E 步骤 2→3 通过 |
| 主动迁移 | 自助改密 / 管理员重置直接写 bcrypt | E2E 步骤 4 通过 |
| 迁移进度观测 | 启动日志输出 total/bcrypt/legacy | 实测 `total=3 bcrypt=0 legacy=3` |
| 算法兼容性 | Go 侧 Encrypt/Decrypt 与 Java 4 组向量逐字节一致 | 单元测试通过 |

> 3 个种子用户只需**各登录一次**（或由管理员重置密码）即可完成迁移。
> 迁移完成后 `legacy_users=0`，可删除 `passwd.go` 的 Legacy 部分。

## 6. 环境接入记录（已完成）

- **MySQL**：sqlpub `mysql2.sqlpub.com:3307/system4ai`，已全量导入 70 张表，连接验证通过。
- **飞书告警**：`FEISHU_WEBHOOK_URL` 已接入，账号锁定/数据库故障时异步推送。
- **Redis**：本地暂无，已实现 fail-open 降级（登录限流/令牌吊销在 Redis 不可用时放行，不影响登录）。
- **登录链路**：已实测跑通（错误密码 401、空参数 400、账号不存在 401 同话术防枚举）。
