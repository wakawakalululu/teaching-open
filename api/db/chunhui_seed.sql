-- ============================================================
-- 春晖青少年编程中心 真实数据种子
-- 产品归属：武胜县沿口春夏秋信息技术中心（个体工商户）
-- 统一社会信用代码：92511622MAKLLHTT5D
-- 成立日期：2026-09-07  登记状态：存续
-- 经营场所：四川省广安市武胜县沿口镇上东街257号（邮编638400）
-- 参考布局：verveit.010124.xyz
--
-- 用法：在 teachingopen2.8.sql 全量导入后执行本文件，
-- 将 Mock 数据替换为春晖真实数据。
-- 密码均为 bcrypt（$2a$12），与 Go 服务端 passwd.Hash 语义一致。
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
-- 7. 课程（teaching_course）——春晖课程体系（6门）
-- ============================================================
DELETE FROM teaching_course;
INSERT INTO teaching_course (id, create_by, create_time, sys_org_code, del_flag, course_name, course_desc, course_icon, course_cover, show_type, course_map, is_shared, show_home, order_num, depart_ids, course_type, course_category) VALUES
 ('ch_course_01', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Scratch 创意编程启蒙',
  '<h2>Scratch 图形化编程启蒙课程</h2><p>适合 6-9 岁零基础学员。通过 MIT 开发的 Scratch 图形化编程工具，孩子可以像搭积木一样拖拽代码块，创作属于自己的互动故事、动画和小游戏。</p><h3>课程亮点</h3><ul><li>零门槛入门，无需打字基础</li><li>项目式学习，每节课完成一个作品</li><li>培养逻辑思维、创造力与问题解决能力</li><li>小班教学（8人/班），确保每个孩子得到关注</li></ul><h3>适合年龄</h3><p>6-9 岁（小学 1-3 年级）</p>',
  NULL, NULL, 1, NULL, 0, 1, 1, 'ch_depart_03', '1', '1'),
 ('ch_course_02', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Python 代码编程进阶',
  '<h2>Python 代码编程进阶课程</h2><p>适合 10-14 岁有一定编程基础的学员。从图形化编程过渡到真正的代码编程，学习 Python 语法、数据结构、算法与项目开发。</p><h3>课程亮点</h3><ul><li>从积木到代码的平滑过渡</li><li>涵盖 Turtle绘图、Pygame游戏开发、数据分析入门</li><li>培养算法思维与工程能力</li><li>完成多个实战项目</li></ul><h3>适合年龄</h3><p>10-14 岁（小学 4 年级 - 初中）</p>',
  NULL, NULL, 1, NULL, 0, 1, 2, 'ch_depart_04', '2', '2'),
 ('ch_course_03', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  '机器人搭建与竞赛',
  '<h2>机器人搭建与竞赛课程</h2><p>适合 8-14 岁学员。学习机器人机械结构搭建、传感器应用与编程控制，备战各类机器人竞赛。</p><h3>课程亮点</h3><ul><li>硬件+软件结合，动手与动脑并行</li><li>竞赛导向训练，参与省市机器人竞赛</li><li>团队协作与项目管理能力培养</li></ul><h3>适合年龄</h3><p>8-14 岁（小学 2 年级 - 初中）</p>',
  NULL, NULL, 1, NULL, 0, 1, 3, 'ch_depart_05', '2', '1'),
 ('ch_course_04', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Scratch 游戏设计专项',
  '<h2>Scratch 游戏设计专项课程</h2><p>在掌握 Scratch 基础操作后，深入学习游戏设计原理：碰撞检测、计分系统、多关卡设计、音效与动画。学员将独立完成 3 个完整游戏作品。</p><h3>课程亮点</h3><ul><li>游戏化学习，激发兴趣</li><li>涵盖游戏设计核心概念</li><li>作品可在 Scratch 社区分享</li></ul><h3>适合年龄</h3><p>7-10 岁（已完成 Scratch 启蒙）</p>',
  NULL, NULL, 1, NULL, 0, 1, 4, 'ch_depart_03', '1', '1'),
 ('ch_course_05', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Python 数据可视化入门',
  '<h2>Python 数据可视化入门课程</h2><p>学习使用 Matplotlib 和 Pandas 进行数据处理与可视化，制作漂亮的图表和报告。</p><h3>课程亮点</h3><ul><li>真实数据集分析（天气、运动、社交媒体）</li><li>培养数据思维与审美</li><li>为后续人工智能学习打基础</li></ul><h3>适合年龄</h3><p>11-14 岁（已掌握 Python 基础）</p>',
  NULL, NULL, 1, NULL, 0, 1, 5, 'ch_depart_04', '2', '2'),
 ('ch_course_06', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  '信息学奥赛 CSP-J 入门',
  '<h2>信息学奥赛 CSP-J 入门课程</h2><p>面向准备参加 CSP-J（入门级）认证的学员，系统学习 C++ 语法、基础算法与数据结构。</p><h3>课程亮点</h3><ul><li>竞赛导向，系统训练</li><li>涵盖历年真题解析</li><li>一对一作业批改与竞赛策略指导</li></ul><h3>适合年龄</h3><p>12-14 岁（已具备 Python 基础，准备学习 C++）</p>',
  NULL, NULL, 1, NULL, 0, 0, 6, 'ch_depart_04', '2', '2');

