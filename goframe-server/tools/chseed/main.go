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
	"fmt"
	"os"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/genv"
)

// 品牌常量（与 api/db/chunhui_seed.sql 保持一致）。
const (
	brandName   = "春晖青少年编程中心"
	ownerName   = "武胜县沿口春夏秋信息技术中心（个体工商户）"
	creditCode  = "92511622MAKLLHTT5D"
	fullAddress = "四川省广安市武胜县沿口镇上东街257号（邮编638400）"
	copyRight   = "© 2026 武胜县沿口春夏秋信息技术中心（个体工商户）· 统一社会信用代码 92511622MAKLLHTT5D"
)

func main() {
	genv.Set("MYSQL_LINK", "mysql:teaching:Teaching2026!@tcp(127.0.0.1:3306)/teachingopen?parseTime=true&loc=Local")
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{Link: genv.Get("MYSQL_LINK").String(), Type: "mysql"},
		},
	})

	if err := g.DB().PingMaster(); err != nil {
		fmt.Println("数据库连接失败:", err)
		os.Exit(1)
	}
	fmt.Println("数据库连接正常")

	// 1. 清除旧数据
	for _, t := range []string{"sys_config", "sys_depart", "sys_user", "sys_user_role", "sys_user_depart",
		"teaching_course", "teaching_course_unit", "teaching_course_dept", "sys_announcement", "sys_announcement_send"} {
		if _, err := g.DB().Exec(context.Background(), "DELETE FROM "+t); err != nil {
			fmt.Printf("清除 %s 失败: %v\n", t, err)
		}
	}
	fmt.Println("旧数据已清除")

	seedDeparts()
	seedRoles()
	seedRolePermissions()
	seedUsers()
	seedUserRoles()
	seedUserDeparts()
	seedConfigs()
	seedCourses()
	seedUnits()
	seedCourseDepts()
	seedAnnouncements()
	seedAnnouncementSends()

	fmt.Println("✓ 春晖青少年编程中心种子数据导入完成")
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
			"show_type": 1, "course_map": 0, "is_shared": 0, "show_home": 1, "order_num": 1, "depart_ids": "ch_depart_03", "course_type": "1", "course_category": "1"},
		{"id": "ch_course_02", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "Python 代码编程进阶",
			"course_desc": `<h2>Python 代码编程进阶课程</h2><p>适合 10-14 岁有一定编程基础的学员。从图形化编程过渡到真正的代码编程，学习 Python 语法、数据结构、算法与项目开发。</p><h3>课程亮点</h3><ul><li>从积木到代码的平滑过渡</li><li>涵盖 Turtle绘图、Pygame游戏开发、数据分析入门</li><li>培养算法思维与工程能力</li><li>完成多个实战项目</li></ul><h3>适合年龄</h3><p>10-14 岁（小学 4 年级 - 初中）</p>`,
			"show_type": 1, "course_map": 0, "is_shared": 0, "show_home": 1, "order_num": 2, "depart_ids": "ch_depart_04", "course_type": "2", "course_category": "2"},
		{"id": "ch_course_03", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "机器人搭建与竞赛",
			"course_desc": `<h2>机器人搭建与竞赛课程</h2><p>适合 8-14 岁学员。学习机器人机械结构搭建、传感器应用与编程控制，备战各类机器人竞赛。</p><h3>课程亮点</h3><ul><li>硬件+软件结合，动手与动脑并行</li><li>竞赛导向训练，参与省市机器人竞赛</li><li>团队协作与项目管理能力培养</li></ul><h3>适合年龄</h3><p>8-14 岁（小学 2 年级 - 初中）</p>`,
			"show_type": 1, "course_map": 0, "is_shared": 0, "show_home": 1, "order_num": 3, "depart_ids": "ch_depart_05", "course_type": "2", "course_category": "1"},
		{"id": "ch_course_04", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "Scratch 游戏设计专项",
			"course_desc": `<h2>Scratch 游戏设计专项课程</h2><p>在掌握 Scratch 基础操作后，深入学习游戏设计原理：碰撞检测、计分系统、多关卡设计、音效与动画。学员将独立完成 3 个完整游戏作品。</p><h3>课程亮点</h3><ul><li>游戏化学习，激发兴趣</li><li>涵盖游戏设计核心概念</li><li>作品可在 Scratch 社区分享</li></ul><h3>适合年龄</h3><p>7-10 岁（已完成 Scratch 启蒙）</p>`,
			"show_type": 1, "course_map": 0, "is_shared": 0, "show_home": 1, "order_num": 4, "depart_ids": "ch_depart_03", "course_type": "1", "course_category": "1"},
		{"id": "ch_course_05", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "Python 数据可视化入门",
			"course_desc": `<h2>Python 数据可视化入门课程</h2><p>学习使用 Matplotlib 和 Pandas 进行数据处理与可视化，制作漂亮的图表和报告。</p><h3>课程亮点</h3><ul><li>真实数据集分析（天气、运动、社交媒体）</li><li>培养数据思维与审美</li><li>为后续人工智能学习打基础</li></ul><h3>适合年龄</h3><p>11-14 岁（已掌握 Python 基础）</p>`,
			"show_type": 1, "course_map": 0, "is_shared": 0, "show_home": 1, "order_num": 5, "depart_ids": "ch_depart_04", "course_type": "2", "course_category": "2"},
		{"id": "ch_course_06", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"course_name": "信息学奥赛 CSP-J 入门",
			"course_desc": `<h2>信息学奥赛 CSP-J 入门课程</h2><p>面向准备参加 CSP-J（入门级）认证的学员，系统学习 C++ 语法、基础算法与数据结构。</p><h3>课程亮点</h3><ul><li>竞赛导向，系统训练</li><li>涵盖历年真题解析</li><li>一对一作业批改与竞赛策略指导</li></ul><h3>适合年龄</h3><p>12-14 岁（已具备 Python 基础，准备学习 C++）</p>`,
			"show_type": 1, "course_map": 0, "is_shared": 0, "show_home": 0, "order_num": 6, "depart_ids": "ch_depart_04", "course_type": "2", "course_category": "2"},
	}
	for _, c := range courses {
		g.DB().Model("teaching_course").Data(c).Save()
	}
	fmt.Println("课程 6 门已插入")
}

