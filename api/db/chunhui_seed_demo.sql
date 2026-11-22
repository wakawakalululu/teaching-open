-- ============================================================
-- 春晖青少年编程中心 种子 · DEMO 层（业务示例数据）
-- 产品归属：武胜县沿口春夏秋信息技术中心（个体工商户）
-- 统一社会信用代码：92511622MAKLLHTT5D
-- 经营场所：四川省广安市武胜县沿口镇上东街257号（邮编638400）
--
-- 前置条件：先执行 chunhui_seed_base.sql（demo 引用 base 的部门/用户/课程授权）。
-- 覆盖内容：课程 6 门 / 单元 22 个 / 课程授权 6 条 / 公告 4 条 + 阅读标记 /
--           资讯 4 条 / 文件登记 24 条 / 学员作品 12 件 / 互评 6 条。
-- 注意：封面 PNG、闯关地图与 .sb3 为 chseed 生成的二进制产物，纯 SQL 部署时
--       数据行完整、静态预览 404（需运行 chseed -layer demo 生成物理文件）。
--
-- 等效工具：cd goframe-server && go run ./tools/chseed -layer demo
-- ============================================================

-- ============================================================
-- 7. 课程（teaching_course）——春晖课程体系（6门）
-- ============================================================
DELETE FROM teaching_course;
INSERT INTO teaching_course (id, create_by, create_time, sys_org_code, del_flag, course_name, course_desc, course_icon, course_cover, show_type, course_map, is_shared, show_home, order_num, depart_ids, course_type, course_category) VALUES
 ('ch_course_01', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Scratch 创意编程启蒙',
  '<h2>Scratch 图形化编程启蒙课程</h2><p>适合 6-9 岁零基础学员。通过 MIT 开发的 Scratch 图形化编程工具，孩子可以像搭积木一样拖拽代码块，创作属于自己的互动故事、动画和小游戏。</p><h3>课程亮点</h3><ul><li>零门槛入门，无需打字基础</li><li>项目式学习，每节课完成一个作品</li><li>培养逻辑思维、创造力与问题解决能力</li><li>小班教学（8人/班），确保每个孩子得到关注</li></ul><h3>适合年龄</h3><p>6-9 岁（小学 1-3 年级）</p>',
  NULL, NULL, 1, 'ch_file_map01', 0, 1, 1, 'ch_depart_03', '1', '1'),
 ('ch_course_02', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Python 代码编程进阶',
  '<h2>Python 代码编程进阶课程</h2><p>适合 10-14 岁有一定编程基础的学员。从图形化编程过渡到真正的代码编程，学习 Python 语法、数据结构、算法与项目开发。</p><h3>课程亮点</h3><ul><li>从积木到代码的平滑过渡</li><li>涵盖 Turtle绘图、Pygame游戏开发、数据分析入门</li><li>培养算法思维与工程能力</li><li>完成多个实战项目</li></ul><h3>适合年龄</h3><p>10-14 岁（小学 4 年级 - 初中）</p>',
  NULL, NULL, 1, 'ch_file_map02', 0, 1, 2, 'ch_depart_04', '2', '2'),
 ('ch_course_03', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  '机器人搭建与竞赛',
  '<h2>机器人搭建与竞赛课程</h2><p>适合 8-14 岁学员。学习机器人机械结构搭建、传感器应用与编程控制，备战各类机器人竞赛。</p><h3>课程亮点</h3><ul><li>硬件+软件结合，动手与动脑并行</li><li>竞赛导向训练，参与省市机器人竞赛</li><li>团队协作与项目管理能力培养</li></ul><h3>适合年龄</h3><p>8-14 岁（小学 2 年级 - 初中）</p>',
  NULL, NULL, 1, 'ch_file_map03', 0, 1, 3, 'ch_depart_05', '2', '1'),
 ('ch_course_04', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Scratch 游戏设计专项',
  '<h2>Scratch 游戏设计专项课程</h2><p>在掌握 Scratch 基础操作后，深入学习游戏设计原理：碰撞检测、计分系统、多关卡设计、音效与动画。学员将独立完成 3 个完整游戏作品。</p><h3>课程亮点</h3><ul><li>游戏化学习，激发兴趣</li><li>涵盖游戏设计核心概念</li><li>作品可在 Scratch 社区分享</li></ul><h3>适合年龄</h3><p>7-10 岁（已完成 Scratch 启蒙）</p>',
  NULL, NULL, 1, 'ch_file_map04', 0, 1, 4, 'ch_depart_03', '1', '1'),
 ('ch_course_05', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  'Python 数据可视化入门',
  '<h2>Python 数据可视化入门课程</h2><p>学习使用 Matplotlib 和 Pandas 进行数据处理与可视化，制作漂亮的图表和报告。</p><h3>课程亮点</h3><ul><li>真实数据集分析（天气、运动、社交媒体）</li><li>培养数据思维与审美</li><li>为后续人工智能学习打基础</li></ul><h3>适合年龄</h3><p>11-14 岁（已掌握 Python 基础）</p>',
  NULL, NULL, 1, 'ch_file_map05', 0, 1, 5, 'ch_depart_04', '2', '2'),
 ('ch_course_06', 'admin', '2026-09-07 00:00:00', 'A01', 0,
  '信息学奥赛 CSP-J 入门',
  '<h2>信息学奥赛 CSP-J 入门课程</h2><p>面向准备参加 CSP-J（入门级）认证的学员，系统学习 C++ 语法、基础算法与数据结构。</p><h3>课程亮点</h3><ul><li>竞赛导向，系统训练</li><li>涵盖历年真题解析</li><li>一对一作业批改与竞赛策略指导</li></ul><h3>适合年龄</h3><p>12-14 岁（已具备 Python 基础，准备学习 C++）</p>',
  NULL, NULL, 1, 'ch_file_map06', 0, 0, 6, 'ch_depart_04', '2', '2');