-- ============================================================
-- 8. 课程单元（teaching_course_unit）——每门课 4-6 个单元
-- ============================================================
DELETE FROM teaching_course_unit;
INSERT INTO teaching_course_unit (id, create_by, create_time, sys_org_code, del_flag, unit_name, unit_intro, course_id, course_video_source, show_course_video, course_case, show_course_case, course_ppt, show_course_ppt, course_work_type, course_work, course_work_answer, course_plan, show_course_plan, map_x, map_y, media_content, order_num) VALUES
 -- Scratch创意编程启蒙
 ('ch_unit_01', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1课 认识Scratch', '了解Scratch界面布局，学会创建第一个动画项目：让小猫说"你好！"', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '<p>本节课学习目标：熟悉Scratch编辑器界面，理解角色、舞台、脚本三个核心概念。</p>', 1),
 ('ch_unit_02', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2课 角色与运动', '学习运动积木块，控制角色移动、旋转和弹性', 'ch_course_01', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 100, 200, '', 2),
 ('ch_unit_03', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3课 外观与声音', '给角色添加造型切换和音效，制作会说话的角色', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 3),
 ('ch_unit_04', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4课 事件与控制', '学习事件触发和条件判断，制作互动故事', 'ch_course_01', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 200, 200, '', 4),
 ('ch_unit_05', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第5课 循环与变量', '使用重复执行和变量，制作计分器和倒计时', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 5),
 -- Python代码编程进阶
 ('ch_unit_10', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1课 Hello Python', '安装Python环境，编写第一个print程序', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 1),
 ('ch_unit_11', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2课 变量与数据类型', '学习变量、整数、浮点数、字符串和布尔值', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 2),
 ('ch_unit_12', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3课 条件判断', '学习if-else语句，编写猜数字游戏', 'ch_course_02', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 200, 200, '', 3),
 ('ch_unit_13', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4课 循环结构', '学习for和while循环，绘制Turtle图形', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 4),
 ('ch_unit_14', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第5课 列表与字典', '学习Python数据结构，制作通讯录管理程序', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 5),
 -- 机器人搭建与竞赛
 ('ch_unit_20', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1课 认识机器人', '了解机器人组成结构，认识主控板和传感器', 'ch_course_03', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 1),
 ('ch_unit_21', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2课 LED与蜂鸣器', '学习点亮LED和驱动蜂鸣器发声', 'ch_course_03', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 2),
 ('ch_unit_22', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3课 电机与轮子', '学习驱动电机，让机器人动起来', 'ch_course_03', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 0, 0, '', 3),
 ('ch_unit_23', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4课 循线行走', '学习红外传感器，让机器人沿黑线行走', 'ch_course_03', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 4),
 -- Scratch游戏设计专项
 ('ch_unit_30', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1课 接苹果游戏', '学习碰撞检测，制作经典接苹果游戏', 'ch_course_04', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 1),
 ('ch_unit_31', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2课 打地鼠游戏', '学习随机出现和计分系统，制作打地鼠', 'ch_course_04', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 2),
 ('ch_unit_32', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3课 迷宫探险', '学习方向键控制和墙壁碰撞，制作迷宫游戏', 'ch_course_04', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 3),
 ('ch_unit_33', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4课 多关卡设计', '学习关卡切换和难度递增，制作多关卡游戏', 'ch_course_04', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 4),
 -- Python数据可视化
 ('ch_unit_40', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1课 初识Matplotlib', '安装Matplotlib，绘制第一条折线图', 'ch_course_05', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 1),
 ('ch_unit_41', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2课 柱状图与饼图', '学习柱状图、饼图和散点图的绘制', 'ch_course_05', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 2),
 -- 信息学奥赛CSP-J
 ('ch_unit_60', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1课 C++ 起步', '搭建 C++ 编译环境，理解程序编译运行流程', 'ch_course_06', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 1),
 ('ch_unit_61', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2课 顺序与分支', 'C++ 基本语句、if 分支与 CSP-J 真题入门', 'ch_course_06', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 0, 0, '', 2);

-- ============================================================
-- 9. 课程-班级授权（teaching_course_dept）
-- ============================================================
DELETE FROM teaching_course_dept;
INSERT INTO teaching_course_dept (id, create_by, create_time, sys_org_code, dept_id, course_id, open_time) VALUES
 ('ch_cd_01', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_03', 'ch_course_01', '2026-09-10 08:00:00'),
 ('ch_cd_02', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_04', 'ch_course_02', '2026-09-10 08:00:00'),
 ('ch_cd_03', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_05', 'ch_course_03', '2026-09-10 08:00:00'),
 ('ch_cd_04', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_03', 'ch_course_04', '2026-10-01 08:00:00'),
 ('ch_cd_05', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_04', 'ch_course_05', '2026-10-01 08:00:00'),
 ('ch_cd_06', 'admin', '2026-09-07 00:00:00', 'A01', 'ch_depart_04', 'ch_course_06', '2026-11-01 08:00:00');

-- ============================================================
-- 10. 公告（sys_announcement）
-- ============================================================
DELETE FROM sys_announcement;
INSERT INTO sys_announcement (id, titile, msg_content, start_time, end_time, sender, priority, msg_category, msg_type, send_status, send_time, cancel_time, del_flag, bus_type, bus_id, open_type, open_page, create_by, create_time, msg_abstract) VALUES
 ('ch_ann_01', '春晖青少年编程中心 2026 秋季学期开学通知',
  '<p>各位家长、同学：</p><p>春晖青少年编程中心 2026 年秋季学期将于 9 月 10 日正式开课，请按时到校。</p><p>上课地址：武胜县沿口镇上东街257号（邮编638400）</p><p>—— 武胜县沿口春夏秋信息技术中心（个体工商户）</p>',
  '2026-09-01 00:00:00', '2027-01-31 23:59:59', 'admin', 'H', '1', 'ALL', '1', '2026-09-01 08:00:00', NULL, '0', NULL, NULL, NULL, NULL, 'admin', '2026-09-07 00:00:00', '秋季学期开学通知'),
 ('ch_ann_02', 'Scratch 创意编程启蒙班课程安排',
  '<p>Scratch 启蒙班每周六上午 9:00-11:00 上课，请提前 10 分钟到教室（2楼 Scratch教室）。</p>',
  '2026-09-01 00:00:00', '2027-01-31 23:59:59', 'teacher1', 'M', '1', 'ALL', '1', '2026-09-05 08:00:00', NULL, '0', NULL, NULL, NULL, NULL, 'teacher1', '2026-09-07 00:00:00', 'Scratch班课程安排'),
 ('ch_ann_03', 'Python 进阶班新学期课程大纲发布',
  '<p>Python 进阶班新学期将学习条件判断、循环结构、列表与字典等内容，请同学们提前预习第一课内容。</p>',
  '2026-09-01 00:00:00', '2027-01-31 23:59:59', 'teacher2', 'M', '1', 'ALL', '1', '2026-09-05 08:00:00', NULL, '0', NULL, NULL, NULL, NULL, 'teacher2', '2026-09-07 00:00:00', 'Python班课程大纲'),
 ('ch_ann_04', '机器人竞赛班即将开启报名',
  '<p>机器人竞赛班将于 10 月开始报名，名额有限（8人/班），请有意向的同学尽快联系陈教练。</p>',
  '2026-09-01 00:00:00', '2026-10-31 23:59:59', 'teacher3', 'H', '1', 'ALL', '1', '2026-09-05 08:00:00', NULL, '0', NULL, NULL, NULL, NULL, 'teacher3', '2026-09-07 00:00:00', '机器人竞赛班报名');

-- ============================================================
-- 11. 用户通告阅读标记（sys_announcement_send）
-- ============================================================
DELETE FROM sys_announcement_send;
INSERT INTO sys_announcement_send (id, annt_id, user_id, read_flag, read_time) VALUES
 ('ch_as_01', 'ch_ann_01', 'ch_user_teacher1', '0', NULL),
 ('ch_as_02', 'ch_ann_01', 'ch_user_teacher2', '0', NULL),
 ('ch_as_03', 'ch_ann_01', 'ch_user_teacher3', '0', NULL),
 ('ch_as_04', 'ch_ann_01', 'ch_user_stu1',     '0', NULL),
 ('ch_as_05', 'ch_ann_01', 'ch_user_stu2',     '0', NULL),
 ('ch_as_06', 'ch_ann_01', 'ch_user_stu3',     '0', NULL),
 ('ch_as_07', 'ch_ann_01', 'ch_user_stu4',     '0', NULL),
 ('ch_as_08', 'ch_ann_01', 'ch_user_stu5',     '0', NULL),
 ('ch_as_09', 'ch_ann_01', 'ch_user_stu6',     '0', NULL),
 ('ch_as_10', 'ch_ann_02', 'ch_user_stu1',     '0', NULL),
 ('ch_as_11', 'ch_ann_02', 'ch_user_stu2',     '0', NULL),
 ('ch_as_12', 'ch_ann_04', 'ch_user_stu5',     '0', NULL),
 ('ch_as_13', 'ch_ann_04', 'ch_user_stu6',     '0', NULL);
