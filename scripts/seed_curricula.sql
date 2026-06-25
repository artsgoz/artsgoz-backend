-- Seed data: 1 curriculum and standard subjects (26 items to cover all categories)
-- Run: docker cp scripts/seed_curricula.sql artsgoz-postgres-local:/tmp/seed_curricula.sql && docker exec artsgoz-postgres-local psql -U artsgoz -d artsgoz -f /tmp/seed_curricula.sql

-- Insert Curriculum (2025 version)
INSERT INTO curricula (id, name, year, total_credits, is_active)
VALUES ('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', 'หลักสูตรอักษรศาสตร์บัณฑิต (ปรับปรุง 2565)', 2565, 129, true)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, year = EXCLUDED.year, total_credits = EXCLUDED.total_credits, is_active = EXCLUDED.is_active;

-- Insert Curriculum categories configs
INSERT INTO curriculum_categories (curriculum_id, category, required_credits, groups)
VALUES
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', 'หมวดวิชาพื้นฐานอักษร', 27, ARRAY['ภาษาไทย', 'ภาษาอังกฤษ']),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', 'หมวดการศึกษาทั่วไป', 30, ARRAY[]::TEXT[]),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', 'หมวดวิชาเลือกเสรี', 6, ARRAY[]::TEXT[]),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', 'หมวดวิชาเอก', 48, ARRAY[]::TEXT[]),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', 'หมวดวิชาโท', 18, ARRAY[]::TEXT[])
ON CONFLICT (curriculum_id, category) DO UPDATE SET required_credits = EXCLUDED.required_credits, groups = EXCLUDED.groups;

-- Insert Subjects
DELETE FROM subjects WHERE curriculum_id = '9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d';

INSERT INTO subjects (curriculum_id, code, name_th, name_en, credits, category, semester, grp, is_custom)
VALUES
-- 1. หมวดวิชาพื้นฐานอักษร (9 วิชา = 27 หน่วยกิต)
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2200111', 'การใช้ภาษาไทย', 'Use of the Thai Language', 3.0, 'หมวดวิชาพื้นฐานอักษร', 1, 'ภาษาไทย', false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201111', 'ภาษาอังกฤษเพื่อการสื่อสาร 1', 'English for Communication I', 3.0, 'หมวดวิชาพื้นฐานอักษร', 1, 'ภาษาอังกฤษ', false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201112', 'ภาษาอังกฤษเพื่อการสื่อสาร 2', 'English for Communication II', 3.0, 'หมวดวิชาพื้นฐานอักษร', 2, 'ภาษาอังกฤษ', false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2203111', 'ประวัติศาสตร์ไทย', 'Thai History', 3.0, 'หมวดวิชาพื้นฐานอักษร', 2, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2204111', 'ปรัชญาเบื้องต้น', 'Introduction to Philosophy', 3.0, 'หมวดวิชาพื้นฐานอักษร', 3, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2208111', 'ภูมิศาสตร์กายภาพ', 'Physical Geography', 3.0, 'หมวดวิชาพื้นฐานอักษร', 3, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201211', 'การอ่านภาษาอังกฤษขั้นสูง', 'Advanced English Reading', 3.0, 'หมวดวิชาพื้นฐานอักษร', 4, 'ภาษาอังกฤษ', false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2200211', 'วากยสัมพันธ์ภาษาไทย', 'Thai Syntax', 3.0, 'หมวดวิชาพื้นฐานอักษร', 4, 'ภาษาไทย', false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2203211', 'ประวัติศาสตร์เอเชียตะวันออกเฉียงใต้', 'History of Southeast Asia', 3.0, 'หมวดวิชาพื้นฐานอักษร', 5, null, false),

-- 2. หมวดการศึกษาทั่วไป (5 วิชา = 15 หน่วยกิต)
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2301115', 'คอมพิวเตอร์และสารสนเทศ', 'Computer and Information', 3.0, 'หมวดการศึกษาทั่วไป', 1, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '5500111', 'การสื่อสารเพื่อความเข้าใจ', 'Communication for Understanding', 3.0, 'หมวดการศึกษาทั่วไป', 1, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '0201111', 'ศิลปะและการใช้ชีวิต', 'Art and Life', 3.0, 'หมวดการศึกษาทั่วไป', 2, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2400111', 'โลกในยุคปัจจุบัน', 'Contemporary World', 3.0, 'หมวดการศึกษาทั่วไป', 2, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2100111', 'สิ่งแวดล้อมและการพัฒนา', 'Environment and Development', 3.0, 'หมวดการศึกษาทั่วไป', 3, null, false),

-- 3. หมวดวิชาเอก (5 วิชา = 15 หน่วยกิต)
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201301', 'การแปลภาษาอังกฤษขั้นต้น', 'Introduction to English Translation', 3.0, 'หมวดวิชาเอก', 5, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201311', 'วรรณคดีอังกฤษเบื้องต้น', 'Introduction to English Literature', 3.0, 'หมวดวิชาเอก', 5, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201401', 'การแปลภาษาอังกฤษขั้นสูง', 'Advanced English Translation', 3.0, 'หมวดวิชาเอก', 6, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201411', 'วรรณคดีอเมริกันวิเคราะห์', 'American Literature Analysis', 3.0, 'หมวดวิชาเอก', 6, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2201451', 'การฟัง-พูดภาษาอังกฤษเชิงวิชาการ', 'Academic Listening and Speaking', 3.0, 'หมวดวิชาเอก', 7, null, false),

-- 4. หมวดวิชาโท (4 วิชา = 12 หน่วยกิต)
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2205111', 'ภาษาฝรั่งเศสเบื้องต้น 1', 'Basic French I', 3.0, 'หมวดวิชาโท', 3, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2205112', 'ภาษาฝรั่งเศสเบื้องต้น 2', 'Basic French II', 3.0, 'หมวดวิชาโท', 4, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2205211', 'ไวยากรณ์ภาษาฝรั่งเศสระดับกลาง', 'Intermediate French Grammar', 3.0, 'หมวดวิชาโท', 5, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2205311', 'วรรณคดีฝรั่งเศสเบื้องต้น', 'Introduction to French Literature', 3.0, 'หมวดวิชาโท', 6, null, false),

-- 5. หมวดวิชาเลือกเสรี (3 วิชา = 9 หน่วยกิต)
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2207111', 'การเขียนบทภาพยนตร์ขั้นต้น', 'Screenplay Writing I', 3.0, 'หมวดวิชาเลือกเสรี', 7, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2209111', 'ภาษาเยอรมันเบื้องต้น 1', 'Basic German I', 3.0, 'หมวดวิชาเลือกเสรี', 7, null, false),
('9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d', '2211111', 'จิตวิทยาทั่วไป', 'General Psychology', 3.0, 'หมวดวิชาเลือกเสรี', 8, null, false);
