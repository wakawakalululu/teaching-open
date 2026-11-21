package usercore

import (
	"encoding/json"
	"testing"
)

// TestSplitIDs 逗号拆分语义：去空白、丢弃空段（前端 ids 普遍带尾逗号，
// Java split 的尾空段在 IN 列表里恒不命中，丢弃后语义一致）。
func TestSplitIDs(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"常规", "a,b", []string{"a", "b"}},
		{"尾逗号", "a,b,", []string{"a", "b"}},
		{"带空白", " a , b ", []string{"a", "b"}},
		{"全空", ",, ", nil},
		{"空串", "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitIDs(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("SplitIDs(%q) = %v, want %v", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("SplitIDs(%q) = %v, want %v", tc.in, got, tc.want)
				}
			}
		})
	}
}

// TestParseFilterRule QueryGenerator 字符串规则子集（convert2Rule + replaceValue）：
// * 通配 → LIKE/LEFT_LIKE/RIGHT_LIKE、逗号 → IN、! 前缀 → NE、其余 EQ。
func TestParseFilterRule(t *testing.T) {
	cases := []struct {
		in    string
		op    string
		value string
	}{
		{"*张*", "LIKE", "张"},
		{"*张", "LEFT_LIKE", "张"},
		{"张*", "RIGHT_LIKE", "张"},
		{"a,b", "IN", "a,b"},
		{"!admin", "<>", "admin"},
		{"plain", "=", "plain"},
		{"", "=", ""},
	}
	for _, tc := range cases {
		got := parseFilterRule(tc.in)
		if got.Op != tc.op || got.Value != tc.value {
			t.Fatalf("parseFilterRule(%q) = {%s %s}, want {%s %s}",
				tc.in, got.Op, got.Value, tc.op, tc.value)
		}
	}
}

// TestCamelToUnderline 对齐 oConvertUtils.camelToUnderline（排序列转换）。
func TestCamelToUnderline(t *testing.T) {
	cases := map[string]string{
		"createTime":   "create_time",
		"realname":     "realname",
		"departIds":    "depart_ids",
		"sex":          "sex",
		"userIdentity": "user_identity",
	}
	for in, want := range cases {
		if got := camelToUnderline(in); got != want {
			t.Fatalf("camelToUnderline(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestOrderColumns 排序白名单：来自实体 orm 标签的全量列，排除 password/salt；
// 必须覆盖前端 UserList.vue 实际会发送的排序键（username、sex_dictText 剥后缀
// 得 sex、mixin 默认 createTime）。
func TestOrderColumns(t *testing.T) {
	for _, col := range []string{"id", "username", "sex", "create_time", "depart_ids", "status"} {
		if _, ok := orderColumns[col]; !ok {
			t.Fatalf("排序列白名单缺少 %s", col)
		}
	}
	for _, col := range []string{"password", "salt"} {
		if _, ok := orderColumns[col]; ok {
			t.Fatalf("排序列白名单不应包含敏感列 %s", col)
		}
	}
}

// TestTranslateTableDict 表字典回译：逗号 id → 逗号名，未命中键不拼接，
// 空串输出空串（对齐 DictAspect.translateDictValue）。
func TestTranslateTableDict(t *testing.T) {
	names := map[string]string{"d1": "研发部", "d2": "教学部"}
	if got := translateTableDict("d1,d2", names); got != "研发部,教学部" {
		t.Fatalf("translateTableDict 多值 = %q", got)
	}
	if got := translateTableDict("d1,missing", names); got != "研发部" {
		t.Fatalf("translateTableDict 未命中键应跳过 = %q", got)
	}
	if got := translateTableDict("", names); got != "" {
		t.Fatalf("translateTableDict 空串应输出空串 = %q", got)
	}
}

// TestFormatDates 生日仅保留日期、时间仅保留到秒（Java @JsonFormat 与
// DictAspect 的 Date 转换）。
func TestFormatDates(t *testing.T) {
	if got := formatDate("2022-08-28 17:46:00"); got != "2022-08-28" {
		t.Fatalf("formatDate = %q", got)
	}
	if got := formatDate("2022-08-28"); got != "2022-08-28" {
		t.Fatalf("formatDate 短串应原样 = %q", got)
	}
	if got := formatDateTime("2022-08-28 17:46:00.123"); got != "2022-08-28 17:46:00" {
		t.Fatalf("formatDateTime = %q", got)
	}
}

// TestUserPayloadPointerSemantics 锁定 add/edit 请求体的指针语义：
// JSON 未传字段保持 nil（镜像 MyBatis-Plus updateById 非 NULL 更新策略），
// 密码/角色等敏感承载字段解析正确。
func TestUserPayloadPointerSemantics(t *testing.T) {
	raw := `{"id":"u1","username":"tom","password":"p@ss","sex":1,
		"selectedroles":"r1,r2","selecteddeparts":"d1,d2,"}`
	var in UserPayload
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if in.Id == nil || *in.Id != "u1" {
		t.Fatalf("id 解析错误: %v", in.Id)
	}
	if in.Sex == nil || *in.Sex != 1 {
		t.Fatalf("sex 解析错误: %v", in.Sex)
	}
	if in.OrgCode != nil {
		t.Fatalf("未传字段必须保持 nil（非 NULL 更新策略）: %v", *in.OrgCode)
	}
	if in.SelectedRoles != "r1,r2" || in.SelectedDeparts != "d1,d2," {
		t.Fatalf("selectedroles/selecteddeparts 解析错误: %q / %q", in.SelectedRoles, in.SelectedDeparts)
	}
}

// TestBizErrorMessages 锁定业务错误话术与 Java 契约逐字一致。
func TestBizErrorMessages(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"权限不足", ErrPermissionDenied, "权限不足"},
		{"未找到对应实体", ErrEntityNotFound, "未找到对应实体"},
		{"无负责部门", ErrNoDeptScope, "您没有负责的班级"},
		{"角色分配越权", ErrRoleForbidden, "权限不足，无法分配所选角色"},
		{"密码缺失", ErrPasswordRequired, "密码不能为空"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.want {
				t.Fatalf("话术漂移: got %q, want %q", tc.err.Error(), tc.want)
			}
		})
	}
}
