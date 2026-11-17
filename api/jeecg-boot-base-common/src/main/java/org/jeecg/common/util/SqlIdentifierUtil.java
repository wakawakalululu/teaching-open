package org.jeecg.common.util;

import java.util.Arrays;
import java.util.HashSet;
import java.util.Set;
import java.util.regex.Pattern;

import lombok.extern.slf4j.Slf4j;

/**
 * SQL 标识符（表名 / 列名）安全校验工具。
 *
 * <p>背景：MyBatis 的 <code>${}</code> 是字符串直接拼接，不做预编译参数绑定。
 * 项目中存在 <code>select ${text} from ${table}</code> 这类动态表字典查询，
 * 若直接把外部输入拼进 SQL，即为典型的 SQL 注入。
 *
 * <p>因此对进入 <code>${}</code> 的标识符做严格白名单校验：
 * <ul>
 *   <li>只允许字母、数字、下划线，且必须以字母或下划线开头；</li>
 *   <li>长度上限 64，避免超长畸形输入；</li>
 *   <li>拒绝 SQL 保留字、注释符、引号、分号、括号、空白等一切可构造注入的字符；</li>
 *   <li>可选支持 <code>库名.表名</code> 形式，逐段校验。</li>
 * </ul>
 *
 * <p>本类只做"拒绝"，不做"清洗"：一旦不合法直接抛异常，避免静默篡改用户输入后
 * 仍拼进 SQL 造成逻辑歧义。
 *
 * @author security-hardening
 */
@Slf4j
public class SqlIdentifierUtil {

    /** 合法标识符：字母/下划线开头，后续为字母、数字、下划线 */
    private static final Pattern IDENTIFIER_PATTERN = Pattern.compile("^[A-Za-z_][A-Za-z0-9_]{0,63}$");

    /** SQL 保留字与危险关键字黑名单（小写比较），进一步收紧 allowlist 命中面 */
    private static final Set<String> RESERVED_WORDS = new HashSet<String>(Arrays.asList(
            "select", "insert", "update", "delete", "drop", "truncate", "alter", "create",
            "grant", "revoke", "union", "exec", "execute", "declare", "into", "outfile",
            "dumpfile", "load_file", "information_schema", "mysql", "performance_schema",
            "sleep", "benchmark", "extractvalue", "updatexml", "procedure", "analyse",
            "and", "or", "xor", "not", "case", "when", "then", "else", "end",
            "where", "from", "join", "on", "order", "group", "by", "having", "limit",
            "script", "javascript", "iframe", "onerror", "onload"
    ));

    private SqlIdentifierUtil() {
    }

    /**
     * 校验单个 SQL 标识符，不合法立即抛异常。
     *
     * @param identifier 待校验的表名或列名
     * @param fieldName  出错的字段名，仅用于异常信息定位
     * @return 原样返回校验通过的标识符，便于链式使用
     * @throws IllegalArgumentException 校验失败
     */
    public static String assertIdentifier(String identifier, String fieldName) {
        if (identifier == null || identifier.trim().isEmpty()) {
            throw new IllegalArgumentException("非法的SQL标识符[" + fieldName + "]: 不能为空");
        }
        String value = identifier.trim();
        // 仅允许一段，或 库名.表名 两段
        String[] segments = value.split("\\.", -1);
        if (segments.length > 2) {
            log.warn("SQL标识符校验失败[{}]: 段数过多 -> {}", fieldName, value);
            throw new IllegalArgumentException("非法的SQL标识符[" + fieldName + "]: " + value);
        }
        for (String segment : segments) {
            if (!IDENTIFIER_PATTERN.matcher(segment).matches()) {
                log.warn("SQL标识符校验失败[{}]: 不符合命名规则 -> {}", fieldName, value);
                throw new IllegalArgumentException("非法的SQL标识符[" + fieldName + "]: " + value);
            }
            if (RESERVED_WORDS.contains(segment.toLowerCase())) {
                log.warn("SQL标识符校验失败[{}]: 命中保留字 -> {}", fieldName, value);
                throw new IllegalArgumentException("非法的SQL标识符[" + fieldName + "]: 禁止使用保留字 " + segment);
            }
        }
        return value;
    }

    /**
     * 校验标识符是否属于给定的白名单集合（大小写不敏感）。
     * <p>用于列名这类取值可枚举的场景，比纯正则更强。
     *
     * @param identifier 待校验列名
     * @param allowed    允许的列名集合
     * @param fieldName  出错字段名
     * @return 校验通过的列名
     */
    public static String assertInWhitelist(String identifier, Set<String> allowed, String fieldName) {
        String value = assertIdentifier(identifier, fieldName);
        if (allowed == null || allowed.isEmpty()) {
            throw new IllegalArgumentException("非法的SQL标识符[" + fieldName + "]: 白名单为空");
        }
        for (String item : allowed) {
            if (item.equalsIgnoreCase(value)) {
                return value;
            }
        }
        log.warn("SQL标识符白名单校验失败[{}]: {} 不在允许集合内", fieldName, value);
        throw new IllegalArgumentException("非法的SQL标识符[" + fieldName + "]: " + value);
    }

    /**
     * 判断是否为合法标识符，不抛异常，供需要自行处理错误的场景使用。
     */
    public static boolean isValidIdentifier(String identifier) {
        if (identifier == null || identifier.trim().isEmpty()) {
            return false;
        }
        String[] segments = identifier.trim().split("\\.", -1);
        if (segments.length > 2) {
            return false;
        }
        for (String segment : segments) {
            if (!IDENTIFIER_PATTERN.matcher(segment).matches()) {
                return false;
            }
            if (RESERVED_WORDS.contains(segment.toLowerCase())) {
                return false;
            }
        }
        return true;
    }
}
