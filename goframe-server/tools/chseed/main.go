// chseed 春晖青少年编程中心种子数据导入工具
// 连接数据库后替换 Mock 数据为春晖真实数据。
//
// 产品归属：武胜县沿口春夏秋信息技术中心（个体工商户）
// 统一社会信用代码：92511622MAKLLHTT5D  成立日期：2026-09-07  登记状态：存续
// 经营场所：四川省广安市武胜县沿口镇上东街257号（邮编638400）
// 页面布局参考：verveit.010124.xyz
//
// 密码：admin→admin123  teacher1/2/3→teacher123  student1-6→student123（bcrypt $2a$12）。
package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/genv"
)

// 品牌常量（与 api/db/chunhui_seed_base.sql / _demo.sql 保持一致）。
const (
	brandName   = "春晖青少年编程中心"
	ownerName   = "武胜县沿口春夏秋信息技术中心（个体工商户）"
	creditCode  = "92511622MAKLLHTT5D"
	fullAddress = "四川省广安市武胜县沿口镇上东街257号（邮编638400）"
	copyRight   = "© 2026 武胜县沿口春夏秋信息技术中心（个体工商户）· 统一社会信用代码 92511622MAKLLHTT5D"
)

func main() {
	layer := flag.String("layer", "all", "种子层级：base（系统表）| demo（业务示例）| all（两者）")
	flag.Parse()

	// 端口约定：MariaDB 13306（高端口防冲突）；可经环境变量 MYSQL_LINK 覆盖。
	link := os.Getenv("MYSQL_LINK")
	if link == "" {
		link = "mysql:teaching:Teaching2026!@tcp(127.0.0.1:13306)/teachingopen?parseTime=true&loc=Local"
	}
	genv.Set("MYSQL_LINK", link)
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{Link: genv.Get("MYSQL_LINK").String(), Type: "mysql"},
		},
	})

	if err := g.DB().PingMaster(); err != nil {
		fmt.Println("数据库连接失败:", err)
		os.Exit(1)
	}
	fmt.Println("数据库连接正常，层级 =", *layer)

	switch strings.ToLower(*layer) {
	case "base":
		seedBaseLayer()
	case "demo":
		seedDemoLayer()
	case "all", "":
		seedBaseLayer()
		seedDemoLayer()
	default:
		fmt.Println("未知层级:", *layer, "（可选 base|demo|all）")
		os.Exit(1)
	}

	fmt.Println("✓ 春晖青少年编程中心种子数据导入完成（layer =", *layer, "）")
}

// baseTables BASE 层清除范围：仅系统表（登录与后台运营所必需的数据）。
var baseTables = []string{
	"sys_config", "sys_depart", "sys_user", "sys_user_role", "sys_user_depart",
	"sys_position", "sys_data_source", "sys_sms_template", "sys_depart_role",
}

// demoTables DEMO 层清除范围：仅业务示例表。
var demoTables = []string{
	"teaching_course", "teaching_course_unit", "teaching_course_dept",
	"sys_announcement", "sys_announcement_send", "teaching_news",
	"teaching_work", "teaching_work_comment",
}

// seedBaseLayer 系统表层：组织、账号、授权、后台基础数据、菜单可见性。
func seedBaseLayer() {
	for _, t := range baseTables {
		if _, err := g.DB().Exec(context.Background(), "DELETE FROM "+t); err != nil {
			fmt.Printf("清除 %s 失败: %v\n", t, err)
		}
	}
	if _, err := g.DB().Exec(context.Background(),
		"DELETE FROM sys_role_permission WHERE role_id = ?", "ch_role_admin"); err != nil {
		fmt.Printf("清除管理员授权失败: %v\n", err)
	}
	fmt.Println("BASE 层旧数据已清除")

	seedDeparts()
	seedRoles()
	seedRolePermissions()
	seedUsers()
	seedUserRoles()
	seedUserDeparts()
	seedConfigs()
	seedAdminBase()
	hideUnavailableMenus()
}

// seedDemoLayer 业务示例层：课程/单元/公告/资讯/作品/评论 + 二进制资源。
func seedDemoLayer() {
	for _, t := range demoTables {
		if _, err := g.DB().Exec(context.Background(), "DELETE FROM "+t); err != nil {
			fmt.Printf("清除 %s 失败: %v\n", t, err)
		}
	}
	if _, err := g.DB().Exec(context.Background(),
		"DELETE FROM sys_file WHERE id LIKE ? OR id LIKE ?", "ch_file_%", "ch_work_%_file"); err != nil {
		fmt.Printf("清除示例文件登记失败: %v\n", err)
	}
	fmt.Println("DEMO 层旧数据已清除")

	seedCourses()
	seedUnits()
	seedCourseDepts()
	seedAnnouncements()
	seedAnnouncementSends()
	seedNews()
	seedFiles()
	seedMaps()
	seedWorks()
	seedWorkComments()
}

