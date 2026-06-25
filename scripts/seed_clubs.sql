-- Seed data: 10 dummy clubs for local development
-- Run: docker cp scripts/seed_clubs.sql artsgoz-postgres-local:/tmp/seed_clubs.sql && docker exec artsgoz-postgres-local psql -U artsgoz -d artsgoz -f /tmp/seed_clubs.sql

INSERT INTO clubs (name, category, description, instagram, image_url, is_active)
VALUES
(
  'Artsband',
  'หมวดดนตรี ศิลปะ และการแสดง',
  'ชมรมดนตรีสากลคณะอักษรศาสตร์ รวมพลคนรักดนตรีและเสียงเพลง',
  'artsband.chula',
  'https://images.unsplash.com/photo-1511671782779-c97d3d27a1d4',
  true
),
(
  'Arts Drama',
  'หมวดดนตรี ศิลปะ และการแสดง',
  'ชมรมศิลปะการละครของคณะอักษรศาสตร์ ร่วมสร้างสรรค์ละครเวทีอันงดงาม',
  'artsdrama.chula',
  'https://images.unsplash.com/photo-1507676184212-d03ab07a01bf',
  true
),
(
  'Arts Boardgame',
  'หมวดเกมและนันทนาการ',
  'ชมรมบอร์ดเกม คลับสำหรับผู้รักการเล่นและสร้างสรรค์บอร์ดเกมเพื่อความสนุกและการคิดวิเคราะห์',
  'artsboardgame',
  'https://images.unsplash.com/photo-1610890716171-6b1bb98ffd09',
  true
),
(
  'Arts E-Sport',
  'หมวดเกมและนันทนาการ',
  'ส่งเสริมการเล่นเกมอย่างสร้างสรรค์และการแข่งขันในระดับมหาวิทยาลัย',
  'artsesport',
  'https://images.unsplash.com/photo-1542751371-adc38448a05e',
  true
),
(
  'Arts Running Club',
  'หมวดกีฬาและการออกกำลังกาย',
  'ชมรมวิ่งเพื่อสุขภาพ สนับสนุนให้นักศึกษามาร่วมวิ่งและออกกำลังกายร่วมกัน',
  null,
  'https://images.unsplash.com/photo-1476480862126-209bfaa8edc8',
  true
),
(
  'Arts Badminton',
  'หมวดกีฬาและการออกกำลังกาย',
  'ชมรมแบดมินตันสำหรับผู้ที่ชื่นชอบการออกกำลังกายด้วยการตีแบด',
  null,
  'https://images.unsplash.com/photo-1626224583764-f87db24ac4ea',
  true
),
(
  'Arts Debate Club',
  'หมวดวิชาการและภาษา',
  'ชมรมโต้วาทีและพัฒนาทักษะการพูดภาษาอังกฤษและภาษาไทยในที่สาธารณะ',
  'artsdebate',
  'https://images.unsplash.com/photo-1524178232363-1fb2b075b655',
  true
),
(
  'Arts Translation Club',
  'หมวดวิชาการและภาษา',
  'ชมรมแปลภาษา ฝึกฝนทักษะการแปลวรรณกรรม บทความ และงานแปลประเภทต่างๆ',
  null,
  'https://images.unsplash.com/photo-1455390582262-044cdead277a',
  true
),
(
  'Arts Photo Club',
  'หมวดดนตรี ศิลปะ และการแสดง',
  'ชมรมถ่ายภาพเพื่อเก็บภาพบรรยากาศและความทรงจำของชาวอักษรฯ',
  'artsphotoclub',
  'https://images.unsplash.com/photo-1452780212940-6f5c0d14d848',
  true
),
(
  'Arts Volunteer',
  'หมวดกีฬาและการออกกำลังกาย',
  'ชมรมค่ายอาสาและพัฒนาชุมชน ออกค่ายสร้างสรรค์สิ่งดีๆ ให้สังคม',
  'artsvolunteer',
  'https://images.unsplash.com/photo-1488521787991-ed7bbaae773c',
  true
);
