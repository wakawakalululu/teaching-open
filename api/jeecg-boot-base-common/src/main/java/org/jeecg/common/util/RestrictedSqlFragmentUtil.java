package org.jeecg.common.util;

import java.util.regex.Pattern;

import lombok.extern.slf4j.Slf4j;

/**
 * 受限 SQL 片段校验工具。
 *
 * <p>项目中存在少量"必须拼接 SQL 片段"的历史设计，例如表字典的
 * <code>filterSql</code>（形如 <code>status = '1' and del_flag = 0</code>）。
 * 这类片段无法用 #{} 绑定，只能做受限白名单。
 *
 * <p>本工具采用「字符白名单 + 危险构造拒绝 + 长度上限」三重限制，
 * 目标是让片段只能表达简单的列比较与逻辑组合，无法闭合引号、
 * 无法堆叠语句、无法注释掉后续条件。
 *
 * <p>注意：这是纵深防御的最后一层，不应作为首选方案。新代码应尽量
 * 改为显式列 + #{} 绑定。
 *
 * @author security-hardening
 */
@Slf4j
public class RestrictedSqlFragmentUtil {

    /**
     * 允许出现的字符：字母、数字、下划线、空格、以及
     * <code>= &lt; &gt; ! ( ) ' , . - *</code>。
     * 刻意不允许 <code>;</code>、<code>#</code>、<code>/</code>、<code>"</code>、
     * <code>%</code>、反斜杠 等可用于注释、转义或多语句堆叠的字符。
     */
    private static final Pattern ALLOWED_CHARS = Pattern.compile("^[A-Za-z0-9_\\s=<>!()',.*\\-]+$");

    /** 危险构造：语句堆叠、注释、系统函数、时间盲注函数、系统库。 */
    private static final String[] FORBIDDEN_FRAGMENTS = {
            ";", "--", "/*", "*/", "#",
            "union", "select ", "insert ", "update ", "delete ", "drop ", "alter ",
            "create ", "truncate ", "grant ", "revoke ",
            "sleep(", "benchmark(", "extractvalue(", "updatexml(", "load_file(",
            "into outfile", "into dumpfile",
            "information_schema", "performance_schema", "mysql.",
            "char(", "concat(", "ascii(", "substring(", "mid(",
    };

    /** 片段长度上限，防止构造超长条件拖慢数据库。 */
    private static final int MAX_LENGTH = 512;

    private RestrictedSqlFragmentUtil() {
    }

    /**
     * 校验受限 SQL 片段，不合法立即抛异常。
     *
     * @param fragment  待校验片段
     * @param fieldName 出错字段名，仅用于异常定位
     * @return 原样返回，便于链式使用
     * @throws IllegalArgumentException 校验失败
     */
    public static String assertSafe(String fragment, String fieldName) {
        if (fragment == null || fragment.trim().isEmpty()) {
            throw new IllegalArgumentException("非法的SQL片段[" + fieldName + "]: 不能为空");
        }
        String value = fragment.trim();
        if (value.length() > MAX_LENGTH) {
            log.warn("SQL片段校验失败[{}]: 超长 {} 字符", fieldName, value.length());
            throw new IllegalArgumentException("非法的SQL片段[" + fieldName + "]: 长度超过 " + MAX_LENGTH);
        }
        if (!ALLOWED_CHARS.matcher(value).matches()) {
            log.warn("SQL片段校验失败[{}]: 含非法字符 -> {}", fieldName, value);
            throw new IllegalArgumentException("非法的SQL片段[" + fieldName + "]: 含不允许的字符");
        }
        String lower = value.toLowerCase();
        for (String bad : FORBIDDEN_FRAGMENTS) {
            if (lower.contains(bad)) {
                log.warn("SQL片段校验失败[{}]: 命中危险构造 {} -> {}", fieldName, bad, value);
                throw new IllegalArgumentException("非法的SQL片段[" + fieldName + "]: 含危险构造 " + bad);
            }
        }
        // 单引号必须成对，避免闭合引号后注入
        if (countOccurrences(value, '\'') % 2 != 0) {
            log.warn("SQL片段校验失败[{}]: 单引号未成对 -> {}", fieldName, value);
            throw new IllegalArgumentException("非法的SQL片段[" + fieldName + "]: 引号不匹配");
        }
        // 括号必须配平，避免构造子查询逃逸
        if (countOccurrences(value, '(') != countOccurrences(value, ')')) {
            log.warn("SQL片段校验失败[{}]: 括号不配平 -> {}", fieldName, value);
            throw new IllegalArgumentException("非法的SQL片段[" + fieldName + "]: 括号不配平");
        }
        return value;
    }

    private static int countOccurrences(String s, char c) {
        int n = 0;
        for (int i = 0; i < s.length(); i++) {
            if (s.charAt(i) == c) {
                n++;
            }
        }
        return n;
    }
}
