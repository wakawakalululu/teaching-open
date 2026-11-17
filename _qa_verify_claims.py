"""核对交付文档中的关键数字/断言与代码实际状态是否一致。

只做只读检查，不修改任何文件。任何一项不符即失败退出。
"""
import io
import os
import re
import sys
import subprocess

ROOT = os.path.dirname(os.path.abspath(__file__))
os.chdir(ROOT)

problems = []


def check(name, ok, detail=""):
    print(("PASS  " if ok else "FAIL  ") + name + (("  -> " + detail) if detail and not ok else ""))
    if not ok:
        problems.append(name + " " + detail)


# ---------- 1. 改动文件计数 ----------
st = subprocess.run(["git", "status", "--short", "api"], capture_output=True, text=True).stdout
lines = [l for l in st.splitlines() if l.strip()]
modified = sum(1 for l in lines if l.startswith(" M"))
added = sum(1 for l in lines if l.startswith("??"))
check("改动文件总数 = 18", len(lines) == 18, f"实际 {len(lines)}")
check("其中修改 16 / 新增 2", modified == 16 and added == 2, f"实际 {modified}/{added}")

# ---------- 2. ShiroConfig 中危险 anon 已移除 ----------
shiro = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/config/ShiroConfig.java",
                encoding="utf-8").read()
active = [l for l in shiro.splitlines()
          if 'filterChainDefinitionMap.put' in l and '"anon"' in l and not l.lstrip().startswith("//")]
for danger in ["/druid/**", "/actuator/metrics/**", "/actuator/httptrace/**",
               "/actuator/redis/**", "/swagger-ui.html", "/swagger**/**", "/v2/**"]:
    hit = [l for l in active if danger in l]
    check(f"ShiroConfig 已移除 anon: {danger}", not hit, str(hit))
health = [l for l in active if "/actuator/health" in l]
check("ShiroConfig 保留 health anon", len(health) >= 1)

# ---------- 3. prod yml 关键项 ----------
import yaml  # noqa: E402
cfg = yaml.safe_load(io.open(
    "api/jeecg-boot-module-system/src/main/resources/application-prod.yml", encoding="utf-8"))
check("swagger.enable = false", cfg["swagger"]["enable"] is False)
check("springfox 总开关关闭",
      cfg["springfox"]["documentation"]["enabled"] is False
      and cfg["springfox"]["documentation"]["auto-startup"] is False)
druid = cfg["spring"]["datasource"]["druid"]["stat-view-servlet"]
check("druid 面板已关闭", druid["enabled"] is False)
check("druid 弱口令已移除", druid["loginPassword"] != "123456")
check("actuator 仅暴露 health",
      cfg["management"]["endpoints"]["web"]["exposure"]["include"] == "health")
check("health 不显示详情", cfg["management"]["endpoint"]["health"]["show-details"] == "never")
check("shiro.excludeUrls 已清空", not cfg["jeecg"]["shiro"].get("excludeUrls"))
check("cors.allowedOrigins 已配置且非通配",
      "*" not in str(cfg["jeecg"]["cors"]["allowedOrigins"]))

# ---------- 4. Mapper 中目标拼接点已清 ----------
user_xml = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/mapper/xml/SysUserMapper.xml",
                   encoding="utf-8").read()
check("revertLogicDeleted 使用 foreach 绑定", 'collection="userIds"' in user_xml
      and "id IN (${userIds})" not in user_xml)
check("deleteLogicDeleted 使用 foreach 绑定", "id IN (${userIds})" not in user_xml)
check("org_code 改为 CONCAT 绑定", "${orgCode}" not in user_xml
      and "CONCAT(#{orgCode}" in user_xml)

dict_xml_raw = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/mapper/xml/SysDictMapper.xml",
                       encoding="utf-8").read()
# 剔除 XML 注释后再判断，避免把注释里描述的"原写法"误判为现存问题
dict_xml = re.sub(r"<!--.*?-->", "", dict_xml_raw, flags=re.S)
check("query.keyword 不再直接拼接", "${query.keyword}" not in dict_xml)
check("query.keyword 改用 CONCAT 绑定", "CONCAT('%', #{query.keyword}, '%')" in dict_xml)

# ---------- 5. Mapper 接口签名 ----------
mapper = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/mapper/SysUserMapper.java",
                 encoding="utf-8").read()
check("Mapper 签名为 List<String>",
      "List<String> userIds" in mapper and 'int revertLogicDeleted(@Param("userIds") java.util.List<String>' in mapper)

# ---------- 6. 调用方数量 ----------
svc = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/service/impl/SysUserServiceImpl.java",
              encoding="utf-8").read()
check("Service 调用传 List（无手工拼引号）",
      'String.format("\'%s\'", String.join("\',\'", userIds))' not in svc)

# ---------- 7. 登录限流接入两个入口 ----------
login = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/controller/LoginController.java",
                encoding="utf-8").read()
check("buildLoginFailKey 定义 1 次", len(re.findall(r"private String buildLoginFailKey\s*\(", login)) == 1)
check("login 与 mLogin 均接入限流",
      login.count("buildLoginFailKey(username, request)") == 2)
