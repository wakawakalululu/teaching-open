-- ============================================================
-- 春晖青少年编程中心 真实数据种子
-- 产品归属：武胜县沿口春夏秋信息技术中心（个体工商户）
-- 统一社会信用代码：92511622MAKLLHTT5D
-- 成立日期：2026-09-07  登记状态：存续
--
-- 用法：在 teachingopen2.8.sql 全量导入后执行本文件，
-- 将 Mock 数据替换为春晖真实数据。
-- 密码均为 bcrypt（$2a$12），与 Go 服务端 passwd.Hash 语义一致。
-- ============================================================

-- ============================================================
-- 1. 系统配置（sys_config）——品牌与站点信息
-- ============================================================
DELETE FROM sys_config WHERE config_key IN
 ('_address','_phone','_linkman','brandName','_indexHtml','_homeHtml','_workShareHtml','_defaultRole','_defaultDepart');
INSERT INTO sys_config (id, config_key, config_value, config_enabled, `comment`) VALUES
 ('ch_cfg_01', 'brandName',      '春晖青少年编程中心', 1, '品牌名称'),
 ('ch_cfg_02', '_address',       '四川省广安市武胜县沿口镇', 1, '地址'),
 ('ch_cfg_03', '_phone',         '', 1, '联系电话'),
 ('ch_cfg_04', '_linkman',       '春晖管理员', 1, '联系人'),
 ('ch_cfg_05', '_defaultRole',   'student', 1, '注册默认角色'),
 ('ch_cfg_06', '_defaultDepart', 'ch_depart_03', 1, '注册默认班级'),
 ('ch_cfg_07', '_indexHtml',
  '<h1 style="text-align:center;">春晖青少年编程中心</h1><p style="text-align:center;">Scratch · Python · 机器人编程</p><p style="text-align:center;color:#999;">武胜县沿口春夏秋信息技术中心（个体工商户）</p><p style="text-align:center;color:#999;">统一社会信用代码：92511622MAKLLHTT5D</p>', 1, '首页富文本'),
 ('ch_cfg_08', '_homeHtml',
  '<h2 style="text-align:center;">欢迎来到春晖青少年编程中心</h2><h3 style="text-align:center;">让每个孩子拥有创造未来的能力</h3>', 1, '自定义首页'),
 ('ch_cfg_09', '_workShareHtml',
  '<h2 style="text-align:center;">春晖青少年编程中心 · 作品展示</h2>', 1, '作品分享页富文本');