// seedUnits 课程单元（22 个，覆盖全部 6 门课程）。
func seedUnits() {
	units := []map[string]interface{}{
		// Scratch 创意编程启蒙
		{"id": "ch_unit_01", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1课 认识Scratch", "unit_intro": "了解Scratch界面布局，学会创建第一个动画项目：让小猫说“你好！”",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"media_content": "<p>本节课学习目标：熟悉Scratch编辑器界面，理解角色、舞台、脚本三个核心概念。</p>", "map_x": 0, "map_y": 0, "order_num": 1},
		{"id": "ch_unit_02", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2课 角色与运动", "unit_intro": "学习运动积木块，控制角色移动、旋转和弹性",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 100, "map_y": 200, "order_num": 2},
		{"id": "ch_unit_03", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3课 外观与声音", "unit_intro": "给角色添加造型切换和音效，制作会说话的角色",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 3},
		{"id": "ch_unit_04", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4课 事件与控制", "unit_intro": "学习事件触发和条件判断，制作互动故事",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 200, "map_y": 200, "order_num": 4},
		{"id": "ch_unit_05", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第5课 循环与变量", "unit_intro": "使用重复执行和变量，制作计分器和倒计时",
			"course_id": "ch_course_01", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 5},
		// Python 代码编程进阶
		{"id": "ch_unit_10", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1课 Hello Python", "unit_intro": "安装Python环境，编写第一个print程序",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 1},
		{"id": "ch_unit_11", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2课 变量与数据类型", "unit_intro": "学习变量、整数、浮点数、字符串和布尔值",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 2},
		{"id": "ch_unit_12", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3课 条件判断", "unit_intro": "学习if-else语句，编写猜数字游戏",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 200, "map_y": 200, "order_num": 3},
		{"id": "ch_unit_13", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4课 循环结构", "unit_intro": "学习for和while循环，绘制Turtle图形",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 4},
		{"id": "ch_unit_14", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第5课 列表与字典", "unit_intro": "学习Python数据结构，制作通讯录管理程序",
			"course_id": "ch_course_02", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 5},
		// 机器人搭建与竞赛
		{"id": "ch_unit_20", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1课 认识机器人", "unit_intro": "了解机器人组成结构，认识主控板和传感器",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 1},
		{"id": "ch_unit_21", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2课 LED与蜂鸣器", "unit_intro": "学习点亮LED和驱动蜂鸣器发声",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 2},
		{"id": "ch_unit_22", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3课 电机与轮子", "unit_intro": "学习驱动电机，让机器人动起来",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"course_work_type": 1, "map_x": 0, "map_y": 0, "order_num": 3},
		{"id": "ch_unit_23", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4课 循线行走", "unit_intro": "学习红外传感器，让机器人沿黑线行走",
			"course_id": "ch_course_03", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 4},
		// Scratch 游戏设计专项
		{"id": "ch_unit_30", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1课 接苹果游戏", "unit_intro": "学习碰撞检测，制作经典接苹果游戏",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 1},
		{"id": "ch_unit_31", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2课 打地鼠游戏", "unit_intro": "学习随机出现和计分系统，制作打地鼠",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 2},
		{"id": "ch_unit_32", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第3课 迷宫探险", "unit_intro": "学习方向键控制和墙壁碰撞，制作迷宫游戏",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 3},
		{"id": "ch_unit_33", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第4课 多关卡设计", "unit_intro": "学习关卡切换和难度递增，制作多关卡游戏",
			"course_id": "ch_course_04", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 4},
		// Python 数据可视化入门
		{"id": "ch_unit_40", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1课 初识Matplotlib", "unit_intro": "安装Matplotlib，绘制第一条折线图",
			"course_id": "ch_course_05", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 1},
		{"id": "ch_unit_41", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2课 柱状图与饼图", "unit_intro": "学习柱状图、饼图和散点图的绘制",
			"course_id": "ch_course_05", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 2},
		// 信息学奥赛 CSP-J 入门
		{"id": "ch_unit_60", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第1课 C++ 起步", "unit_intro": "搭建 C++ 编译环境，理解程序编译运行流程",
			"course_id": "ch_course_06", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 1},
		{"id": "ch_unit_61", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "sys_org_code": "A01", "del_flag": 0,
			"unit_name": "第2课 顺序与分支", "unit_intro": "C++ 基本语句、if 分支与 CSP-J 真题入门",
			"course_id": "ch_course_06", "course_video_source": 1, "show_course_video": 1, "show_course_case": 1, "show_course_ppt": 0, "show_course_plan": 0,
			"map_x": 0, "map_y": 0, "order_num": 2},
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
			"start_time": "2026-09-01 00:00:00", "end_time": "2027-01-31 23:59:59", "sender": "admin", "priority": "H",
			"msg_category": "1", "msg_type": "ALL", "send_status": "1", "send_time": "2026-09-01 08:00:00",
			"del_flag": "0", "create_by": "admin", "create_time": "2026-09-07 00:00:00", "msg_abstract": "秋季学期开学通知"},
		{"id": "ch_ann_02", "titile": "Scratch 创意编程启蒙班课程安排",
			"msg_content": "<p>Scratch 启蒙班每周六上午 9:00-11:00 上课，请提前 10 分钟到教室（2楼 Scratch教室）。</p>",
			"start_time": "2026-09-01 00:00:00", "end_time": "2027-01-31 23:59:59", "sender": "teacher1", "priority": "M",
			"msg_category": "1", "msg_type": "ALL", "send_status": "1", "send_time": "2026-09-05 08:00:00",
			"del_flag": "0", "create_by": "teacher1", "create_time": "2026-09-07 00:00:00", "msg_abstract": "Scratch班课程安排"},
		{"id": "ch_ann_03", "titile": "Python 进阶班新学期课程大纲发布",
			"msg_content": "<p>Python 进阶班新学期将学习条件判断、循环结构、列表与字典等内容，请同学们提前预习第一课内容。</p>",
			"start_time": "2026-09-01 00:00:00", "end_time": "2027-01-31 23:59:59", "sender": "teacher2", "priority": "M",
			"msg_category": "1", "msg_type": "ALL", "send_status": "1", "send_time": "2026-09-05 08:00:00",
			"del_flag": "0", "create_by": "teacher2", "create_time": "2026-09-07 00:00:00", "msg_abstract": "Python班课程大纲"},
		{"id": "ch_ann_04", "titile": "机器人竞赛班即将开启报名",
			"msg_content": "<p>机器人竞赛班将于 10 月开始报名，名额有限（8人/班），请有意向的同学尽快联系陈教练。</p>",
			"start_time": "2026-09-01 00:00:00", "end_time": "2026-10-31 23:59:59", "sender": "teacher3", "priority": "H",
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
