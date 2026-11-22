-- ============================================================
-- 春晖青少年编程中心 种子 · BASE 层（系统表）
-- 产品归属：武胜县沿口春夏秋信息技术中心（个体工商户）
-- 统一社会信用代码：92511622MAKLLHTT5D
-- 经营场所：四川省广安市武胜县沿口镇上东街257号（邮编638400）
--
-- 层级说明（2026-11-22 拆分）：
--   base = 系统表（配置/部门/角色/用户/授权 + 管理后台基础数据 + 菜单可见性）
--   demo = 业务示例数据（课程/单元/公告/资讯/作品/评论）
-- 执行顺序：先 base 后 demo；demo 依赖 base 写入的部门与用户。
-- 密码均为 bcrypt（$2a$12），与 Go 服务端 passwd.Hash 语义一致。
--
-- 等效工具：cd goframe-server && go run ./tools/chseed -layer base
-- ============================================================

-- ============================================================
-- 1. 系统配置（sys_config）——品牌与站点信息
-- ============================================================
DELETE FROM sys_config WHERE config_key IN
 ('_address','_phone','_linkman','brandName','_indexHtml','_homeHtml','_workShareHtml','_defaultRole','_defaultDepart','_copyRight');
INSERT INTO sys_config (id, config_key, config_value, config_enabled, `comment`) VALUES
 ('ch_cfg_01', 'brandName',      '春晖青少年编程中心', 1, '品牌名称'),
 ('ch_cfg_02', '_address',       '四川省广安市武胜县沿口镇上东街257号（邮编638400）', 1, '经营场所地址'),
 ('ch_cfg_03', '_phone',         '', 1, '联系电话'),
 ('ch_cfg_04', '_linkman',       '春晖管理员', 1, '联系人'),
 ('ch_cfg_05', '_defaultRole',   'student', 1, '注册默认角色'),
 ('ch_cfg_06', '_defaultDepart', 'ch_depart_03', 1, '注册默认班级'),
 ('ch_cfg_07', '_copyRight',     '© 2026 武胜县沿口春夏秋信息技术中心（个体工商户）· 统一社会信用代码 92511622MAKLLHTT5D', 1, '版权信息'),
 ('ch_cfg_08', '_indexHtml',
  '<div style="max-width:1200px;margin:0 auto;padding:40px 20px;font-family:Microsoft YaHei,sans-serif;">'
  '<div style="text-align:center;margin-bottom:30px;">'
  '<h1 style="color:#2c3e50;font-size:36px;margin-bottom:10px;">春晖青少年编程中心</h1>'
  '<p style="color:#7f8c8d;font-size:16px;">Scratch 创意编程 · Python 代码进阶 · 机器人竞赛</p>'
  '<p style="color:#95a5a6;font-size:14px;">让每个孩子拥有创造未来的能力</p></div>'
  '<div style="display:flex;justify-content:space-around;text-align:center;flex-wrap:wrap;">'
  '<div style="width:30%;background:#f8f9fa;border-radius:8px;padding:20px;margin:10px;">'
  '<h3 style="color:#3498db;">Scratch 启蒙</h3><p style="color:#666;">6-9岁 · 图形化编程<br>培养逻辑思维</p></div>'
  '<div style="width:30%;background:#f8f9fa;border-radius:8px;padding:20px;margin:10px;">'
  '<h3 style="color:#e67e22;">Python 进阶</h3><p style="color:#666;">10-14岁 · 代码编程<br>掌握算法思维</p></div>'
  '<div style="width:30%;background:#f8f9fa;border-radius:8px;padding:20px;margin:10px;">'
  '<h3 style="color:#e74c3c;">机器人竞赛</h3><p style="color:#666;">8-14岁 · 硬件编程<br>锻炼动手能力</p></div>'
  '</div>'
  '<div style="text-align:center;margin-top:30px;padding:15px;background:#f0f3f7;border-radius:8px;">'
  '<p style="color:#666;font-size:14px;">武胜县沿口春夏秋信息技术中心（个体工商户）<br>统一社会信用代码：92511622MAKLLHTT5D<br>经营场所：四川省广安市武胜县沿口镇上东街257号（邮编638400）</p></div></div>', 1, '首页富文本'),
 ('ch_cfg_09', '_homeHtml',
  '<div style="max-width:1200px;margin:0 auto;padding:30px 20px;font-family:Microsoft YaHei,sans-serif;">'
  '<h2 style="text-align:center;color:#2c3e50;">欢迎来到春晖青少年编程中心</h2>'
  '<p style="text-align:center;color:#7f8c8d;font-size:18px;margin:20px 0;">让每个孩子拥有创造未来的能力</p>'
  '<div style="display:flex;flex-wrap:wrap;justify-content:center;gap:20px;">'
  '<div style="width:350px;background:linear-gradient(135deg,#667eea,#764ba2);border-radius:12px;padding:25px;color:white;">'
  '<h3>Scratch 创意编程</h3><p>图形化积木式编程，适合 6-9 岁零基础学员<br>培养逻辑思维与创造力</p></div>'
  '<div style="width:350px;background:linear-gradient(135deg,#f093fb,#f5576c);border-radius:12px;padding:25px;color:white;">'
  '<h3>Python 代码进阶</h3><p>真正的代码编程，适合 10-14 岁学员<br>从游戏到人工智能</p></div>'
  '<div style="width:350px;background:linear-gradient(135deg,#4facfe,#00f2fe);border-radius:12px;padding:25px;color:white;">'
  '<h3>机器人竞赛</h3><p>硬件编程与竞赛训练，适合 8-14 岁<br>锻炼动手与团队协作</p></div>'
  '</div></div>', 1, '自定义首页'),
 ('ch_cfg_10', '_workShareHtml',
  '<div style="max-width:1000px;margin:0 auto;padding:20px;font-family:Microsoft YaHei,sans-serif;">'
  '<h2 style="text-align:center;color:#2c3e50;">春晖青少年编程中心 · 作品展示</h2>'
  '<p style="text-align:center;color:#95a5a6;">每一个作品都是孩子创造力的结晶</p></div>', 1, '作品分享页富文本');

