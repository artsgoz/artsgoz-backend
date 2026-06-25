-- Seed data: 10 dummy documents for local development
-- Run: docker cp scripts/seed_documents.sql artsgoz-postgres-local:/tmp/seed_documents.sql && docker exec artsgoz-postgres-local psql -U artsgoz -d artsgoz -f /tmp/seed_documents.sql

INSERT INTO documents (name, details, category, status, file_url, file_name, file_size, is_active)
VALUES
(
  'ฟอร์มขอเอกสารขอความอนุเคราะห์ฝึกงาน.pdf',
  'ใช้ยื่นสำหรับฝึกงานภาคฤดูร้อนประจำปีการศึกษา 2568 กำหนดส่งภายในเดือนมีนาคม',
  'เกี่ยวกับฝึกงาน',
  'pending',
  'https://cdn.artsgoz.chula.ac.th/documents/internship-request-2026.pdf',
  'internship-request-2026.pdf',
  245760,
  true
),
(
  'ใบสมัครขอรับทุนการศึกษาเรียนดีคณะอักษรศาสตร์.pdf',
  'ทุนสนับสนุนสำหรับนิสิตเรียนดีแต่ขาดแคลนทุนทรัพย์ ประจำเทอม 1/2569',
  'ทุนการศึกษา',
  'neutral',
  'https://cdn.artsgoz.chula.ac.th/documents/scholarship-academic-excellence.pdf',
  'scholarship-academic-excellence.pdf',
  512000,
  true
),
(
  'คำร้องทั่วไป กอศ. 01.pdf',
  'ใช้ยื่นสำหรับแจ้งคำร้องทั่วไปและติดต่อฝ่ายทะเบียนคณะ',
  'ฟอร์มต่าง ๆ ของกอศ.',
  'none',
  'https://cdn.artsgoz.chula.ac.th/documents/general-request-form-01.pdf',
  'general-request-form-01.pdf',
  128500,
  true
),
(
  'คำร้องขอลงทะเบียนเรียนรายวิชานอกคณะ.pdf',
  'ใช้กรณีต้องการลงทะเบียนเรียนวิชาเลือกเสรีนอกคณะอักษรศาสตร์',
  'เกี่ยวกับวิชาเรียน',
  'none',
  'https://cdn.artsgoz.chula.ac.th/documents/cross-faculty-registration.pdf',
  'cross-faculty-registration.pdf',
  358000,
  true
),
(
  'คู่มือนิสิตใหม่คณะอักษรศาสตร์ ปี 2568.pdf',
  'คู่มือรวบรวมข้อมูลการปฏิบัติตัว กฎระเบียบ และการใช้ชีวิตในรั้วมหาวิทยาลัย',
  'อื่น ๆ',
  'none',
  'https://cdn.artsgoz.chula.ac.th/documents/freshman-handbook-2025.pdf',
  'freshman-handbook-2025.pdf',
  5242880,
  true
),
(
  'แบบฟอร์มเบิกจ่ายค่าเดินทางจัดกิจกรรมชมรม.pdf',
  'ฟอร์มสำหรับคณะกรรมการชมรมส่งเบิกจ่ายงบประมาณโครงการพิเศษ',
  'ฟอร์มต่าง ๆ ของกอศ.',
  'none',
  'https://cdn.artsgoz.chula.ac.th/documents/club-expense-travel.pdf',
  'club-expense-travel.pdf',
  198000,
  true
),
(
  'แบบประเมินผลการฝึกงาน (สำหรับองค์กร).pdf',
  'ฟอร์มสำหรับให้พี่เลี้ยงหรือหัวหน้างานประเมินคะแนนนิสิตหลังฝึกงานเสร็จสิ้น',
  'เกี่ยวกับฝึกงาน',
  'none',
  'https://cdn.artsgoz.chula.ac.th/documents/internship-evaluation-company.pdf',
  'internship-evaluation-company.pdf',
  405000,
  true
),
(
  'ใบสมัครทุนวิจัยระดับนิสิตปริญญาตรี.pdf',
  'ทุนสำหรับนิสิตที่ร่วมทำโครงงานวิจัยกับอาจารย์ที่ปรึกษา',
  'ทุนการศึกษา',
  'none',
  'https://cdn.artsgoz.chula.ac.th/documents/research-grant-undergrad.pdf',
  'research-grant-undergrad.pdf',
  890000,
  true
),
(
  'คำร้องขอเปิดวิชาเรียนกรณีพิเศษ (กลุ่มย่อย).pdf',
  'คำร้องขอเปิดเซคชั่นใหม่กรณีนักศึกษาไม่พอหรือไม่สามารถลงทะเบียนตามปกติได้',
  'เกี่ยวกับวิชาเรียน',
  'failed',
  'https://cdn.artsgoz.chula.ac.th/documents/special-class-section-request.pdf',
  'special-class-section-request.pdf',
  310000,
  true
),
(
  'แบบฟอร์มยินยอมเปิดเผยข้อมูลส่วนบุคคล (PDPA).pdf',
  'เอกสารให้ความยินยอมสำหรับกิจกรรมต่างๆ ของคณะและงานทะเบียน',
  'อื่น ๆ',
  'none',
  'https://cdn.artsgoz.chula.ac.th/documents/pdpa-consent-form-general.pdf',
  'pdpa-consent-form-general.pdf',
  156000,
  true
);
