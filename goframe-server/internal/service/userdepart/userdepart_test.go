// Package userdepart 纯逻辑单测：不依赖任何真实数据库连接，
// 只覆盖 QueryGenerator 规则决策、列白名单、标识符转换、前端契约
// JSON 形状与通讯录合并逻辑。
package userdepart

import (
	"encoding/json"
	"testing"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// TestParseUserFilterRule 覆盖 Java QueryGenerator convert2Rule 子集语义：
// `*` 通配 → LIKE（前后缀位置由 * 决定）、逗号 → IN、`!=` 前缀 → <>、
// 其余 → 精确匹配；空串按精确空值处理（调用方对空值直接跳过）。
func TestParseUserFilterRule(t *testing.T) {
	cases := []struct {
		in   string
		want userFilterRule
	}{
		{in: "admin", want: userFilterRule{Op: "=", Value: "admin"}},
		{in: "  admin  ", want: userFilterRule{Op: "=", Value: "admin"}},
		{in: "*张*", want: userFilterRule{Op: "LIKE", Value: "%张%"}},
		{in: "*张", want: userFilterRule{Op: "LIKE", Value: "%张"}},
		{in: "张*", want: userFilterRule{Op: "LIKE", Value: "张%"}},
		{in: "a,b,c", want: userFilterRule{Op: "IN", Value: "a,b,c"}},
		{in: "!=admin", want: userFilterRule{Op: "<>", Value: "admin"}},
	}
	for _, c := range cases {
		got := parseUserFilterRule(c.in)
		if got != c.want {
			t.Fatalf("parseUserFilterRule(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

// TestCamelToUnderline 排序列驼峰转下划线（对齐 oConvertUtils.camelToUnderline）。
func TestCamelToUnderline(t *testing.T) {
	cases := map[string]string{
		"createTime":  "create_time",
		"createTime2": "create_time2",
		"OrgCode":     "org_code",
		"username":    "username",
		"realname":    "realname",
		"aBC":         "a_b_c",
	}
	for in, want := range cases {
		if got := camelToUnderline(in); got != want {
			t.Fatalf("camelToUnderline(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestSplitIDs 逗号拆分：兼容前端 deleteUserInDepartBatch 的尾逗号与空白项。
func TestSplitIDs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{in: "id1,id2,", want: []string{"id1", "id2"}},
		{in: " id1 , id2 ", want: []string{"id1", "id2"}},
		{in: "", want: nil},
		{in: ",", want: nil},
		{in: "only", want: []string{"only"}},
	}
	for _, c := range cases {
		got := SplitIDs(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("SplitIDs(%q) = %v, want %v", c.in, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("SplitIDs(%q) = %v, want %v", c.in, got, c.want)
			}
		}
	}
}

// TestSysUserOrderColumns 排序列白名单来自 dao.SysUser 列集合，
// 必须覆盖前端 JeecgListMixin 默认排序列 create_time。
func TestSysUserOrderColumns(t *testing.T) {
	for _, col := range []string{"create_time", "id", "username", "realname", "org_code"} {
		if _, ok := sysUserOrderColumns[col]; !ok {
			t.Fatalf("排序列白名单缺少 %s", col)
		}
	}
	// dao 列集合与反射构建结果一致
	cols := dao.SysUser.Columns()
	if _, ok := sysUserOrderColumns[cols.CreateTime]; !ok {
		t.Fatalf("排序列白名单缺少 %s", cols.CreateTime)
	}
	if _, ok := sysUserOrderColumns[cols.OrgCode]; !ok {
		t.Fatalf("排序列白名单缺少 %s", cols.OrgCode)
	}
}

// TestDepartUserRecordJSON departUserList 记录 JSON 契约：
// 内嵌 SysUser 平铺输出、sex_dictText 键存在、password/salt 永不输出、
// orgCode 被覆写为部门名。
func TestDepartUserRecordJSON(t *testing.T) {
	rec := DepartUserRecord{
		SysUser: entity.SysUser{
			Id:       "u1",
			Username: "admin",
			Realname: "管理员",
			Password: "should-not-leak",
			Salt:     "should-not-leak",
			OrgCode:  "部门一,部门二",
			Sex:      1,
		},
		SexDictText: "男",
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if _, ok := m["sex_dictText"]; !ok {
		t.Fatalf("缺少 sex_dictText 字段: %s", string(data))
	}
	if _, ok := m["username"]; !ok {
		t.Fatalf("SysUser 字段未平铺输出: %s", string(data))
	}
	if _, ok := m["password"]; ok {
		t.Fatalf("password 不允许输出: %s", string(data))
	}
	if _, ok := m["salt"]; ok {
		t.Fatalf("salt 不允许输出: %s", string(data))
	}
	if m["orgCode"] != "部门一,部门二" {
		t.Fatalf("orgCode 应被覆写为部门名: %v", m["orgCode"])
	}
}

// TestMergeAddressRecords 通讯录合并：同用户多部门 departName 以 " | "
// 拼接、id 更名为 userId、首条之外去重、顺序保持。
func TestMergeAddressRecords(t *testing.T) {
	list := []SysUserSysDepartModel{
		{Id: "u1", Realname: "张三", DepartId: "d1", DepartName: "研发部"},
		{Id: "u2", Realname: "李四", DepartId: "d2", DepartName: "教务部"},
		{Id: "u1", Realname: "张三", DepartId: "d3", DepartName: "市场部"},
	}
	merged := mergeAddressRecords(list)
	if len(merged) != 2 {
		t.Fatalf("合并后应只剩 2 条, got %d", len(merged))
	}
	if merged[0].UserId != "u1" || merged[0].DepartName != "研发部 | 市场部" {
		t.Fatalf("u1 合并结果不符: %+v", merged[0])
	}
	if merged[0].DepartId != "d1" {
		t.Fatalf("u1 应保留首条部门 id: %+v", merged[0])
	}
	if merged[1].UserId != "u2" || merged[1].DepartName != "教务部" {
		t.Fatalf("u2 结果不符: %+v", merged[1])
	}
}

// TestMergeAddressRecordsJSON 通讯录记录 JSON 字段集合对齐 Java
// JSON.toJSONString(SysUserSysDepartModel) 去掉 id 后的输出。
func TestMergeAddressRecordsJSON(t *testing.T) {
	rec := SysUserSysDepartAddressModel{
		Realname:   "张三",
		WorkNo:     "T001",
		Post:       "1,2",
		Telephone:  "0755-1234",
		Email:      "a@b.c",
		Phone:      "13800000000",
		DepartId:   "d1",
		DepartName: "研发部",
		UserId:     "u1",
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	for _, key := range []string{"userId", "realname", "workNo", "post", "telephone", "email", "phone", "departId", "departName"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("缺少契约字段 %s: %s", key, string(data))
		}
	}
	if _, ok := m["id"]; ok {
		t.Fatalf("通讯录记录不允许输出 id（应更名为 userId）: %s", string(data))
	}
}
