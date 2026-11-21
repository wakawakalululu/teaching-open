package userrole

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// userRecord 是 userRoleList 的白名单输出契约，字段集合必须与
// 前端真实消费（RoleUserList.vue 表格/UserModal.edit 回填）+ Java
// DictAspect 注入的 status_dictText 完全一致，不得多透出敏感列
// （password/salt/createTime 等）。
func TestUserRecordJSONWhitelist(t *testing.T) {
	rec := newUserRecord(entity.SysUser{
		Id:           "u1",
		Username:     "admin",
		Realname:     "管理员",
		Password:     "should-not-leak",
		Salt:         "should-not-leak",
		Birthday:     "1990-01-02 03:04:05",
		Sex:          1,
		Email:        "a@b.c",
		Phone:        "13800000000",
		Status:       1,
		OrgCode:      "A01",
		ActivitiSync: 1,
		WorkNo:       "W01",
		Post:         "教师",
		School:       "测试学校",
		Telephone:    "010-1234567",
		UserIdentity: 2,
		DepartIds:    "d1,d2",
	}, map[string]string{"1": "正常"})

	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("marshal userRecord: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wantKeys := []string{
		"id", "username", "realname", "avatar", "birthday",
		"sex", "email", "phone", "status", "status_dictText",
		"orgCode", "activitiSync", "workNo", "post", "school",
		"telephone", "userIdentity", "departIds",
	}
	var gotKeys []string
	for k := range got {
		gotKeys = append(gotKeys, k)
	}
	if !reflect.DeepEqual(sortedStrings(gotKeys), sortedStrings(wantKeys)) {
		t.Errorf("userRecord JSON 字段集不匹配 got=%v want=%v", gotKeys, wantKeys)
	}
	if _, ok := got["password"]; ok {
		t.Errorf("password 不得出现在输出中")
	}
	if _, ok := got["salt"]; ok {
		t.Errorf("salt 不得出现在输出中")
	}
	// birthday 与 Java @JsonFormat(pattern="yyyy-MM-dd") 对齐，仅保留日期部分
	if rec.Birthday != "1990-01-02" {
		t.Errorf("Birthday = %q, want %q", rec.Birthday, "1990-01-02")
	}
	// status_dictText 注入（对齐 DictAspect @Dict(dicCode="user_status")）
	if rec.StatusDictText != "正常" {
		t.Errorf("StatusDictText = %q, want %q", rec.StatusDictText, "正常")
	}
}

// 字典未命中时 status_dictText 为空串（Java 为 null，前端展示同为空白）。
func TestNewUserRecordStatusDictMiss(t *testing.T) {
	rec := newUserRecord(entity.SysUser{Status: 9}, map[string]string{"1": "正常"})
	if rec.StatusDictText != "" {
		t.Errorf("StatusDictText = %q, want empty", rec.StatusDictText)
	}
}

// sortedStrings 返回升序副本，保证键集合比较不受 map 遍历顺序影响。
func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