// hideUnavailableMenus 隐藏本部署未实现的模块菜单（对齐 chunhui_seed_base.sql §15）。
func hideUnavailableMenus() {
	const stmt = `UPDATE sys_permission SET hidden = 1
 WHERE hidden = 0
   AND (url LIKE '/online/%'
        OR url LIKE '/modules/bpm/%'
        OR url LIKE '/modules/extbpm/%'
        OR url = '/exception/404')`
	res, err := g.DB().Exec(context.Background(), stmt)
	if err != nil {
		fmt.Printf("隐藏未启用菜单失败: %v\n", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		fmt.Printf("未启用模块菜单已隐藏 %d 条\n", n)
	}
}

// seedDeparts 部门/班级：春晖中心 → 教学部 + 三个班。
func seedDeparts() {
	deps := []map[string]interface{}{
		{"id": "ch_depart_01", "parent_id": "", "depart_name": brandName, "depart_name_en": "Chunhui Youth Coding Center",
			"depart_name_abbr": "春晖编程", "depart_order": 0, "description": ownerName,
			"org_category": "1", "org_type": "1", "org_code": "A01", "address": "四川省广安市武胜县沿口镇上东街257号",
			"memo": "统一社会信用代码：" + creditCode, "status": "0", "del_flag": "0", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_depart_02", "parent_id": "ch_depart_01", "depart_name": "教学部", "depart_order": 1,
			"description": "课程教学与教研", "org_category": "2", "org_type": "2", "org_code": "A01A01",
			"status": "0", "del_flag": "0", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_depart_03", "parent_id": "ch_depart_01", "depart_name": "Scratch启蒙班", "depart_order": 2,
			"description": "Scratch 图形化编程（6-9岁）", "org_category": "3", "org_type": "3", "org_code": "A01A02",
			"address": "上东街257号 2楼", "status": "0", "del_flag": "0", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_depart_04", "parent_id": "ch_depart_01", "depart_name": "Python进阶班", "depart_order": 3,
			"description": "Python 代码编程（10-14岁）", "org_category": "3", "org_type": "3", "org_code": "A01A03",
			"address": "上东街257号 3楼", "status": "0", "del_flag": "0", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_depart_05", "parent_id": "ch_depart_01", "depart_name": "机器人竞赛班", "depart_order": 4,
			"description": "机器人搭建与竞赛", "org_category": "3", "org_type": "3", "org_code": "A01A04",
			"address": "上东街257号 1楼", "status": "0", "del_flag": "0", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
	}
	for _, d := range deps {
		if _, err := g.DB().Model("sys_depart").Data(d).Save(); err != nil {
			fmt.Printf("插入部门失败: %v\n", err)
		}
	}
	fmt.Println("部门 5 条已插入")
}

// seedRoles 角色。
func seedRoles() {
	roles := []map[string]interface{}{
		{"id": "ch_role_admin", "role_name": "管理员", "role_code": "admin", "description": brandName + "管理员", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "role_level": 99},
		{"id": "ch_role_dev", "role_name": "开发调试", "role_code": "dev", "description": "系统开发调试", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "role_level": 98},
		{"id": "ch_role_teacher", "role_name": "教师", "role_code": "teacher", "description": "授课教师", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "role_level": 50},
		{"id": "ch_role_student", "role_name": "学生", "role_code": "student", "description": "注册学生", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "role_level": 0},
	}
	for _, r := range roles {
		g.DB().Model("sys_role").Data(r).Save()
	}
	fmt.Println("角色 4 条已插入")
}

// seedRolePermissions 管理员角色绑定全部菜单权限。
// 对齐 Java 演示数据模式：sys_role_permission 是角色→菜单授权唯一来源，
// QueryByUser 严格按该表连接过滤（无 admin 旁路），不绑定则动态管理菜单为空。
func seedRolePermissions() {
	if _, err := g.DB().Model("sys_role_permission").Where("role_id", "ch_role_admin").Delete(); err != nil {
		fmt.Printf("清除旧管理员授权失败: %v\n", err)
	}
	perms, err := g.DB().Model("sys_permission").Fields("id").All()
	if err != nil {
		fmt.Printf("读取菜单权限失败: %v\n", err)
		return
	}
	n := 0
	for i, row := range perms {
		pid := row["id"].String()
		if pid == "" {
			continue
		}
		if _, err := g.DB().Model("sys_role_permission").Data(map[string]interface{}{
			"id":            fmt.Sprintf("ch_rp_%04d", i),
			"role_id":       "ch_role_admin",
			"permission_id": pid,
		}).Save(); err != nil {
			fmt.Printf("插入授权 %s 失败: %v\n", pid, err)
			continue
		}
		n++
	}
	fmt.Printf("管理员授权 %d 条已插入\n", n)
}

// seedUsers 用户：1 管理员 + 3 教师 + 6 学生。
func seedUsers() {
	bcryptAdmin := "$2a$12$LWLf.Ti0FX/ajbsEVuUtC.227qHk3IryExS9MHIvfQJ6GuvbYHY3G"   // admin123
	bcryptTeacher := "$2a$12$N2gprWtzbRX/axPrmQ6lg.IL16PeAyflNXSSUL/gkmrWOaJ4NDvXK" // teacher123
	bcryptStudent := "$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC" // student123
	users := []map[string]interface{}{
		{"id": "ch_user_admin", "username": "admin", "realname": "春晖管理员", "password": bcryptAdmin, "salt": "", "sex": 1, "org_code": "A01", "status": 1, "del_flag": 0, "activiti_sync": 1, "work_no": "CH001", "post": "校长", "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 2, "depart_ids": "ch_depart_01,ch_depart_02"},
		{"id": "ch_user_teacher1", "username": "teacher1", "realname": "王老师", "password": bcryptTeacher, "salt": "", "sex": 1, "birthday": "1995-03-15", "org_code": "A01", "status": 1, "del_flag": 0, "activiti_sync": 1, "work_no": "CH002", "post": "Scratch讲师", "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 2, "depart_ids": "ch_depart_02,ch_depart_03"},
		{"id": "ch_user_teacher2", "username": "teacher2", "realname": "李老师", "password": bcryptTeacher, "salt": "", "sex": 2, "birthday": "1998-08-22", "org_code": "A01", "status": 1, "del_flag": 0, "activiti_sync": 1, "work_no": "CH003", "post": "Python讲师", "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 2, "depart_ids": "ch_depart_02,ch_depart_04"},
		{"id": "ch_user_teacher3", "username": "teacher3", "realname": "陈教练", "password": bcryptTeacher, "salt": "", "sex": 1, "birthday": "1992-12-05", "org_code": "A01", "status": 1, "del_flag": 0, "activiti_sync": 1, "work_no": "CH004", "post": "机器人教练", "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 2, "depart_ids": "ch_depart_02,ch_depart_05"},
		{"id": "ch_user_stu1", "username": "student1", "realname": "张小明", "password": bcryptStudent, "salt": "", "sex": 1, "birthday": "2018-05-20", "org_code": "A01A02", "status": 1, "del_flag": 0, "activiti_sync": 1, "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 1, "depart_ids": "ch_depart_03"},
		{"id": "ch_user_stu2", "username": "student2", "realname": "李小红", "password": bcryptStudent, "salt": "", "sex": 2, "birthday": "2019-01-10", "org_code": "A01A02", "status": 1, "del_flag": 0, "activiti_sync": 1, "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 1, "depart_ids": "ch_depart_03"},
		{"id": "ch_user_stu3", "username": "student3", "realname": "陈大壮", "password": bcryptStudent, "salt": "", "sex": 1, "birthday": "2016-09-01", "org_code": "A01A03", "status": 1, "del_flag": 0, "activiti_sync": 1, "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 1, "depart_ids": "ch_depart_04"},
		{"id": "ch_user_stu4", "username": "student4", "realname": "刘小慧", "password": bcryptStudent, "salt": "", "sex": 2, "birthday": "2017-04-18", "org_code": "A01A03", "status": 1, "del_flag": 0, "activiti_sync": 1, "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 1, "depart_ids": "ch_depart_04"},
		{"id": "ch_user_stu5", "username": "student5", "realname": "周小杰", "password": bcryptStudent, "salt": "", "sex": 1, "birthday": "2015-07-12", "org_code": "A01A04", "status": 1, "del_flag": 0, "activiti_sync": 1, "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 1, "depart_ids": "ch_depart_05"},
		{"id": "ch_user_stu6", "username": "student6", "realname": "吴小婷", "password": bcryptStudent, "salt": "", "sex": 2, "birthday": "2016-02-25", "org_code": "A01A04", "status": 1, "del_flag": 0, "activiti_sync": 1, "school": brandName, "create_by": "admin", "create_time": "2026-09-07 00:00:00", "user_identity": 1, "depart_ids": "ch_depart_05"},
	}
	for _, u := range users {
		if _, err := g.DB().Model("sys_user").Data(u).Save(); err != nil {
			fmt.Printf("插入用户 %v 失败: %v\n", u["username"], err)
		}
	}
	fmt.Println("用户 10 条已插入")
}

// seedUserRoles 用户-角色。
func seedUserRoles() {
	urs := []map[string]interface{}{
		{"id": "ch_ur_01", "user_id": "ch_user_admin", "role_id": "ch_role_admin"},
		{"id": "ch_ur_02", "user_id": "ch_user_teacher1", "role_id": "ch_role_teacher"},
		{"id": "ch_ur_03", "user_id": "ch_user_teacher2", "role_id": "ch_role_teacher"},
		{"id": "ch_ur_04", "user_id": "ch_user_teacher3", "role_id": "ch_role_teacher"},
		{"id": "ch_ur_05", "user_id": "ch_user_stu1", "role_id": "ch_role_student"},
		{"id": "ch_ur_06", "user_id": "ch_user_stu2", "role_id": "ch_role_student"},
		{"id": "ch_ur_07", "user_id": "ch_user_stu3", "role_id": "ch_role_student"},
		{"id": "ch_ur_08", "user_id": "ch_user_stu4", "role_id": "ch_role_student"},
		{"id": "ch_ur_09", "user_id": "ch_user_stu5", "role_id": "ch_role_student"},
		{"id": "ch_ur_10", "user_id": "ch_user_stu6", "role_id": "ch_role_student"},
	}
	for _, ur := range urs {
		g.DB().Model("sys_user_role").Data(ur).Save()
	}
	fmt.Println("用户角色 10 条已插入")
}

// seedUserDeparts 用户-班级。
func seedUserDeparts() {
	uds := []map[string]interface{}{
		{"id": "ch_ud_01", "user_id": "ch_user_admin", "dep_id": "ch_depart_01"},
		{"id": "ch_ud_02", "user_id": "ch_user_teacher1", "dep_id": "ch_depart_02"},
		{"id": "ch_ud_03", "user_id": "ch_user_teacher1", "dep_id": "ch_depart_03"},
		{"id": "ch_ud_04", "user_id": "ch_user_teacher2", "dep_id": "ch_depart_02"},
		{"id": "ch_ud_05", "user_id": "ch_user_teacher2", "dep_id": "ch_depart_04"},
		{"id": "ch_ud_06", "user_id": "ch_user_teacher3", "dep_id": "ch_depart_02"},
		{"id": "ch_ud_07", "user_id": "ch_user_teacher3", "dep_id": "ch_depart_05"},
		{"id": "ch_ud_08", "user_id": "ch_user_stu1", "dep_id": "ch_depart_03"},
		{"id": "ch_ud_09", "user_id": "ch_user_stu2", "dep_id": "ch_depart_03"},
		{"id": "ch_ud_10", "user_id": "ch_user_stu3", "dep_id": "ch_depart_04"},
		{"id": "ch_ud_11", "user_id": "ch_user_stu4", "dep_id": "ch_depart_04"},
		{"id": "ch_ud_12", "user_id": "ch_user_stu5", "dep_id": "ch_depart_05"},
		{"id": "ch_ud_13", "user_id": "ch_user_stu6", "dep_id": "ch_depart_05"},
	}
	for _, ud := range uds {
		g.DB().Model("sys_user_depart").Data(ud).Save()
	}
	fmt.Println("用户班级 13 条已插入")
}

// indexHtml 首页富文本（参考 verveit.010124.xyz 布局：居中主视觉 + 三卡片 + 备案脚注）。
const indexHtml = `<div style="max-width:1200px;margin:0 auto;padding:40px 20px;font-family:Microsoft YaHei,sans-serif;">
<div style="text-align:center;margin-bottom:30px;">
<h1 style="color:#2c3e50;font-size:36px;margin-bottom:10px;">` + brandName + `</h1>
<p style="color:#7f8c8d;font-size:16px;">Scratch 创意编程 · Python 代码进阶 · 机器人竞赛</p>
<p style="color:#95a5a6;font-size:14px;">让每个孩子拥有创造未来的能力</p></div>
<div style="display:flex;justify-content:space-around;text-align:center;flex-wrap:wrap;">
<div style="width:30%;background:#f8f9fa;border-radius:8px;padding:20px;margin:10px;">
<h3 style="color:#3498db;">Scratch 启蒙</h3><p style="color:#666;">6-9岁 · 图形化编程<br>培养逻辑思维</p></div>
<div style="width:30%;background:#f8f9fa;border-radius:8px;padding:20px;margin:10px;">
<h3 style="color:#e67e22;">Python 进阶</h3><p style="color:#666;">10-14岁 · 代码编程<br>掌握算法思维</p></div>
<div style="width:30%;background:#f8f9fa;border-radius:8px;padding:20px;margin:10px;">
<h3 style="color:#e74c3c;">机器人竞赛</h3><p style="color:#666;">8-14岁 · 硬件编程<br>锻炼动手能力</p></div>
</div>
<div style="text-align:center;margin-top:30px;padding:15px;background:#f0f3f7;border-radius:8px;">
<p style="color:#666;font-size:14px;">` + ownerName + `<br>统一社会信用代码：` + creditCode + `<br>经营场所：` + fullAddress + `</p></div></div>`

// homeHtml 自定义首页（渐变卡片布局）。
const homeHtml = `<div style="max-width:1200px;margin:0 auto;padding:30px 20px;font-family:Microsoft YaHei,sans-serif;">
<h2 style="text-align:center;color:#2c3e50;">欢迎来到` + brandName + `</h2>
<p style="text-align:center;color:#7f8c8d;font-size:18px;margin:20px 0;">让每个孩子拥有创造未来的能力</p>
<div style="display:flex;flex-wrap:wrap;justify-content:center;gap:20px;">
<div style="width:350px;background:linear-gradient(135deg,#667eea,#764ba2);border-radius:12px;padding:25px;color:white;">
<h3>Scratch 创意编程</h3><p>图形化积木式编程，适合 6-9 岁零基础学员<br>培养逻辑思维与创造力</p></div>
<div style="width:350px;background:linear-gradient(135deg,#f093fb,#f5576c);border-radius:12px;padding:25px;color:white;">
<h3>Python 代码进阶</h3><p>真正的代码编程，适合 10-14 岁学员<br>从游戏到人工智能</p></div>
<div style="width:350px;background:linear-gradient(135deg,#4facfe,#00f2fe);border-radius:12px;padding:25px;color:white;">
<h3>机器人竞赛</h3><p>硬件编程与竞赛训练，适合 8-14 岁<br>锻炼动手与团队协作</p></div>
</div></div>`

// workShareHtml 作品分享页头。
const workShareHtml = `<div style="max-width:1000px;margin:0 auto;padding:20px;font-family:Microsoft YaHei,sans-serif;">
<h2 style="text-align:center;color:#2c3e50;">` + brandName + ` · 作品展示</h2>
<p style="text-align:center;color:#95a5a6;">每一个作品都是孩子创造力的结晶</p></div>`

// seedConfigs 站点配置（含富文本页面布局）。
func seedConfigs() {
	cfgs := []map[string]interface{}{
		{"id": "ch_cfg_01", "config_key": "brandName", "config_value": brandName, "config_enabled": 1, "comment": "品牌名称"},
		{"id": "ch_cfg_02", "config_key": "_address", "config_value": fullAddress, "config_enabled": 1, "comment": "经营场所地址"},
		{"id": "ch_cfg_03", "config_key": "_phone", "config_value": "", "config_enabled": 1, "comment": "联系电话"},
		{"id": "ch_cfg_04", "config_key": "_linkman", "config_value": "春晖管理员", "config_enabled": 1, "comment": "联系人"},
		{"id": "ch_cfg_05", "config_key": "_defaultRole", "config_value": "student", "config_enabled": 1, "comment": "注册默认角色"},
		{"id": "ch_cfg_06", "config_key": "_defaultDepart", "config_value": "ch_depart_03", "config_enabled": 1, "comment": "注册默认班级"},
		{"id": "ch_cfg_07", "config_key": "_copyRight", "config_value": copyRight, "config_enabled": 1, "comment": "版权信息"},
		{"id": "ch_cfg_08", "config_key": "_indexHtml", "config_value": indexHtml, "config_enabled": 1, "comment": "首页富文本"},
		{"id": "ch_cfg_09", "config_key": "_homeHtml", "config_value": homeHtml, "config_enabled": 1, "comment": "自定义首页"},
		{"id": "ch_cfg_10", "config_key": "_workShareHtml", "config_value": workShareHtml, "config_enabled": 1, "comment": "作品分享页富文本"},
	}
	for _, c := range cfgs {
		g.DB().Model("sys_config").Data(c).Save()
	}
	fmt.Println("配置 10 条已插入")
}

// seedCourses 课程体系（6 门，参考 verveit.010124.xyz 的课程卡片布局）。
func seedCourses() {
	courses := []map[string]interface{}{
		{"id": "ch_course_01", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "Scratch 创意编程启蒙",
			"course_desc": `<h2>Scratch 图形化编程启蒙课程</h2><p>适合 6-9 岁零基础学员。通过 MIT 开发的 Scratch 图形化编程工具，孩子可以像搭积木一样拖拽代码块，创作属于自己的互动故事、动画和小游戏。</p><h3>课程亮点</h3><ul><li>零门槛入门，无需打字基础</li><li>项目式学习，每节课完成一个作品</li><li>培养逻辑思维、创造力与问题解决能力</li><li>小班教学（8人/班），确保每个孩子得到关注</li></ul><h3>适合年龄</h3><p>6-9 岁（小学 1-3 年级）</p>`,
			"show_type":   1, "course_map": mapThemeOf("ch_course_01").fileID, "is_shared": 0, "show_home": 1, "order_num": 1, "depart_ids": "ch_depart_03", "course_type": "1", "course_category": "1"},
		{"id": "ch_course_02", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "Python 代码编程进阶",
			"course_desc": `<h2>Python 代码编程进阶课程</h2><p>适合 10-14 岁有一定编程基础的学员。从图形化编程过渡到真正的代码编程，学习 Python 语法、数据结构、算法与项目开发。</p><h3>课程亮点</h3><ul><li>从积木到代码的平滑过渡</li><li>涵盖 Turtle绘图、Pygame游戏开发、数据分析入门</li><li>培养算法思维与工程能力</li><li>完成多个实战项目</li></ul><h3>适合年龄</h3><p>10-14 岁（小学 4 年级 - 初中）</p>`,
			"show_type":   1, "course_map": mapThemeOf("ch_course_02").fileID, "is_shared": 0, "show_home": 1, "order_num": 2, "depart_ids": "ch_depart_04", "course_type": "2", "course_category": "2"},
		{"id": "ch_course_03", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "机器人搭建与竞赛",
			"course_desc": `<h2>机器人搭建与竞赛课程</h2><p>适合 8-14 岁学员。学习机器人机械结构搭建、传感器应用与编程控制，备战各类机器人竞赛。</p><h3>课程亮点</h3><ul><li>硬件+软件结合，动手与动脑并行</li><li>竞赛导向训练，参与省市机器人竞赛</li><li>团队协作与项目管理能力培养</li></ul><h3>适合年龄</h3><p>8-14 岁（小学 2 年级 - 初中）</p>`,
			"show_type":   1, "course_map": mapThemeOf("ch_course_03").fileID, "is_shared": 0, "show_home": 1, "order_num": 3, "depart_ids": "ch_depart_05", "course_type": "2", "course_category": "1"},
		{"id": "ch_course_04", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "Scratch 游戏设计专项",
			"course_desc": `<h2>Scratch 游戏设计专项课程</h2><p>在掌握 Scratch 基础操作后，深入学习游戏设计原理：碰撞检测、计分系统、多关卡设计、音效与动画。学员将独立完成 3 个完整游戏作品。</p><h3>课程亮点</h3><ul><li>游戏化学习，激发兴趣</li><li>涵盖游戏设计核心概念</li><li>作品可在 Scratch 社区分享</li></ul><h3>适合年龄</h3><p>7-10 岁（已完成 Scratch 启蒙）</p>`,
			"show_type":   1, "course_map": mapThemeOf("ch_course_04").fileID, "is_shared": 0, "show_home": 1, "order_num": 4, "depart_ids": "ch_depart_03", "course_type": "1", "course_category": "1"},
		{"id": "ch_course_05", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "Python 数据可视化入门",
			"course_desc": `<h2>Python 数据可视化入门课程</h2><p>学习使用 Matplotlib 和 Pandas 进行数据处理与可视化，制作漂亮的图表和报告。</p><h3>课程亮点</h3><ul><li>真实数据集分析（天气、运动、社交媒体）</li><li>培养数据思维与审美</li><li>为后续人工智能学习打基础</li></ul><h3>适合年龄</h3><p>11-14 岁（已掌握 Python 基础）</p>`,
			"show_type":   1, "course_map": mapThemeOf("ch_course_05").fileID, "is_shared": 0, "show_home": 1, "order_num": 5, "depart_ids": "ch_depart_04", "course_type": "2", "course_category": "2"},
		{"id": "ch_course_06", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "信息学奥赛 CSP-J 入门",
			"course_desc": `<h2>信息学奥赛 CSP-J 入门课程</h2><p>面向准备参加 CSP-J（入门级）认证的学员，系统学习 C++ 语法、基础算法与数据结构。</p><h3>课程亮点</h3><ul><li>竞赛导向，系统训练</li><li>涵盖历年真题解析</li><li>一对一作业批改与竞赛策略指导</li></ul><h3>适合年龄</h3><p>12-14 岁（已具备 Python 基础，准备学习 C++）</p>`,
			"show_type":   1, "course_map": mapThemeOf("ch_course_06").fileID, "is_shared": 0, "show_home": 0, "order_num": 6, "depart_ids": "ch_depart_04", "course_type": "2", "course_category": "2"},
	}
	for _, c := range courses {
		g.DB().Model("teaching_course").Data(c).Save()
	}
	fmt.Println("课程 6 门已插入")
}

// seedUnits 课程单元（22 个，覆盖全部 6 门课程）：闯关式命名（第N关）+
// 冒险路径坐标（map_x/map_y 对应 course_map 背景上的关卡节点），
// 每门课最后一关为作品提交 Boss 关（course_work_type=1）。
func seedUnits() {
	units := []map[string]interface{}{
		// Scratch 创意编程启蒙（关卡路径：从初识王国出发，蛇形前往 Boss 关）
		{"id": "ch_unit_01", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1关 初识Scratch", "unit_intro": "了解Scratch界面布局，创建第一个动画项目：让小猫说“你好！\"",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"media_content": "<p>关卡目标：熟悉Scratch编辑器界面，理解角色、舞台、脚本三个核心概念。</p>", "map_x": 100, "map_y": 150, "order_num": 1},
		{"id": "ch_unit_02", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2关 角色与运动", "unit_intro": "学习运动积木块，控制角色移动、旋转和弹性",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 310, "map_y": 130, "order_num": 2},
		{"id": "ch_unit_03", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3关 外观与声音", "unit_intro": "给角色添加造型切换和音效，制作会说话的角色",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 520, "map_y": 160, "order_num": 3},
		{"id": "ch_unit_04", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4关 事件与控制", "unit_intro": "学习事件触发和条件判断，制作互动故事",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 730, "map_y": 230, "order_num": 4},
		{"id": "ch_unit_05", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第5关 循环与变量【Boss】", "unit_intro": "关卡挑战：运用重复执行和变量制作计分器作品，提交作品即通关！",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 560, "map_y": 390, "order_num": 5},
		// Python 代码编程进阶
		{"id": "ch_unit_10", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1关 Hello Python", "unit_intro": "安装Python环境，编写第一个print程序",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 800, "map_y": 150, "order_num": 1},
		{"id": "ch_unit_11", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2关 变量与类型", "unit_intro": "学习变量、整数、浮点数、字符串和布尔值",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 590, "map_y": 130, "order_num": 2},
		{"id": "ch_unit_12", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3关 条件判断", "unit_intro": "学习if-else语句，编写猜数字小游戏",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 380, "map_y": 160, "order_num": 3},
		{"id": "ch_unit_13", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4关 循环结构", "unit_intro": "学习for和while循环，绘制Turtle图形",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 170, "map_y": 230, "order_num": 4},
		{"id": "ch_unit_14", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第5关 列表与字典【Boss】", "unit_intro": "关卡挑战：用列表与字典完成通讯录程序，提交作品即通关！",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 340, "map_y": 390, "order_num": 5},
		// 机器人搭建与竞赛
		{"id": "ch_unit_20", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1关 认识机器人", "unit_intro": "了解机器人组成结构，认识主控板和传感器",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 110, "map_y": 160, "order_num": 1},
		{"id": "ch_unit_21", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2关 LED与蜂鸣器", "unit_intro": "学习点亮LED和驱动蜂鸣器发声",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 340, "map_y": 130, "order_num": 2},
		{"id": "ch_unit_22", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3关 电机与轮子", "unit_intro": "学习驱动电机，让机器人动起来",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 570, "map_y": 170, "order_num": 3},
		{"id": "ch_unit_23", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4关 循线行走【Boss】", "unit_intro": "关卡挑战：调试红外传感器完成循线赛道，提交走通全程的作品视频即通关！",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 790, "map_y": 300, "order_num": 4},
		// Scratch 游戏设计专项
		{"id": "ch_unit_30", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1关 接苹果", "unit_intro": "学习碰撞检测，制作经典接苹果游戏",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 110, "map_y": 150, "order_num": 1},
		{"id": "ch_unit_31", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2关 打地鼠", "unit_intro": "学习随机出现和计分系统，制作打地鼠",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 340, "map_y": 130, "order_num": 2},
		{"id": "ch_unit_32", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3关 迷宫探险", "unit_intro": "学习方向键控制和墙壁碰撞，制作迷宫游戏",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 570, "map_y": 160, "order_num": 3},
		{"id": "ch_unit_33", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4关 多关卡设计【Boss】", "unit_intro": "关卡挑战：设计一款含 3 个关卡的完整游戏并提交，即通关毕业！",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 790, "map_y": 280, "order_num": 4},
		// Python 数据可视化入门
		{"id": "ch_unit_40", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1关 初识Matplotlib", "unit_intro": "安装Matplotlib，绘制第一条折线图",
			"course_id": "ch_course_05", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 160, "map_y": 160, "order_num": 1},
		{"id": "ch_unit_41", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2关 柱状图与饼图【Boss】", "unit_intro": "关卡挑战：用真实数据集绘制一张柱状图与一张饼图，提交即通关！",
			"course_id": "ch_course_05", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 460, "map_y": 130, "order_num": 2},
		// 信息学奥赛 CSP-J 入门
		{"id": "ch_unit_60", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1关 C++ 起步", "unit_intro": "搭建 C++ 编译环境，理解程序编译运行流程",
			"course_id": "ch_course_06", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 740, "map_y": 160, "order_num": 1},
		{"id": "ch_unit_61", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2关 顺序与分支【Boss】", "unit_intro": "关卡挑战：完成一套 CSP-J 入门真题的顺序与分支题目，提交即通关！",
			"course_id": "ch_course_06", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 440, "map_y": 130, "order_num": 2},
	}
	for _, u := range units {
		g.DB().Model("teaching_course_unit").Data(u).Save()
	}
	fmt.Println("课程单元 22 个已插入")
}

// seedCourseDepts 课程-班级授权。
func seedCourseDepts() {
	cds := []map[string]interface{}{
		{"id": "ch_cd_01", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "dept_id": "ch_depart_03", "course_id": "ch_course_01", "open_time": "2026-09-10 08:00:00"},
		{"id": "ch_cd_02", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "dept_id": "ch_depart_04", "course_id": "ch_course_02", "open_time": "2026-09-10 08:00:00"},
		{"id": "ch_cd_03", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "dept_id": "ch_depart_05", "course_id": "ch_course_03", "open_time": "2026-09-10 08:00:00"},
		{"id": "ch_cd_04", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "dept_id": "ch_depart_03", "course_id": "ch_course_04", "open_time": "2026-10-01 08:00:00"},
		{"id": "ch_cd_05", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "dept_id": "ch_depart_04", "course_id": "ch_course_05", "open_time": "2026-10-01 08:00:00"},
		{"id": "ch_cd_06", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "dept_id": "ch_depart_04", "course_id": "ch_course_06", "open_time": "2026-11-01 08:00:00"},
	}
	for _, cd := range cds {
		g.DB().Model("teaching_course_dept").Data(cd).Save()
	}
	fmt.Println("课程授权 6 条已插入")
}

// seedAnnouncements 公告（4 条，含完整经营场所）。
func seedAnnouncements() {
	anns := []map[string]interface{}{
		{"id": "ch_ann_01", "titile": brandName + " 2026 秋季学期开学通知",
			"msg_content": "<p>各位家长、同学：</p><p>" + brandName + " 2026 年秋季学期将于 9 月 10 日正式开课，请按时到校。</p><p>上课地址：武胜县沿口镇上东街257号（邮编638400）</p><p>—— " + ownerName + "</p>",
			"start_time":  "2026-09-01 00:00:00", "end_time": "2027-01-31 23:59:59", "sender": "admin", "priority": "H",
			"msg_category": "1", "msg_type": "ALL", "send_status": "1", "send_time": "2026-09-01 08:00:00",
			"del_flag": "0", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "msg_abstract": "秋季学期开学通知"},
		{"id": "ch_ann_02", "titile": "Scratch 创意编程启蒙班课程安排",
			"msg_content": "<p>Scratch 启蒙班每周六上午 9:00-11:00 上课，请提前 10 分钟到教室（2楼 Scratch教室）。</p>",
			"start_time":  "2026-09-01 00:00:00", "end_time": "2027-01-31 23:59:59", "sender": "teacher1", "priority": "M",
			"msg_category": "1", "msg_type": "ALL", "send_status": "1", "send_time": "2026-09-05 08:00:00",
			"del_flag": "0", "create_by": "teacher1", "create_time": "2026-09-07 00:00:00", "msg_abstract": "Scratch班课程安排"},
		{"id": "ch_ann_03", "titile": "Python 进阶班新学期课程大纲发布",
			"msg_content": "<p>Python 进阶班新学期将学习条件判断、循环结构、列表与字典等内容，请同学们提前预习第一课内容。</p>",
			"start_time":  "2026-09-01 00:00:00", "end_time": "2027-01-31 23:59:59", "sender": "teacher2", "priority": "M",
			"msg_category": "1", "msg_type": "ALL", "send_status": "1", "send_time": "2026-09-05 08:00:00",
			"del_flag": "0", "create_by": "teacher2", "create_time": "2026-09-07 00:00:00", "msg_abstract": "Python班课程大纲"},
		{"id": "ch_ann_04", "titile": "机器人竞赛班即将开启报名",
			"msg_content": "<p>机器人竞赛班将于 10 月开始报名，名额有限（8人/班），请有意向的同学尽快联系陈教练。</p>",
			"start_time":  "2026-09-01 00:00:00", "end_time": "2026-10-31 23:59:59", "sender": "teacher3", "priority": "H",
			"msg_category": "1", "msg_type": "ALL", "send_status": "1", "send_time": "2026-09-05 08:00:00",
			"del_flag": "0", "create_by": "teacher3", "create_time": "2026-09-07 00:00:00", "msg_abstract": "机器人竞赛班报名"},
	}
	for _, a := range anns {
		g.DB().Model("sys_announcement").Data(a).Save()
	}
	fmt.Println("公告 4 条已插入")
}

// seedAnnouncementSends 公告阅读标记。
func seedAnnouncementSends() {
	as := []map[string]interface{}{
		{"id": "ch_as_01", "annt_id": "ch_ann_01", "user_id": "ch_user_teacher1", "read_flag": "0"},
		{"id": "ch_as_02", "annt_id": "ch_ann_01", "user_id": "ch_user_teacher2", "read_flag": "0"},
		{"id": "ch_as_03", "annt_id": "ch_ann_01", "user_id": "ch_user_teacher3", "read_flag": "0"},
		{"id": "ch_as_04", "annt_id": "ch_ann_01", "user_id": "ch_user_stu1", "read_flag": "0"},
		{"id": "ch_as_05", "annt_id": "ch_ann_01", "user_id": "ch_user_stu2", "read_flag": "0"},
		{"id": "ch_as_06", "annt_id": "ch_ann_01", "user_id": "ch_user_stu3", "read_flag": "0"},
		{"id": "ch_as_07", "annt_id": "ch_ann_01", "user_id": "ch_user_stu4", "read_flag": "0"},
		{"id": "ch_as_08", "annt_id": "ch_ann_01", "user_id": "ch_user_stu5", "read_flag": "0"},
		{"id": "ch_as_09", "annt_id": "ch_ann_01", "user_id": "ch_user_stu6", "read_flag": "0"},
		{"id": "ch_as_10", "annt_id": "ch_ann_02", "user_id": "ch_user_stu1", "read_flag": "0"},
		{"id": "ch_as_11", "annt_id": "ch_ann_02", "user_id": "ch_user_stu2", "read_flag": "0"},
		{"id": "ch_as_12", "annt_id": "ch_ann_04", "user_id": "ch_user_stu5", "read_flag": "0"},
		{"id": "ch_as_13", "annt_id": "ch_ann_04", "user_id": "ch_user_stu6", "read_flag": "0"},
	}
	for _, a := range as {
		g.DB().Model("sys_announcement_send").Data(a).Save()
	}
	fmt.Println("阅读标记 13 条已插入")
}

// seedNews 资讯（teaching_news，news_status>=1 才在门户 newsList 可见）。
func seedNews() {
	news := []map[string]interface{}{
		{"id": "ch_news_01", "news_title": "春晖青少年编程中心 2026 秋季学期顺利开课",
			"news_content": "<p>9 月 10 日，" + brandName + " 2026 年秋季学期正式开课。本学期开设 Scratch 创意编程启蒙、Python 代码编程进阶、机器人搭建与竞赛等 6 门课程，共 6 个教学班，学员 100 余人次。</p><p>经营场所：武胜县沿口镇上东街257号（邮编638400）。</p><p>—— " + ownerName + "</p>",
			"news_status":  1, "create_by": "admin", "create_time": "2026-09-10 09:00:00"},
		{"id": "ch_news_02", "news_title": "我中心学员在广安市青少年编程展演活动中获得佳绩",
			"news_content": "<p>在近日举办的广安市青少年编程展演活动中，我中心 Scratch 启蒙班与 Python 进阶班共 6 名学员晋级展示环节，其中 2 名学员获得一等奖。</p><p>春晖编程将继续以项目式学习为核心，让每个孩子拥有创造未来的能力。</p>",
			"news_status":  1, "create_by": "teacher1", "create_time": "2026-10-18 14:00:00"},
		{"id": "ch_news_03", "news_title": "Scratch 创意编程公益公开课圆满结束",
			"news_content": "<p>10 月下旬，中心举办公益公开课两场，共 30 组家庭参与。公开课围绕“图形化编程启蒙”主题，孩子们完成了自己的第一个动画作品。</p>",
			"news_status":  1, "create_by": "teacher1", "create_time": "2026-10-30 16:00:00"},
		{"id": "ch_news_04", "news_title": "寒假 Python 专项班与 CSP-J 冲刺班报名开始",
			"news_content": "<p>寒假期间将开设 Python 数据可视化专项班与信息学奥赛 CSP-J 冲刺班，每班限额 8 人，报满即止。请联系各班讲师或到中心前台咨询。</p><p>地址：武胜县沿口镇上东街257号（邮编638400）。</p>",
			"news_status":  1, "create_by": "teacher2", "create_time": "2026-11-15 10:00:00"},
	}
	for _, n := range news {
		g.DB().Model("teaching_news").Data(n).Save()
	}
	fmt.Println("资讯 4 条已插入")
}

// workSeed 一条作品种子定义。
type workSeed struct {
	id     string
	stu    string // sys_user id
	course string
	unit   string
	name   string
	status int // 3=已提交 4=已分享（leaderboard 过滤 work_status>=3）
	star   int
	view   int
}

var workSeeds = []workSeed{
	{"ch_work_01", "ch_user_stu1", "ch_course_01", "ch_unit_01", "我的第一个动画：小猫说你好", 4, 12, 58},
	{"ch_work_02", "ch_user_stu2", "ch_course_01", "ch_unit_02", "小猫追球", 4, 9, 41},
	{"ch_work_03", "ch_user_stu1", "ch_course_01", "ch_unit_03", "会说话的角色", 3, 3, 17},
	{"ch_work_04", "ch_user_stu3", "ch_course_02", "ch_unit_10", "Hello Python", 4, 11, 66},
	{"ch_work_05", "ch_user_stu4", "ch_course_02", "ch_unit_12", "猜数字游戏", 4, 15, 73},
	{"ch_work_06", "ch_user_stu3", "ch_course_02", "ch_unit_13", "Turtle 画星星", 3, 5, 22},
	{"ch_work_07", "ch_user_stu5", "ch_course_03", "ch_unit_22", "循线机器人", 4, 8, 39},
	{"ch_work_08", "ch_user_stu6", "ch_course_03", "ch_unit_21", "LED 流水灯", 4, 6, 28},
	{"ch_work_09", "ch_user_stu5", "ch_course_03", "ch_unit_23", "避障小车", 3, 2, 13},
	{"ch_work_10", "ch_user_stu1", "ch_course_04", "ch_unit_30", "接苹果", 4, 13, 61},
	{"ch_work_11", "ch_user_stu2", "ch_course_04", "ch_unit_31", "打地鼠", 4, 10, 47},
	{"ch_work_12", "ch_user_stu4", "ch_course_04", "ch_unit_32", "迷宫探险", 4, 4, 25},
}

// coverDef 封面定义：sys_file id + 渐变起止色。
type coverDef struct {
	id     string
	c1, c2 [3]uint8
}

// coverDefs 六门课程的渐变封面（配色与门户 _homeHtml 卡片一致）。
var coverDefs = []coverDef{
	{"ch_file_cv01", [3]uint8{102, 126, 234}, [3]uint8{118, 75, 162}},
	{"ch_file_cv02", [3]uint8{240, 147, 251}, [3]uint8{245, 87, 108}},
	{"ch_file_cv03", [3]uint8{79, 172, 254}, [3]uint8{0, 242, 254}},
	{"ch_file_cv04", [3]uint8{250, 200, 80}, [3]uint8{240, 120, 60}},
	{"ch_file_cv05", [3]uint8{67, 233, 145}, [3]uint8{35, 166, 138}},
	{"ch_file_cv06", [3]uint8{130, 140, 160}, [3]uint8{60, 70, 90}},
}

// seedFiles 生成作品封面（真实 PNG 渐变图）与 .sb3 作品占位文件，
// 写入本地存储根目录并登记 sys_file（静态预览走 /api/sys/common/static/*path）。
func seedFiles() {
	root := os.Getenv("STORAGE_LOCAL_ROOT")
	if root == "" {
		root = "/tmp/chunhui_upload"
	}
	n := 0
	for _, c := range coverDefs {
		path := fmt.Sprintf("chunhui/covers/%s.png", c.id)
		if err := writeGradientPNG(filepath.Join(root, filepath.FromSlash(path)), 480, 300, c.c1, c.c2); err != nil {
			fmt.Printf("生成封面 %s 失败: %v\n", path, err)
			continue
		}
		saveFile(c.id, "cover.png", path, 0)
		n++
	}
	w := 0
	for _, wk := range workSeeds {
		path := fmt.Sprintf("chunhui/works/%s.sb3", wk.id)
		body := fmt.Sprintf(`{"project":"%s","targets":[],"meta":{"platform":"Chunhui Scratch"}}`, wk.name)
		abs := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
			continue
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			fmt.Printf("写入作品文件 %s 失败: %v\n", path, err)
			continue
		}
		w++
		saveFile(wk.id+"_file", wk.name+".sb3", path, 1)
	}
	fmt.Printf("作品文件 %d 个 + 封面 %d 张已生成并登记 sys_file\n", w, n)
}

// saveFile 登记一条 sys_file（file_type: 0=图片 1=作品；file_location 1=本地）。
func saveFile(id, name, path string, fileType int) {
	g.DB().Model("sys_file").Data(map[string]interface{}{
		"id": id, "file_name": name, "file_path": path, "file_type": fileType,
		"file_location": 1, "file_tag": "chunhui", "sys_org_code": "A01",
		"create_by": "admin", "create_time": "2026-09-07 00:00:00", "del_flag": 0,
	}).Save()
}

// writeGradientPNG 绘制 480x300 垂直双渐变色封面 PNG。
func writeGradientPNG(abs string, w, h int, top, bottom [3]uint8) error {
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		// 全程 int 域插值后再转 uint8：uint8 域直接相减会下溢（如 75-126）。
		t := float64(y) / float64(h-1)
		r := uint8(int(float64(top[0]) + (float64(bottom[0])-float64(top[0]))*t))
		gg := uint8(int(float64(top[1]) + (float64(bottom[1])-float64(top[1]))*t))
		b := uint8(int(float64(top[2]) + (float64(bottom[2])-float64(top[2]))*t))
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{r, gg, b, 255})
		}
	}
	f, err := os.Create(abs)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

// seedWorks 学员作品（teaching_work，work_status>=3 门户 leaderboard 可见）。
func seedWorks() {
	n := 0
	for _, wk := range workSeeds {
		// work_scene 列为 varchar(20)，仅存极简 JSON 标记。
		_, err := g.DB().Model("teaching_work").Data(map[string]interface{}{
			"id": wk.id, "create_by": wk.stu, "create_time": "2026-09-20 10:00:00",
			"sys_org_code": "A01", "user_id": wk.stu,
			"depart_id":   mapCourseDepart(wk.course),
			"course_id":   wk.course,
			"work_name":   wk.name,
			"work_type":   1,
			"work_file":   wk.id + "_file",
			"work_cover":  coverOf(wk.course),
			"work_status": wk.status,
			"star_num":    wk.star, "collect_num": wk.star / 3, "view_num": wk.view,
			"work_scene": `{"v":1}`, "has_cloud_data": 0,
			"del_flag": 0,
		}).Save()
		if err != nil {
			fmt.Printf("插入作品 %s 失败: %v\n", wk.id, err)
			continue
		}
		n++
	}
	fmt.Printf("作品 %d 件已插入\n", n)
}

// mapCourseDepart 课程 → 授权班级。
func mapCourseDepart(course string) string {
	switch course {
	case "ch_course_01", "ch_course_04":
		return "ch_depart_03"
	case "ch_course_02", "ch_course_05", "ch_course_06":
		return "ch_depart_04"
	default:
		return "ch_depart_05"
	}
}

// coverOf 课程 → 封面 sys_file id。
func coverOf(course string) string {
	m := map[string]string{
		"ch_course_01": "ch_file_cv01", "ch_course_02": "ch_file_cv02",
		"ch_course_03": "ch_file_cv03", "ch_course_04": "ch_file_cv04",
		"ch_course_05": "ch_file_cv05", "ch_course_06": "ch_file_cv06",
	}
	return m[course]
}

// seedWorkComments 作品互评。
func seedWorkComments() {
	cmts := []map[string]interface{}{
		{"id": "ch_wc_01", "work_id": "ch_work_01", "user_id": "ch_user_stu2", "comment": "小猫的造型切换好流畅，我也要试试！", "create_by": "ch_user_stu2", "create_time": "2026-09-21 09:00:00"},
		{"id": "ch_wc_02", "work_id": "ch_work_01", "user_id": "ch_user_teacher1", "comment": "第一次就能完成完整动画，继续保持！", "create_by": "ch_user_teacher1", "create_time": "2026-09-21 10:00:00"},
		{"id": "ch_wc_03", "work_id": "ch_work_05", "user_id": "ch_user_stu3", "comment": "猜数字的循环写得真棒，用了二分吗？", "create_by": "ch_user_stu3", "create_time": "2026-09-25 15:00:00"},
		{"id": "ch_wc_04", "work_id": "ch_work_07", "user_id": "ch_user_stu6", "comment": "循线机器人转弯很稳，速度也快！", "create_by": "ch_user_stu6", "create_time": "2026-09-28 11:00:00"},
		{"id": "ch_wc_05", "work_id": "ch_work_10", "user_id": "ch_user_stu5", "comment": "苹果接住时有音效，细节满分。", "create_by": "ch_user_stu5", "create_time": "2026-10-09 16:00:00"},
		{"id": "ch_wc_06", "work_id": "ch_work_04", "user_id": "ch_user_teacher2", "comment": "代码缩进规范，注释也很清楚。", "create_by": "ch_user_teacher2", "create_time": "2026-10-11 09:30:00"},
	}
	for _, c := range cmts {
		g.DB().Model("teaching_work_comment").Data(c).Save()
	}
	fmt.Println("作品评论 6 条已插入")
}

// mapNode 关卡节点坐标（900x1000 地图画布上的像素位）。
type mapNode struct{ x, y int }

// mapTheme 课程闯关地图主题：背景图配色 + 关卡路径。
// 节点顺序即通关顺序（与单元 order_num 对应）。
type mapTheme struct {
	courseID string
	fileID   string
	road     [3]uint8 // 路径色（与课程封面渐变同族）
	skyTop   [3]uint8
	skyBot   [3]uint8
	nodes    []mapNode
}

// mapThemes 六门课程的闯关路径（蛇形冒险路线，越往后越高/越靠近 Boss 关）。
var mapThemes = []mapTheme{
	{"ch_course_01", "ch_file_map01", [3]uint8{118, 75, 162}, [3]uint8{228, 238, 255}, [3]uint8{185, 190, 240},
		[]mapNode{{100, 150}, {310, 130}, {520, 160}, {730, 230}, {560, 390}}},
	{"ch_course_02", "ch_file_map02", [3]uint8{245, 87, 108}, [3]uint8{255, 236, 244}, [3]uint8{250, 205, 220},
		[]mapNode{{800, 150}, {590, 130}, {380, 160}, {170, 230}, {340, 390}}},
	{"ch_course_03", "ch_file_map03", [3]uint8{0, 150, 200}, [3]uint8{224, 246, 255}, [3]uint8{180, 225, 245},
		[]mapNode{{110, 160}, {340, 130}, {570, 170}, {790, 300}}},
	{"ch_course_04", "ch_file_map04", [3]uint8{240, 120, 60}, [3]uint8{255, 244, 224}, [3]uint8{250, 216, 180},
		[]mapNode{{110, 150}, {340, 130}, {570, 160}, {790, 280}}},
	{"ch_course_05", "ch_file_map05", [3]uint8{35, 166, 138}, [3]uint8{224, 250, 240}, [3]uint8{180, 235, 215},
		[]mapNode{{160, 160}, {460, 130}}},
	{"ch_course_06", "ch_file_map06", [3]uint8{90, 100, 130}, [3]uint8{235, 238, 248}, [3]uint8{195, 200, 220},
		[]mapNode{{740, 160}, {440, 130}}},
}

// mapThemeOf 课程 id → 地图主题。
func mapThemeOf(courseID string) mapTheme {
	for i := range mapThemes {
		if mapThemes[i].courseID == courseID {
			return mapThemes[i]
		}
	}
	return mapTheme{}
}

// seedMaps 生成 6 张闯关式课程地图背景（天空渐变 + 云朵 + 蛇形冒险路径 +
// 关卡节点圆台 + 出发/终局标记），登记 sys_file 并供 course_map 引用。
func seedMaps() {
	root := os.Getenv("STORAGE_LOCAL_ROOT")
	if root == "" {
		root = "/tmp/chunhui_upload"
	}
	n := 0
	for i := range mapThemes {
		t := &mapThemes[i]
		path := fmt.Sprintf("chunhui/maps/%s.png", t.courseID)
		if err := drawAdventureMap(filepath.Join(root, filepath.FromSlash(path)), t); err != nil {
			fmt.Printf("生成地图 %s 失败: %v\n", path, err)
			continue
		}
		saveFile(t.fileID, "course-map.png", path, 0)
		n++
	}
	fmt.Printf("闯关地图背景 %d 张已生成并登记 sys_file\n", n)
}

// drawAdventureMap 绘制单张闯关地图（900x1000）。
func drawAdventureMap(abs string, t *mapTheme) error {
	const w, h = 900, 1000
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// 天空渐变（int 域插值防 uint8 下溢）
	for y := 0; y < h; y++ {
		tt := float64(y) / float64(h-1)
		r := uint8(int(float64(t.skyTop[0]) + (float64(t.skyBot[0])-float64(t.skyTop[0]))*tt))
		gg := uint8(int(float64(t.skyTop[1]) + (float64(t.skyBot[1])-float64(t.skyTop[1]))*tt))
		b := uint8(int(float64(t.skyTop[2]) + (float64(t.skyBot[2])-float64(t.skyTop[2]))*tt))
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{r, gg, b, 255})
		}
	}
	// 装饰云朵（不透明浅白，圆簇）
	cloud := color.RGBA{252, 253, 255, 255}
	for _, c := range [][2]int{{160, 90}, {640, 70}, {430, 210}} {
		fillCircle(img, c[0], c[1], 34, cloud)
		fillCircle(img, c[0]-36, c[1]+8, 24, cloud)
		fillCircle(img, c[0]+38, c[1]+10, 26, cloud)
	}
	// 蛇形路径：相邻节点连线，圆头粗路 + 浅色中线
	roadC := color.RGBA{t.road[0], t.road[1], t.road[2], 255}
	midC := color.RGBA{uint8(t.road[0] + 60), uint8(t.road[1] + 60), uint8(t.road[2] + 60), 255}
	if midC.R < t.road[0] {
		midC = roadC
	}
	for i := 0; i < len(t.nodes)-1; i++ {
		drawThickLine(img, t.nodes[i], t.nodes[i+1], 13, roadC)
		drawDashedLine(img, t.nodes[i], t.nodes[i+1], 3, midC)
	}
	// 关卡节点圆台：起点绿环、终点金色 Boss 环、中间浅色平台
	for i, nd := range t.nodes {
		var ring [3]uint8
		switch i {
		case 0:
			ring = [3]uint8{70, 180, 90}
		case len(t.nodes) - 1:
			ring = [3]uint8{230, 170, 40}
		default:
			ring = [3]uint8{t.road[0] + 30, t.road[1] + 30, t.road[2] + 30}
		}
		fillCircle(img, nd.x, nd.y, 36, color.RGBA{ring[0], ring[1], ring[2], 255})
		fillCircle(img, nd.x, nd.y, 30, color.RGBA{252, 252, 252, 255})
	}
	f, err := os.Create(abs)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}

// fillCircle 圆填充（欧氏距离判定）。
func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			if x < img.Bounds().Min.X || x >= img.Bounds().Max.X ||
				y < img.Bounds().Min.Y || y >= img.Bounds().Max.Y {
				continue
			}
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, c)
			}
		}
	}
}

// drawThickLine 圆头粗线：沿段步进盖圆。
func drawThickLine(img *image.RGBA, a, b mapNode, r int, c color.RGBA) {
	steps := max(1, abs(b.x-a.x)+abs(b.y-a.y))
	for s := 0; s <= steps; s++ {
		tt := float64(s) / float64(steps)
		x := a.x + int(float64(b.x-a.x)*tt)
		y := a.y + int(float64(b.y-a.y)*tt)
		fillCircle(img, x, y, r, c)
	}
}

// drawDashedLine 虚线中线：每隔一段画一个短点。
func drawDashedLine(img *image.RGBA, a, b mapNode, r int, c color.RGBA) {
	steps := max(1, abs(b.x-a.x)+abs(b.y-a.y))
	for s := 0; s <= steps; s += 22 {
		tt := float64(s) / float64(steps)
		x := a.x + int(float64(b.x-a.x)*tt)
		y := a.y + int(float64(b.y-a.y)*tt)
		fillCircle(img, x, y, r, c)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// seedAdminBase 管理后台基础数据（position/数据源/消息模板/部门角色）。
// 这些表在 V2.8 演示库中有数据，全量替换 Mock 后被清空 → 管理页显示「暂无数据」。
func seedAdminBase() {
	positions := []map[string]interface{}{
		{"id": "ch_pos_01", "code": "principal", "name": "校长", "post_rank": "1", "company_id": "ch_depart_01", "sys_org_code": "A01", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_pos_02", "code": "teacher", "name": "讲师", "post_rank": "2", "company_id": "ch_depart_01", "sys_org_code": "A01", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_pos_03", "code": "coach", "name": "竞赛教练", "post_rank": "2", "company_id": "ch_depart_01", "sys_org_code": "A01", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_pos_04", "code": "assistant", "name": "助教", "post_rank": "3", "company_id": "ch_depart_01", "sys_org_code": "A01", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_pos_05", "code": "affairs", "name": "教务", "post_rank": "3", "company_id": "ch_depart_02", "sys_org_code": "A01", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
	}
	for _, p := range positions {
		g.DB().Model("sys_position").Data(p).Save()
	}

	// 主数据源（指向本库，供「多数据源管理」页展示与在线表单引用）
	g.DB().Model("sys_data_source").Data(map[string]interface{}{
		"id": "ch_ds_master", "code": "master", "name": "春晖教学平台主库",
		"remark":  "本系统默认数据源（teachingopen）",
		"db_type": "mysql", "db_driver": "com.mysql.cj.jdbc.Driver",
		"db_url":  "jdbc:mysql://127.0.0.1:13306/teachingopen?useUnicode=true&characterEncoding=utf8&serverTimezone=Asia/Shanghai",
		"db_name": "teachingopen", "db_username": "teaching", "db_password": "",
		"create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01",
	}).Save()

	tpls := []map[string]interface{}{
		{"id": "ch_tpl_01", "template_name": "账号注册验证码", "template_code": "SMS_VERIFY_CODE",
			"template_type": "1", "template_content": "您的验证码为 ${code}，10 分钟内有效。",
			"template_test_json": `{"code":"123456"}`, "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_tpl_02", "template_name": "作业布置通知", "template_code": "WORK_ASSIGN_NOTICE",
			"template_type": "2", "template_content": "【春晖编程】老师布置了新作业：${workName}，请在 ${deadline} 前提交。",
			"template_test_json": `{"workName":"第2关 角色与运动","deadline":"本周六"}`, "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_tpl_03", "template_name": "作品被点赞通知", "template_code": "WORK_LIKE_NOTICE",
			"template_type": "2", "template_content": "【春晖编程】你的作品《${workName}》获得了新的点赞，当前 ${starNum} 赞！",
			"template_test_json": `{"workName":"接苹果","starNum":"13"}`, "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
	}
	for _, t := range tpls {
		g.DB().Model("sys_sms_template").Data(t).Save()
	}

	// 部门角色（班级管理员角色，供通讯录/部门角色页展示）
	roles := []map[string]interface{}{
		{"id": "ch_dr_01", "depart_id": "ch_depart_03", "role_name": "Scratch班管理员", "role_code": "scratch_admin",
			"description": "Scratch启蒙班班级管理", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_dr_02", "depart_id": "ch_depart_04", "role_name": "Python班管理员", "role_code": "python_admin",
			"description": "Python进阶班班级管理", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
		{"id": "ch_dr_03", "depart_id": "ch_depart_05", "role_name": "机器人班管理员", "role_code": "robot_admin",
			"description": "机器人竞赛班班级管理", "create_by": "admin", "create_time": "2026-09-07 00:00:00"},
	}
	for _, r := range roles {
		g.DB().Model("sys_depart_role").Data(r).Save()
	}
	fmt.Println("管理基础数据：职务 5 / 数据源 1 / 消息模板 3 / 部门角色 3 已插入")
}