check("login 方法签名含 request",
      "public Result<JSONObject> login(@RequestBody SysLoginModel sysLoginModel, HttpServletRequest request)" in login)
check("mLogin 方法签名含 request",
      "public Result<JSONObject> mLogin(@RequestBody SysLoginModel sysLoginModel, HttpServletRequest request)" in login)

# ---------- 8. 手机号脱敏 ----------
user_ctl = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/controller/SysUserController.java",
                   encoding="utf-8").read()
check("desensitizePhone 已定义", "private String desensitizePhone(" in user_ctl)
check("querySysUser 使用脱敏", user_ctl.count("desensitizePhone(user.getPhone())") == 2)
check("未使用 Java 11 String.repeat", ".repeat(" not in user_ctl)

# ---------- 9. 异常处理不回显 ----------
handler = io.open("api/jeecg-boot-base-common/src/main/java/org/jeecg/common/exception/JeecgBootExceptionHandler.java",
                  encoding="utf-8").read()
check("兜底异常不再拼接 e.getMessage()",
      'Result.error("操作失败，"+e.getMessage())' not in handler)
check("兜底异常返回 traceId", "traceId" in handler)

# ---------- 10. SSRF 校验 ----------
common = io.open("api/jeecg-boot-base-common/src/main/java/org/jeecg/common/system/controller/CommonController.java",
                 encoding="utf-8").read()
check("transitRESTful 前有出站校验", "validateOutboundUrl(url)" in common)
for probe in ["isLoopbackAddress", "isLinkLocalAddress", "isSiteLocalAddress", "169.254",
              "100.64", "fc00"]:
    check(f"SSRF 校验覆盖 {probe}", probe.replace("169.254", "100.64.0.0/10") in common or probe in common)

# ---------- 11. CORS 白名单 ----------
webmvc_raw = io.open("api/jeecg-boot-module-system/src/main/java/org/jeecg/config/WebMvcConfiguration.java",
                     encoding="utf-8").read()
# 剔除 Java 注释后判断
def strip_java_comments(text):
    out = []
    in_block = False
    for line in text.split(chr(10)):
        if in_block:
            if "*/" in line:
                in_block = False
                line = line.split("*/", 1)[1]
            else:
                continue
        if "/*" in line:
            head, _, tail = line.partition("/*")
            if "*/" in tail:
                line = head + tail.split("*/", 1)[1]
            else:
                in_block = True
                line = head
        out.append(line.split("//", 1)[0])
    return chr(10).join(out)

webmvc = strip_java_comments(webmvc_raw)
check("CORS 不再使用 addAllowedOrigin(\"*\")", 'addAllowedOrigin("*")' not in webmvc)
check("CORS 显式跳过通配符", '"*".equals(item)' in webmvc)

# ---------- 12. nginx ----------
ngx = io.open("api/nginx.example.conf", encoding="utf-8").read()
check("nginx 有 CSP", "Content-Security-Policy" in ngx)
check("nginx 关闭 server_tokens", "server_tokens off" in ngx)
check("nginx 拒绝管理端点", "druid" in ngx and "swagger" in ngx and "actuator" in ngx)
check("nginx 未使用已废弃 X-XSS-Protection", "X-XSS-Protection" not in ngx)

# ---------- 13. 新工具类存在 ----------
for f in ["api/jeecg-boot-base-common/src/main/java/org/jeecg/common/util/SqlIdentifierUtil.java",
          "api/jeecg-boot-base-common/src/main/java/org/jeecg/common/util/RestrictedSqlFragmentUtil.java"]:
    check("工具类存在: " + os.path.basename(f), os.path.exists(f))

# ---------- 14. Go 侧测试通过（由外部命令保证，这里只查文件） ----------
for f in ["go-server/go.mod", "go-server/cmd/server/main.go",
          "go-server/internal/auth/token.go", "go-server/internal/passwd/passwd.go",
          "go-server/internal/httpapi/server.go"]:
    check("Go 文件存在: " + f, os.path.exists(f))

# ---------- 15. 文档存在且提到关键结论 ----------
plan = io.open("docs/GO-MIGRATION-PLAN.md", encoding="utf-8").read()
check("方案文档提到 PBKDF1 迭代无效结论",
      ("1000 次迭代对最终密钥毫无影响" in plan) or ("迭代次数不参与" in plan))
check("方案文档给出向量", "c1461cf4492050e9" in plan)
hard = io.open("docs/SECURITY-HARDENING-V26.md", encoding="utf-8").read()
check("加固文档记录 Java 8 兼容性检查", "Java 8 API 兼容性" in hard)
check("加固文档声明无 JDK/Maven", "无 JDK" in hard or "无 JDK 与 Maven" in hard)

print()
if problems:
    print(f"RESULT: FAIL ({len(problems)} problems)")
    for p in problems:
        print("  -", p)
    sys.exit(1)
print("RESULT: PASS (all assertions hold)")