-- ============================================================
-- 2. 部门/班级（sys_depart）——春晖组织架构
-- ============================================================
DELETE FROM sys_depart;
INSERT INTO sys_depart (id, parent_id, depart_name, depart_name_en, depart_name_abbr, depart_order, description, org_category, org_type, org_code, mobile, fax, address, memo, status, del_flag, create_by, create_time) VALUES
 ('ch_depart_01', '',              '春晖青少年编程中心', 'Chunhui Youth Coding Center', '春晖编程', 0, '武胜县沿口春夏秋信息技术中心（个体工商户）', '1', '1', 'A01',       '', '', '四川省广安市武胜县沿口镇上东街257号', '统一社会信用代码：92511622MAKLLHTT5D', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_02', 'ch_depart_01', '教学部',             NULL, NULL, 1, '课程教学与教研', '2', '2', 'A01A01',    '', '', '', '', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_03', 'ch_depart_01', 'Scratch启蒙班',      NULL, NULL, 2, 'Scratch 图形化编程（6-9岁）', '3', '3', 'A01A02',    '', '', '上东街257号 2楼', '', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_04', 'ch_depart_01', 'Python进阶班',       NULL, NULL, 3, 'Python 代码编程（10-14岁）', '3', '3', 'A01A03',    '', '', '上东街257号 3楼', '', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_05', 'ch_depart_01', '机器人竞赛班',       NULL, NULL, 4, '机器人搭建与竞赛', '3', '3', 'A01A04',    '', '', '上东街257号 1楼', '', '0', '0', 'admin', '2026-09-07 00:00:00');

-- ============================================================
-- 3. 角色（sys_role）
-- ============================================================
DELETE FROM sys_role;
INSERT INTO sys_role (id, role_name, role_code, description, create_by, create_time, role_level) VALUES
 ('ch_role_admin',    '管理员',   'admin',    '春晖青少年编程中心管理员', 'admin', '2026-09-07 00:00:00', 99),
 ('ch_role_dev',      '开发调试', 'dev',      '系统开发调试',             'admin', '2026-09-07 00:00:00', 98),
 ('ch_role_teacher',  '教师',     'teacher',  '授课教师',                 'admin', '2026-09-07 00:00:00', 50),
 ('ch_role_student',  '学生',     'student',  '注册学生',                 'admin', '2026-09-07 00:00:00', 0);

