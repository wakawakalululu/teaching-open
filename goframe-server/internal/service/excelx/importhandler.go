// importstudent.go SysUser importStudent（Excel 批量导入学生）
// （契约蓝本：SysUserController#importStudent，POST /sys/user/importStudent）。
//
// Java 语义：multipart file + departIds（默认班级）；读 SysUser @Excel 列；
// username 为空跳过「第 N 行：账号为空，忽略导入。」；密码缺省 "123456"
// 加盐加密（Go 为 bcrypt，与既有域决策一致）；用户名已存在跳过、手机号
// 已存在跳过；student 角色绑定 + departIds（默认班级）关联。
package excelx

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// studentSpec 学生导入列（SysUser @Excel 消费面，Header=中文表头）。
var studentSpec = Spec{Cols: []ExcelCol{
	{"username", "登录账号"}, {"realname", "真实姓名"}, {"sex", "性别"},
	{"phone", "手机号"}, {"email", "邮箱"}, {"password", "密码"},
}}

// ImportStudentResult 导入结果（Java 返回 Result ok(message+result)）。
type ImportStudentResult struct {
	SuccessLines int
	ErrorLines   int
	Errors       []string
}

// ImportStudent 执行学生批量导入。
//
// departIDs 为默认班级（逗号分隔，departIds 参数）；规则：
//   - username 空 → 跳过并记「第 N 行：账号为空，忽略导入。」；
//   - 用户名已存在 → 跳过「第 N 行：用户名已经存在，忽略导入。」；
//   - 手机号已存在 → 跳过「第 N 行：手机号已经存在，忽略导入。」；
//   - 密码：Excel 提供则用（bcrypt），缺省 "123456"（bcrypt 化——Java 为
//     DES 加盐，Go 统一 bcrypt 为既定决策）；
//   - student 角色绑定 + 默认班级 sys_user_depart 关联。
func ImportStudent(ctx context.Context, fileData []byte, departIDs string, bcryptHash func(string) string) (*ImportStudentResult, error) {
	headers, rows, err := Parse(fileData, 2, 1)
	if err != nil {
		return nil, err
	}
	res := &ImportStudentResult{}
	sc := dao.SysRole.Columns()
	var studentRole entity.SysRole
	if err := g.DB().Model("sys_role").Where(sc.RoleCode, "student").
		Limit(1).Scan(&studentRole); err != nil {
		return nil, fmt.Errorf("查询学生角色失败: %w", err)
	}
	if studentRole.Id == "" {
		return nil, fmt.Errorf("student角色不存在")
	}
	for i, values := range rows {
		lineNo := i + 4 // 2 标题 + 1 表头之后的行号（1-based，含表头行）
		get := func(header string) string {
			for j, h := range headers {
				if strings.TrimSpace(h) == header && j < len(values) {
					return strings.TrimSpace(values[j])
				}
			}
			return ""
		}
		username := get("登录账号")
		if username == "" {
			res.ErrorLines++
			res.Errors = append(res.Errors, fmt.Sprintf("第 %d 行：账号为空，忽略导入。", lineNo))
			continue
		}
		// 用户名查重
		if cnt, err := g.DB().Model("sys_user").
			Where("username", username).Where("del_flag", 0).Count(); err == nil && cnt > 0 {
			res.ErrorLines++
			res.Errors = append(res.Errors, fmt.Sprintf("第 %d 行：用户名已经存在，忽略导入。", lineNo))
			continue
		}
		phone := get("手机号")
		if phone != "" {
			if cnt, err := g.DB().Model("sys_user").
				Where("phone", phone).Where("del_flag", 0).Count(); err == nil && cnt > 0 {
				res.ErrorLines++
				res.Errors = append(res.Errors, fmt.Sprintf("第 %d 行：手机号已经存在，忽略导入。", lineNo))
				continue
			}
		}
		password := get("密码")
		if password == "" {
			password = "123456"
		}
		userID := guid.S()
		if _, err := g.DB().Model("sys_user").Data(map[string]interface{}{
			"id": userID, "username": username,
			"realname": get("真实姓名"), "password": bcryptHash(password),
			"phone": phone, "email": get("邮箱"),
			"status": 1, "del_flag": 0, "activiti_sync": 1,
			"create_time": gtime.Now(),
		}).Insert(); err != nil {
			res.ErrorLines++
			res.Errors = append(res.Errors, fmt.Sprintf("第 %d 行：%v", lineNo, err))
			continue
		}
		// 学生角色绑定
		if _, err := g.DB().Model("sys_user_role").Data(map[string]interface{}{
			"id": guid.S(), "user_id": userID, "role_id": studentRole.Id,
		}).Insert(); err != nil {
			return nil, fmt.Errorf("绑定学生角色失败: %w", err)
		}
		// 默认班级关联
		for _, depID := range strings.Split(departIDs, ",") {
			depID = strings.TrimSpace(depID)
			if depID == "" {
				continue
			}
			_, _ = g.DB().Model("sys_user_depart").Data(map[string]interface{}{
				"id": guid.S(), "user_id": userID, "dep_id": depID,
			}).Insert()
		}
		res.SuccessLines++
	}
	return res, nil
}
