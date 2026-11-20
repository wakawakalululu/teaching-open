package dict

import "testing"

// 表字典动态标识符白名单是防 SQL 注入的核心闸门，
// 必须拒绝一切非 [a-zA-Z0-9_] 组成的表名/列名。
func TestValidIdent(t *testing.T) {
	valid := []string{"sys_user", "realname", "id", "T2", "a_b_c"}
	for _, s := range valid {
		if !ValidIdent(s) {
			t.Errorf("ValidIdent(%q) = false, want true", s)
		}
	}
	invalid := []string{
		"",                  // 空串
		"sys_user,realname", // 逗号
		"sys user",          // 空格
		"id; DROP TABLE x",  // 语句拼接
		"(SELECT 1)",        // 括号
		"id'--",             // 注释/引号
		"1 OR 1=1",          // 布尔注入
		"a`b",               // 反引号
	}
	for _, s := range invalid {
		if ValidIdent(s) {
			t.Errorf("ValidIdent(%q) = true, want false", s)
		}
	}
}