-- ============================================================
-- 4. 用户（sys_user）——密码 bcrypt
-- ============================================================
DELETE FROM sys_user;
INSERT INTO sys_user (id, username, realname, password, salt, avatar, birthday, sex, email, phone, org_code, status, del_flag, activiti_sync, work_no, post, school, telephone, create_by, create_time, user_identity, depart_ids) VALUES
 ('ch_user_admin',    'admin',    '春晖管理员', '$2a$12$LWLf.Ti0FX/ajbsEVuUtC.227qHk3IryExS9MHIvfQJ6GuvbYHY3G', '', NULL, NULL, 1, NULL, NULL, 'A01',       1, 0, 1, 'CH001', '校长', '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 2, 'ch_depart_01,ch_depart_02'),
 ('ch_user_teacher1', 'teacher1', '王老师',     '$2a$12$N2gprWtzbRX/axPrmQ6lg.IL16PeAyflNXSSUL/gkmrWOaJ4NDvXK', '', NULL, '1995-03-15', 1, NULL, NULL, 'A01',       1, 0, 1, 'CH002', 'Scratch讲师', '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 2, 'ch_depart_02,ch_depart_03'),
 ('ch_user_teacher2', 'teacher2', '李老师',     '$2a$12$N2gprWtzbRX/axPrmQ6lg.IL16PeAyflNXSSUL/gkmrWOaJ4NDvXK', '', NULL, '1998-08-22', 2, NULL, NULL, 'A01',       1, 0, 1, 'CH003', 'Python讲师', '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 2, 'ch_depart_02,ch_depart_04'),
 ('ch_user_teacher3', 'teacher3', '陈教练',     '$2a$12$N2gprWtzbRX/axPrmQ6lg.IL16PeAyflNXSSUL/gkmrWOaJ4NDvXK', '', NULL, '1992-12-05', 1, NULL, NULL, 'A01',       1, 0, 1, 'CH004', '机器人教练', '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 2, 'ch_depart_02,ch_depart_05'),
 ('ch_user_stu1',     'student1', '张小明',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2018-05-20', 1, NULL, NULL, 'A01A02',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_03'),
 ('ch_user_stu2',     'student2', '李小红',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2019-01-10', 2, NULL, NULL, 'A01A02',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_03'),
 ('ch_user_stu3',     'student3', '陈大壮',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2016-09-01', 1, NULL, NULL, 'A01A03',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_04'),
 ('ch_user_stu4',     'student4', '刘小慧',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2017-04-18', 2, NULL, NULL, 'A01A03',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_04'),
 ('ch_user_stu5',     'student5', '周小杰',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2015-07-12', 1, NULL, NULL, 'A01A04',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_05'),
 ('ch_user_stu6',     'student6', '吴小婷',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2016-02-25', 2, NULL, NULL, 'A01A04',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_05');

-- ============================================================
-- 4.5 管理员角色授权（sys_role_permission）
-- QueryByUser 严格按 sys_role_permission 过滤（无 admin 旁路），
-- 不绑定则动态管理菜单为空（与 Java 演示数据的角色授权模式一致）。
-- ============================================================
DELETE FROM sys_role_permission WHERE role_id = 'ch_role_admin';
INSERT INTO sys_role_permission (id, role_id, permission_id)
SELECT MD5(CONCAT('ch_rp_', id)), 'ch_role_admin', id FROM sys_permission;

-- ============================================================
-- 5. 用户-角色（sys_user_role）
-- ============================================================
DELETE FROM sys_user_role;
INSERT INTO sys_user_role (id, user_id, role_id) VALUES
 ('ch_ur_01', 'ch_user_admin',    'ch_role_admin'),
 ('ch_ur_02', 'ch_user_teacher1', 'ch_role_teacher'),
 ('ch_ur_03', 'ch_user_teacher2', 'ch_role_teacher'),
 ('ch_ur_04', 'ch_user_teacher3', 'ch_role_teacher'),
 ('ch_ur_05', 'ch_user_stu1',     'ch_role_student'),
 ('ch_ur_06', 'ch_user_stu2',     'ch_role_student'),
 ('ch_ur_07', 'ch_user_stu3',     'ch_role_student'),
 ('ch_ur_08', 'ch_user_stu4',     'ch_role_student'),
 ('ch_ur_09', 'ch_user_stu5',     'ch_role_student'),
 ('ch_ur_10', 'ch_user_stu6',     'ch_role_student');

-- ============================================================
-- 6. 用户-班级（sys_user_depart）
-- ============================================================
DELETE FROM sys_user_depart;
INSERT INTO sys_user_depart (ID, user_id, dep_id) VALUES
 ('ch_ud_01', 'ch_user_admin',    'ch_depart_01'),
 ('ch_ud_02', 'ch_user_teacher1', 'ch_depart_02'),
 ('ch_ud_03', 'ch_user_teacher1', 'ch_depart_03'),
 ('ch_ud_04', 'ch_user_teacher2', 'ch_depart_02'),
 ('ch_ud_05', 'ch_user_teacher2', 'ch_depart_04'),
 ('ch_ud_06', 'ch_user_teacher3', 'ch_depart_02'),
 ('ch_ud_07', 'ch_user_teacher3', 'ch_depart_05'),
 ('ch_ud_08', 'ch_user_stu1',     'ch_depart_03'),
 ('ch_ud_09', 'ch_user_stu2',     'ch_depart_03'),
 ('ch_ud_10', 'ch_user_stu3',     'ch_depart_04'),
 ('ch_ud_11', 'ch_user_stu4',     'ch_depart_04'),
 ('ch_ud_12', 'ch_user_stu5',     'ch_depart_05'),
 ('ch_ud_13', 'ch_user_stu6',     'ch_depart_05');

-- ============================================================
-- 14. 管理后台基础数据（职务/数据源/消息模板/部门角色）
-- 全量替换 Mock 后这些表被清空 → 管理页显示「暂无数据」，
-- 补齐后「职务管理 / 多数据源管理 / 模板管理 / 通讯录-部门角色」可用。
-- ============================================================
DELETE FROM sys_position;
INSERT INTO sys_position (id, code, name, post_rank, company_id, sys_org_code, create_by, create_time) VALUES
 ('ch_pos_01', 'principal', '校长',     '1', 'ch_depart_01', 'A01', 'admin', '2026-09-07 00:00:00'),
 ('ch_pos_02', 'teacher',   '讲师',     '2', 'ch_depart_01', 'A01', 'admin', '2026-09-07 00:00:00'),
 ('ch_pos_03', 'coach',     '竞赛教练', '2', 'ch_depart_01', 'A01', 'admin', '2026-09-07 00:00:00'),
 ('ch_pos_04', 'assistant', '助教',     '3', 'ch_depart_01', 'A01', 'admin', '2026-09-07 00:00:00'),
 ('ch_pos_05', 'affairs',   '教务',     '3', 'ch_depart_02', 'A01', 'admin', '2026-09-07 00:00:00');

DELETE FROM sys_data_source;
INSERT INTO sys_data_source (id, code, name, remark, db_type, db_driver, db_url, db_name, db_username, db_password, create_by, create_time, sys_org_code) VALUES
 ('ch_ds_master', 'master', '春晖教学平台主库', '本系统默认数据源（teachingopen）', 'mysql', 'com.mysql.cj.jdbc.Driver',
  'jdbc:mysql://127.0.0.1:13306/teachingopen?useUnicode=true&characterEncoding=utf8&serverTimezone=Asia/Shanghai',
  'teachingopen', 'teaching', '', 'admin', '2026-09-07 00:00:00', 'A01');

DELETE FROM sys_sms_template;
INSERT INTO sys_sms_template (id, template_name, template_code, template_type, template_content, template_test_json, create_by, create_time) VALUES
 ('ch_tpl_01', '账号注册验证码', 'SMS_VERIFY_CODE', '1', '您的验证码为 ${code}，10 分钟内有效。', '{"code":"123456"}', 'admin', '2026-09-07 00:00:00'),
 ('ch_tpl_02', '作业布置通知', 'WORK_ASSIGN_NOTICE', '2', '【春晖编程】老师布置了新作业：${workName}，请在 ${deadline} 前提交。', '{"workName":"第2关 角色与运动","deadline":"本周六"}', 'admin', '2026-09-07 00:00:00'),
 ('ch_tpl_03', '作品被点赞通知', 'WORK_LIKE_NOTICE', '2', '【春晖编程】你的作品《${workName}》获得了新的点赞，当前 ${starNum} 赞！', '{"workName":"接苹果","starNum":"13"}', 'admin', '2026-09-07 00:00:00');

DELETE FROM sys_depart_role;
INSERT INTO sys_depart_role (id, depart_id, role_name, role_code, description, create_by, create_time) VALUES
 ('ch_dr_01', 'ch_depart_03', 'Scratch班管理员',  'scratch_admin', 'Scratch启蒙班班级管理', 'admin', '2026-09-07 00:00:00'),
 ('ch_dr_02', 'ch_depart_04', 'Python班管理员',   'python_admin',  'Python进阶班班级管理',  'admin', '2026-09-07 00:00:00'),
 ('ch_dr_03', 'ch_depart_05', '机器人班管理员',   'robot_admin',   '机器人竞赛班班级管理',  'admin', '2026-09-07 00:00:00');

-- ============================================================
-- 15. 未启用模块菜单隐藏（避免点击后 404 造成「系统坏了」的错觉）
-- 在线表单（online/*）、工作流（bpm/extbpm）、404 示例页在本部署未实现，
-- 统一置 hidden=1（Jeecg 前端不再渲染该菜单项，路由与数据保留）。
-- ============================================================
UPDATE sys_permission SET hidden = 1
 WHERE hidden = 0
   AND (url LIKE '/online/%'
        OR url LIKE '/modules/bpm/%'
        OR url LIKE '/modules/extbpm/%'
        OR url = '/exception/404');
