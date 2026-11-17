"""静态结构检查：括号配对 + Java 基本语法一致性。"""
import io
import os
import sys

FILES = [
    r"api/jeecg-boot-base-common/src/main/java/org/jeecg/common/util/SqlIdentifierUtil.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/config/ShiroConfig.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/controller/LoginController.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/controller/SysUserController.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/controller/SysDictController.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/controller/SysCategoryController.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/controller/SysDepartPermissionController.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/ngalain/controller/NgAlainController.java",
    r"api/jeecg-boot-base-common/src/main/java/org/jeecg/common/system/controller/CommonController.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/service/impl/SysUserServiceImpl.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/modules/system/mapper/SysUserMapper.java",
    r"api/jeecg-boot-base-common/src/main/java/org/jeecg/common/exception/JeecgBootExceptionHandler.java",
    r"api/jeecg-boot-base-common/src/main/java/org/jeecg/common/util/RestrictedSqlFragmentUtil.java",
    r"api/jeecg-boot-module-system/src/main/java/org/jeecg/config/WebMvcConfiguration.java",
]

SQ = chr(39)
DQ = chr(34)


def strip_code(text):
    """去掉注释与字符串字面量，仅保留结构字符。"""
    out = []
    i, n = 0, len(text)
    while i < n:
        c = text[i]
        if c == "/" and i + 1 < n and text[i + 1] == "/":
            while i < n and text[i] != "\n":
                i += 1
            continue
        if c == "/" and i + 1 < n and text[i + 1] == "*":
            i += 2
            while i + 1 < n and not (text[i] == "*" and text[i + 1] == "/"):
                i += 1
            i += 2
            continue
        if c == DQ or c == SQ:
            quote = c
            i += 1
            while i < n:
                if text[i] == "\\":
                    i += 2
                    continue
                if text[i] == quote:
                    i += 1
                    break
                i += 1
            continue
        out.append(c)
        i += 1
    return "".join(out)


def check(path):
    with io.open(path, encoding="utf-8") as fh:
        raw = fh.read()
    code = strip_code(raw)
    res = []
    for open_c, close_c in (("{", "}"), ("(", ")"), ("[", "]")):
        diff = code.count(open_c) - code.count(close_c)
        res.append((open_c + close_c, diff))
    return res, raw


failed = False
for f in FILES:
    if not os.path.exists(f):
        print("MISSING", f)
        failed = True
        continue
    res, raw = check(f)
    flag = "OK " if all(d == 0 for _, d in res) else "BAD"
    if flag == "BAD":
        failed = True
    detail = " ".join("%s:%+d" % (name, d) for name, d in res)
    print("%s %-42s %s" % (flag, os.path.basename(f), detail))
    if raw.count("\n") == 0:
        print("   WARN empty file")

print("RESULT:", "FAIL" if failed else "PASS")
sys.exit(1 if failed else 0)