-- ============================================================
-- 8. 课程单元（teaching_course_unit）——每门课 4-6 个单元
-- ============================================================

--    map_x/map_y 对应 course_map 背景图上的关卡节点像素位
-- ============================================================
DELETE FROM teaching_course_unit;
INSERT INTO teaching_course_unit (id, create_by, create_time, sys_org_code, del_flag, unit_name, unit_intro, course_id, course_video_source, show_course_video, course_case, show_course_case, course_ppt, show_course_ppt, course_work_type, course_work, course_work_answer, course_plan, show_course_plan, map_x, map_y, media_content, order_num) VALUES
 -- Scratch创意编程启蒙（关卡路径）
 ('ch_unit_01', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1关 初识Scratch', '了解Scratch界面布局，创建第一个动画项目：让小猫说“你好！”', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 100, 150, '<p>关卡目标：熟悉Scratch编辑器界面，理解角色、舞台、脚本三个核心概念。</p>', 1),
 ('ch_unit_02', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2关 角色与运动', '学习运动积木块，控制角色移动、旋转和弹性', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 310, 130, '', 2),
 ('ch_unit_03', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3关 外观与声音', '给角色添加造型切换和音效，制作会说话的角色', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 520, 160, '', 3),
 ('ch_unit_04', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4关 事件与控制', '学习事件触发和条件判断，制作互动故事', 'ch_course_01', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 730, 230, '', 4),
 ('ch_unit_05', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第5关 循环与变量【Boss】', '关卡挑战：运用重复执行和变量制作计分器作品，提交作品即通关！', 'ch_course_01', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 560, 390, '', 5),
 -- Python代码编程进阶
 ('ch_unit_10', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1关 Hello Python', '安装Python环境，编写第一个print程序', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 800, 150, '', 1),
 ('ch_unit_11', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2关 变量与类型', '学习变量、整数、浮点数、字符串和布尔值', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 590, 130, '', 2),
 ('ch_unit_12', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3关 条件判断', '学习if-else语句，编写猜数字小游戏', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 380, 160, '', 3),
 ('ch_unit_13', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4关 循环结构', '学习for和while循环，绘制Turtle图形', 'ch_course_02', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 170, 230, '', 4),
 ('ch_unit_14', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第5关 列表与字典【Boss】', '关卡挑战：用列表与字典完成通讯录程序，提交作品即通关！', 'ch_course_02', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 340, 390, '', 5),
 -- 机器人搭建与竞赛
 ('ch_unit_20', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1关 认识机器人', '了解机器人组成结构，认识主控板和传感器', 'ch_course_03', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 110, 160, '', 1),
 ('ch_unit_21', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2关 LED与蜂鸣器', '学习点亮LED和驱动蜂鸣器发声', 'ch_course_03', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 340, 130, '', 2),
 ('ch_unit_22', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3关 电机与轮子', '学习驱动电机，让机器人动起来', 'ch_course_03', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 570, 170, '', 3),
 ('ch_unit_23', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4关 循线行走【Boss】', '关卡挑战：调试红外传感器完成循线赛道，提交走通全程的作品视频即通关！', 'ch_course_03', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 790, 300, '', 4),
 -- Scratch游戏设计专项
 ('ch_unit_30', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1关 接苹果', '学习碰撞检测，制作经典接苹果游戏', 'ch_course_04', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 110, 150, '', 1),
 ('ch_unit_31', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2关 打地鼠', '学习随机出现和计分系统，制作打地鼠', 'ch_course_04', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 340, 130, '', 2),
 ('ch_unit_32', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第3关 迷宫探险', '学习方向键控制和墙壁碰撞，制作迷宫游戏', 'ch_course_04', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 570, 160, '', 3),
 ('ch_unit_33', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第4关 多关卡设计【Boss】', '关卡挑战：设计一款含 3 个关卡的完整游戏并提交，即通关毕业！', 'ch_course_04', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 790, 280, '', 4),
 -- Python数据可视化
 ('ch_unit_40', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1关 初识Matplotlib', '安装Matplotlib，绘制第一条折线图', 'ch_course_05', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 160, 160, '', 1),
 ('ch_unit_41', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2关 柱状图与饼图【Boss】', '关卡挑战：用真实数据集绘制一张柱状图与一张饼图，提交即通关！', 'ch_course_05', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 460, 130, '', 2),
 -- 信息学奥赛CSP-J
 ('ch_unit_60', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第1关 C++ 起步', '搭建 C++ 编译环境，理解程序编译运行流程', 'ch_course_06', 1, 1, '', 1, '', 0, NULL, '', '', '', 0, 740, 160, '', 1),
 ('ch_unit_61', 'admin', '2026-09-07 00:00:00', 'A01', 0, '第2关 顺序与分支【Boss】', '关卡挑战：完成一套 CSP-J 入门真题的顺序与分支题目，提交即通关！', 'ch_course_06', 1, 1, '', 1, '', 0, 1, '', '', '', 0, 440, 130, '', 2);

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


-- ============================================================
-- 12. 资讯（teaching_news）——门户 newsList（news_status>=1 可见）
-- ============================================================
DELETE FROM teaching_news;
INSERT INTO teaching_news (id, news_title, news_content, news_status, create_by, create_time) VALUES
 ('ch_news_01', '春晖青少年编程中心 2026 秋季学期顺利开课',
  '<p>9 月 10 日，春晖青少年编程中心 2026 年秋季学期正式开课。本学期开设 Scratch 创意编程启蒙、Python 代码编程进阶、机器人搭建与竞赛等 6 门课程。</p><p>经营场所：武胜县沿口镇上东街257号（邮编638400）。</p>', 1, 'admin', '2026-09-10 09:00:00'),
 ('ch_news_02', '我中心学员在广安市青少年编程展演活动中获得佳绩',
  '<p>在广安市青少年编程展演活动中，我中心 Scratch 启蒙班与 Python 进阶班共 6 名学员晋级展示环节，其中 2 名学员获得一等奖。</p>', 1, 'teacher1', '2026-10-18 14:00:00'),
 ('ch_news_03', 'Scratch 创意编程公益公开课圆满结束',
  '<p>10 月下旬，中心举办公益公开课两场，共 30 组家庭参与，孩子们完成了自己的第一个动画作品。</p>', 1, 'teacher1', '2026-10-30 16:00:00'),
 ('ch_news_04', '寒假 Python 专项班与 CSP-J 冲刺班报名开始',
  '<p>寒假期间将开设 Python 数据可视化专项班与信息学奥赛 CSP-J 冲刺班，每班限额 8 人，报满即止。</p><p>地址：武胜县沿口镇上东街257号（邮编638400）。</p>', 1, 'teacher2', '2026-11-15 10:00:00');

-- ============================================================
-- 13. 作品文件登记（sys_file）+ 闯关地图背景 + 学员作品 + 互评
-- 注意：封面 PNG、地图背景与 .sb3 占位文件为 chseed 工具生成的
-- 二进制产物（写入 STORAGE_LOCAL_ROOT），纯 SQL 部署时数据行完整、
-- 静态预览 404（需运行 chseed 生成物理文件）。
-- ============================================================
DELETE FROM sys_file WHERE id LIKE 'ch_file_%' OR id LIKE 'ch_work_%_file';
INSERT INTO sys_file (id, file_name, file_path, file_type, file_location, file_tag, sys_org_code, create_by, create_time, del_flag) VALUES
 ('ch_file_cv01', 'cover.png', 'chunhui/covers/ch_file_cv01.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_cv02', 'cover.png', 'chunhui/covers/ch_file_cv02.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_cv03', 'cover.png', 'chunhui/covers/ch_file_cv03.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_cv04', 'cover.png', 'chunhui/covers/ch_file_cv04.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_cv05', 'cover.png', 'chunhui/covers/ch_file_cv05.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_cv06', 'cover.png', 'chunhui/covers/ch_file_cv06.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_map01', 'course-map.png', 'chunhui/maps/ch_course_01.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_map02', 'course-map.png', 'chunhui/maps/ch_course_02.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_map03', 'course-map.png', 'chunhui/maps/ch_course_03.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_map04', 'course-map.png', 'chunhui/maps/ch_course_04.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_map05', 'course-map.png', 'chunhui/maps/ch_course_05.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_file_map06', 'course-map.png', 'chunhui/maps/ch_course_06.png', 0, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_01_file', '我的第一个动画.sb3', 'chunhui/works/ch_work_01.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_02_file', '小猫追球.sb3', 'chunhui/works/ch_work_02.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_03_file', '会说话的角色.sb3', 'chunhui/works/ch_work_03.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_04_file', 'Hello Python.sb3', 'chunhui/works/ch_work_04.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_05_file', '猜数字游戏.sb3', 'chunhui/works/ch_work_05.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_06_file', 'Turtle 画星星.sb3', 'chunhui/works/ch_work_06.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_07_file', '循线机器人.sb3', 'chunhui/works/ch_work_07.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_08_file', 'LED 流水灯.sb3', 'chunhui/works/ch_work_08.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_09_file', '避障小车.sb3', 'chunhui/works/ch_work_09.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_10_file', '接苹果.sb3', 'chunhui/works/ch_work_10.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_11_file', '打地鼠.sb3', 'chunhui/works/ch_work_11.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0),
 ('ch_work_12_file', '迷宫探险.sb3', 'chunhui/works/ch_work_12.sb3', 1, 1, 'chunhui', 'A01', 'admin', '2026-09-07 00:00:00', 0);

DELETE FROM teaching_work;
INSERT INTO teaching_work (id, create_by, create_time, sys_org_code, user_id, depart_id, course_id, work_name, work_type, work_file, work_cover, work_status, star_num, collect_num, view_num, work_scene, has_cloud_data, del_flag) VALUES
 ('ch_work_01', 'ch_user_stu1', '2026-09-20 10:00:00', 'A01', 'ch_user_stu1', 'ch_depart_03', 'ch_course_01', '我的第一个动画：小猫说你好', 1, 'ch_work_01_file', 'ch_file_cv01', 4, 12, 4, 58, '{"v":1}', 0, 0),
 ('ch_work_02', 'ch_user_stu2', '2026-09-20 10:00:00', 'A01', 'ch_user_stu2', 'ch_depart_03', 'ch_course_01', '小猫追球', 1, 'ch_work_02_file', 'ch_file_cv01', 4, 9, 3, 41, '{"v":1}', 0, 0),
 ('ch_work_03', 'ch_user_stu1', '2026-09-20 10:00:00', 'A01', 'ch_user_stu1', 'ch_depart_03', 'ch_course_01', '会说话的角色', 1, 'ch_work_03_file', 'ch_file_cv01', 3, 3, 1, 17, '{"v":1}', 0, 0),
 ('ch_work_04', 'ch_user_stu3', '2026-09-20 10:00:00', 'A01', 'ch_user_stu3', 'ch_depart_04', 'ch_course_02', 'Hello Python', 1, 'ch_work_04_file', 'ch_file_cv02', 4, 11, 3, 66, '{"v":1}', 0, 0),
 ('ch_work_05', 'ch_user_stu4', '2026-09-20 10:00:00', 'A01', 'ch_user_stu4', 'ch_depart_04', 'ch_course_02', '猜数字游戏', 1, 'ch_work_05_file', 'ch_file_cv02', 4, 15, 5, 73, '{"v":1}', 0, 0),
 ('ch_work_06', 'ch_user_stu3', '2026-09-20 10:00:00', 'A01', 'ch_user_stu3', 'ch_depart_04', 'ch_course_02', 'Turtle 画星星', 1, 'ch_work_06_file', 'ch_file_cv02', 3, 5, 1, 22, '{"v":1}', 0, 0),
 ('ch_work_07', 'ch_user_stu5', '2026-09-20 10:00:00', 'A01', 'ch_user_stu5', 'ch_depart_05', 'ch_course_03', '循线机器人', 1, 'ch_work_07_file', 'ch_file_cv03', 4, 8, 2, 39, '{"v":1}', 0, 0),
 ('ch_work_08', 'ch_user_stu6', '2026-09-20 10:00:00', 'A01', 'ch_user_stu6', 'ch_depart_05', 'ch_course_03', 'LED 流水灯', 1, 'ch_work_08_file', 'ch_file_cv03', 4, 6, 2, 28, '{"v":1}', 0, 0),
 ('ch_work_09', 'ch_user_stu5', '2026-09-20 10:00:00', 'A01', 'ch_user_stu5', 'ch_depart_05', 'ch_course_03', '避障小车', 1, 'ch_work_09_file', 'ch_file_cv03', 3, 2, 0, 13, '{"v":1}', 0, 0),
 ('ch_work_10', 'ch_user_stu1', '2026-09-20 10:00:00', 'A01', 'ch_user_stu1', 'ch_depart_03', 'ch_course_04', '接苹果', 1, 'ch_work_10_file', 'ch_file_cv04', 4, 13, 4, 61, '{"v":1}', 0, 0),
 ('ch_work_11', 'ch_user_stu2', '2026-09-20 10:00:00', 'A01', 'ch_user_stu2', 'ch_depart_03', 'ch_course_04', '打地鼠', 1, 'ch_work_11_file', 'ch_file_cv04', 4, 10, 3, 47, '{"v":1}', 0, 0),
 ('ch_work_12', 'ch_user_stu4', '2026-09-20 10:00:00', 'A01', 'ch_user_stu4', 'ch_depart_04', 'ch_course_04', '迷宫探险', 1, 'ch_work_12_file', 'ch_file_cv04', 4, 4, 1, 25, '{"v":1}', 0, 0);

DELETE FROM teaching_work_comment;
INSERT INTO teaching_work_comment (id, work_id, user_id, comment, create_by, create_time, sys_org_code) VALUES
 ('ch_wc_01', 'ch_work_01', 'ch_user_stu2', '小猫的造型切换好流畅，我也要试试！', 'ch_user_stu2', '2026-09-21 09:00:00', 'A01'),
 ('ch_wc_02', 'ch_work_01', 'ch_user_teacher1', '第一次就能完成完整动画，继续保持！', 'ch_user_teacher1', '2026-09-21 10:00:00', 'A01'),
 ('ch_wc_03', 'ch_work_05', 'ch_user_stu3', '猜数字的循环写得真棒，用了二分吗？', 'ch_user_stu3', '2026-09-25 15:00:00', 'A01'),
 ('ch_wc_04', 'ch_work_07', 'ch_user_stu6', '循线机器人转弯很稳，速度也快！', 'ch_user_stu6', '2026-09-28 11:00:00', 'A01'),
 ('ch_wc_05', 'ch_work_10', 'ch_user_stu5', '苹果接住时有音效，细节满分。', 'ch_user_stu5', '2026-10-09 16:00:00', 'A01'),
 ('ch_wc_06', 'ch_work_04', 'ch_user_teacher2', '代码缩进规范，注释也很清楚。', 'ch_user_teacher2', '2026-10-11 09:30:00', 'A01');