-- ============================================================
-- 2. 部门/班级（sys_depart）——春晖组织架构
--    org_code 遵循 YouBian 流水码：根 A01 起，每级 3 字符
--    org_category: 1=组织机构 2=岗位 3=班级
--    org_type: 1=一级 2=子部门 3=班级
-- ============================================================
DELETE FROM sys_depart;
INSERT INTO sys_depart (id, parent_id, depart_name, depart_name_en, depart_name_abbr, depart_order, description, org_category, org_type, org_code, mobile, fax, address, memo, status, del_flag, create_by, create_time) VALUES
 ('ch_depart_01', '',              '春晖青少年编程中心', 'Chunhui Youth Coding Center', '春晖编程', 0, '武胜县沿口春夏秋信息技术中心（个体工商户）', '1', '1', 'A01',       '', '', '四川省广安市武胜县沿口镇', '统一社会信用代码：92511622MAKLLHTT5D', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_02', 'ch_depart_01', '教学部',             NULL, NULL, 1, '课程教学与教研', '2', '2', 'A01A01',    '', '', '', '', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_03', 'ch_depart_01', 'Scratch启蒙班',      NULL, NULL, 2, 'Scratch 图形化编程（6-9岁）', '3', '3', 'A01A02',    '', '', '', '', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_04', 'ch_depart_01', 'Python进阶班',       NULL, NULL, 3, 'Python 代码编程（10-14岁）', '3', '3', 'A01A03',    '', '', '', '', '0', '0', 'admin', '2026-09-07 00:00:00'),
 ('ch_depart_05', 'ch_depart_01', '机器人竞赛班',       NULL, NULL, 4, '机器人搭建与竞赛', '3', '3', 'A01A04',    '', '', '', '', '0', '0', 'admin', '2026-09-07 00:00:00');

-- ============================================================
-- 3. 角色（sys_role）——春晖角色体系
-- ============================================================
DELETE FROM sys_role;
INSERT INTO sys_role (id, role_name, role_code, description, create_by, create_time, role_level) VALUES
 ('ch_role_admin',    '管理员',   'admin',    '春晖青少年编程中心管理员', 'admin', '2026-09-07 00:00:00', 99),
 ('ch_role_dev',      '开发调试', 'dev',      '系统开发调试',             'admin', '2026-09-07 00:00:00', 98),
 ('ch_role_teacher',  '教师',     'teacher',  '授课教师',                 'admin', '2026-09-07 00:00:00', 50),
 ('ch_role_student',  '学生',     'student',  '注册学生',                 'admin', '2026-09-07 00:00:00', 0);

-- ============================================================
-- 4. 用户（sys_user）——春晖真实账号
--    密码 bcrypt（$2a$12），Go passwd.Hash 生成
--    user_identity: 1=普通 2=上级（教师/管理员）
-- ============================================================
DELETE FROM sys_user;
INSERT INTO sys_user (id, username, realname, password, salt, avatar, birthday, sex, email, phone, org_code, status, del_flag, activiti_sync, work_no, post, school, telephone, create_by, create_time, user_identity, depart_ids) VALUES
 -- 管理员
 ('ch_user_admin',    'admin',    '春晖管理员', '$2a$12$LWLf.Ti0FX/ajbsEVuUtC.227qHk3IryExS9MHIvfQJ6GuvbYHY3G', '', NULL, NULL, 1, '', '', 'A01',       1, 0, 1, 'CH001', '校长', '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 2, 'ch_depart_01,ch_depart_02'),
 -- 教师
 ('ch_user_teacher1', 'teacher1', '王老师',     '$2a$12$N2gprWtzbRX/axPrmQ6lg.IL16PeAyflNXSSUL/gkmrWOaJ4NDvXK', '', NULL, '1995-03-15', 1, '', '', 'A01',       1, 0, 1, 'CH002', 'Scratch讲师', '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 2, 'ch_depart_02,ch_depart_03'),
 ('ch_user_teacher2', 'teacher2', '李老师',     '$2a$12$N2gprWtzbRX/axPrmQ6lg.IL16PeAyflNXSSUL/gkmrWOaJ4NDvXK', '', NULL, '1998-08-22', 2, '', '', 'A01',       1, 0, 1, 'CH003', 'Python讲师', '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 2, 'ch_depart_02,ch_depart_04'),
 -- 学生
 ('ch_user_stu1',     'student1', '张小明',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2018-05-20', 1, '', '', 'A01A02',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_03'),
 ('ch_user_stu2',     'student2', '李小红',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2019-01-10', 2, '', '', 'A01A02',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_03'),
 ('ch_user_stu3',     'student3', '陈大壮',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2016-09-01', 1, '', '', 'A01A03',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_04'),
 ('ch_user_stu4',     'student4', '刘小慧',     '$2a$12$8oFhzs/lBg4XyO5KOQTFTeOpkB51zq.FFje9iA7J4iouC/bsyPjHC', '', NULL, '2017-04-18', 2, '', '', 'A01A03',    1, 0, 1, NULL, NULL, '春晖青少年编程中心', '', 'admin', '2026-09-07 00:00:00', 1, 'ch_depart_04');

-- ============================================================
-- 5. 用户-角色（sys_user_role）
-- ============================================================
DELETE FROM sys_user_role;
INSERT INTO sys_user_role (id, user_id, role_id) VALUES
 ('ch_ur_01', 'ch_user_admin',    'ch_role_admin'),
 ('ch_ur_02', 'ch_user_teacher1', 'ch_role_teacher'),
 ('ch_ur_03', 'ch_user_teacher2', 'ch_role_teacher'),
 ('ch_ur_04', 'ch_user_stu1',     'ch_role_student'),
 ('ch_ur_05', 'ch_user_stu2',     'ch_role_student'),
 ('ch_ur_06', 'ch_user_stu3',     'ch_role_student'),
 ('ch_ur_07', 'ch_user_stu4',     'ch_role_student');

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
 ('ch_ud_06', 'ch_user_stu1',     'ch_depart_03'),
 ('ch_ud_07', 'ch_user_stu2',     'ch_depart_03'),
 ('ch_ud_08', 'ch_user_stu3',     'ch_depart_04'),
 ('ch_ud_09', 'ch_user_stu4',     'ch_depart_04');

-- ============================================================
-- 7. 课程（teaching_course）——春晖课程体系
-- ============================================================
DELETE FROM teaching_course;
INSERT INTO teaching_course (id, create_by, create_time, sys_org_code, del_flag, course_name, course_desc, course_icon, course_cover, show_type, course_map, is_shared, show_home, order_num, depart_ids, course_type, course_category) VALUES
 ('ch_course_01', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Scratch创意编程启蒙',
  '<h2>Scratch 图形化编程</h2><p>适合 6-9 岁零基础学员，通过拖拽积木块的方式学习编程思维，制作动画、故事和小游戏。</p><p><strong>课程目标：</strong>培养逻辑思维、创造力与问题解决能力。</p>',
  NULL, NULL, 1, NULL, 0, 1, 1, 'ch_depart_03', '1', '1'),
 ('ch_course_02', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Python 代码编程进阶',
  '<h2>Python 代码编程</h2><p>适合 10-14 岁有一定编程基础的学员，学习 Python 语法、算法与项目开发。</p><p><strong>课程目标：</strong>掌握 Python 编程、数据结构与算法思维。</p>',
  NULL, NULL, 1, NULL, 0, 1, 2, 'ch_depart_04', '2', '2'),
 ('ch_course_03', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  '机器人搭建与竞赛',
  '<h2>机器人编程竞赛</h2><p>适合 8-14 岁学员，学习机器人搭建、传感器编程与竞赛策略。</p><p><strong>课程目标：</strong>培养动手能力与团队协作。</p>',
  NULL, NULL, 1, NULL, 0, 1, 3, 'ch_depart_05', '2', '1');

-- ============================================================
-- 8. 课程单元（teaching_course_unit）
-- ============================================================
DELETE FROM teaching_course_unit;
INSERT INTO teaching_course_unit (id, create_by, create_time, sys_org_code, del_flag, unit_name, unit_intro, course_id, course_video_source, show_course_video, course_case, show_course_case, course_ppt, show_course_ppt, course_work_type, course_work, course_work_answer, course_plan, show_course_plan, map_x, map_y, media_content, order_num) VALUES
 ('ch_unit_01', 'admin', '2026-09-07 00:00:00', 'A01', 0, '初识Scratch', '认识 Scratch 界面与基本操作', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 1),
 ('ch_unit_02', 'admin', '2026-09-07 00:00:00', 'A01', 0, '角色与舞台', '学习角色、舞台与坐标系', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 2),
 ('ch_unit_03', 'admin', '2026-09-07 00:00:00', 'A01', 0, 'Python起步', '安装 Python 与第一个程序', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 1),
 ('ch_unit_04', 'admin', '2026-09-07 00:00:00', 'A01', 0, '变量与运算', '学习变量赋值与四则运算', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 2);

-- ============================================================
-- 9. 课程-班级授权（teaching_course_dept）
-- ============================================================
DELETE FROM teaching_course_dept;
INSERT INTO teaching_course_dept (id, create_by, create_time, sys_org_code, dept_id, course_id, open_time) VALUES
 ('ch_cd_01', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_03', 'ch_course_01', '2026-09-10 08:00:00'),
 ('ch_cd_02', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_04', 'ch_course_02', '2026-09-10 08:00:00'),
 ('ch_cd_03', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_05', 'ch_course_03', '2026-09-10 08:00:00');

-- ============================================================
-- 10. 公告（sys_announcement）
-- ============================================================
DELETE FROM sys_announcement;
INSERT INTO sys_announcement (id, titile, msg_content, start_time, end_time, sender, priority, msg_category, msg_type, send_status, send_time, cancel_time, del_flag, bus_type, bus_id, open_type, open_page, create_by, create_time, msg_abstract) VALUES
 ('ch_ann_01', '春晖青少年编程中心 2026 秋季学期开学通知',
  '<p>各位家长、同学：</p><p>春晖青少年编程中心 2026 年秋季学期将于 9 月 10 日正式开课，请按时到校。</p><p>—— 武胜县沿口春夏秋信息技术中心</p>',
  '2026-09-01 00:00:00', '2027-01-31 23:59:59', 'admin', 'H', '1', 'ALL', '1', '2026-09-01 08:00:00', NULL, '0', NULL, NULL, NULL, NULL, 'admin', '2026-09-07 00:00:00', '秋季学期开学通知'),
 ('ch_ann_02', 'Scratch 创意编程启蒙班课程安排',
  '<p>Scratch 启蒙班每周六上午 9:00-11:00 上课，请提前 10 分钟到教室。</p>',
  '2026-09-01 00:00:00', '2027-01-31 23:59:59', 'teacher1', 'M', '1', 'ALL', '1', '2026-09-05 08:00:00', NULL, '0', NULL, NULL, NULL, NULL, 'teacher1', '2026-09-07 00:00:00', 'Scratch班课程安排');

-- ============================================================
-- 11. 用户通告阅读标记（sys_announcement_send）
-- ============================================================
DELETE FROM sys_announcement_send;
INSERT INTO sys_announcement_send (id, annt_id, user_id, read_flag, read_time) VALUES
 ('ch_as_01', 'ch_ann_01', 'ch_user_teacher1', '0', NULL),
 ('ch_as_02', 'ch_ann_01', 'ch_user_teacher2', '0', NULL),
 ('ch_as_03', 'ch_ann_01', 'ch_user_stu1',     '0', NULL),
 ('ch_as_04', 'ch_ann_01', 'ch_user_stu2',     '0', NULL),
 ('ch_as_05', 'ch_ann_02', 'ch_user_stu1',     '0', NULL);

-- ============================================================
-- 12. 字典项补充——春晖品牌相关字典
-- ============================================================
-- 字典 sys_dict 中 course_type/course_category 等已在 V2.8 种子数据中定义，
-- 此处仅补充 春晖 专属字典项（如有需要）。

-- ============================================================
-- 13. 清理 Mock 用户关联数据
-- ============================================================
DELETE FROM sys_user_role WHERE user_id NOT IN (SELECT id FROM sys_user);
DELETE FROM sys_user_depart WHERE user_id NOT IN (SELECT id FROM sys_user);
