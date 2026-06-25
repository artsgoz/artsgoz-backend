-- Seed data: 10 dummy professors for local development
-- Run: docker cp scripts/seed_professors.sql artsgoz-postgres-local:/tmp/seed.sql && docker exec artsgoz-postgres-local psql -U artsgoz -d artsgoz -f /tmp/seed.sql

DELETE FROM professors;

INSERT INTO professors (name, name_en, department, location, achievements, qualifications, courses, is_active)
VALUES
(
  'ผศ.ดร.สมหญิง รักภาษา',
  'Asst. Prof. Dr. Somying Rakphasa',
  'ภาควิชาภาษาไทย',
  'อาคารมหาจุฬาลงกรณ์ ห้อง 301',
  ARRAY['รางวัลอาจารย์ดีเด่น คณะอักษรศาสตร์ ปี 2565', 'รางวัลงานวิจัยดีเด่น ปี 2563'],
  ARRAY['Ph.D. Thai Language, Chulalongkorn University', 'M.A. Linguistics, Chulalongkorn University'],
  ARRAY['2200111 ภาษาไทยเพื่อการสื่อสาร', '2200211 วากยสัมพันธ์ภาษาไทย'],
  true
),
(
  'รศ.ดร.ประทีป วรรณศิลป์',
  'Assoc. Prof. Dr. Prateep Wannasin',
  'ภาควิชาภาษาอังกฤษ',
  'อาคารมหาจุฬาลงกรณ์ ห้อง 412',
  ARRAY['Best Paper Award, PACLIC 2022', 'Fulbright Scholar 2019'],
  ARRAY['Ph.D. Applied Linguistics, University of Edinburgh', 'M.A. TESOL, University of Leeds'],
  ARRAY['2201111 ภาษาอังกฤษเพื่อการสื่อสาร', '2201311 การแปลเบื้องต้น'],
  true
),
(
  'อ.ดร.วิภาพร ประวัติศาสตร์',
  'Dr. Wipaporn Prawattisart',
  'ภาควิชาประวัติศาสตร์',
  'อาคารบรมราชกุมารี ห้อง 205',
  ARRAY['ทุนวิจัย Newton Fund ปี 2564'],
  ARRAY['Ph.D. History, SOAS University of London', 'M.A. Southeast Asian Studies, Kyoto University'],
  ARRAY['2203111 ประวัติศาสตร์ไทย', '2203211 ประวัติศาสตร์เอเชียตะวันออกเฉียงใต้'],
  true
),
(
  'ผศ.ดร.มณีรัตน์ ปรัชญาดี',
  'Asst. Prof. Dr. Maneerat Prachyadee',
  'ภาควิชาปรัชญา',
  'อาคารมหาจุฬาลงกรณ์ ห้อง 508',
  ARRAY['รางวัลนักวิชาการดีเด่น สมาคมปรัชญาไทย ปี 2566'],
  ARRAY['Ph.D. Philosophy, University of Cambridge', 'B.A. Philosophy, Chulalongkorn University'],
  ARRAY['2204111 ปรัชญาเบื้องต้น', '2204311 จริยศาสตร์'],
  true
),
(
  'อ.นิโคลาส์ มาร์แตง',
  'Nicolas Martin',
  'ภาควิชาภาษาฝรั่งเศส',
  'อาคารบรมราชกุมารี ห้อง 118',
  ARRAY[]::TEXT[],
  ARRAY['M.A. Litterature Francaise, Sorbonne', 'DELF C2'],
  ARRAY['2205111 ภาษาฝรั่งเศสเบื้องต้น 1', '2205112 ภาษาฝรั่งเศสเบื้องต้น 2'],
  true
),
(
  'รศ.ดร.กนกวรรณ ศิลปการละคร',
  'Assoc. Prof. Dr. Kanokwan Silpakarnlakorn',
  'ภาควิชาศิลปการละคร',
  'อาคารศิลปการละคร ห้อง 301',
  ARRAY['รางวัลศิษย์เก่าดีเด่น จุฬาฯ ปี 2560', 'ศิลปินแห่งชาติ สาขาการแสดง ปี 2568'],
  ARRAY['M.F.A. Theatre Directing, New York University', 'B.A. Dramatic Arts, Chulalongkorn University'],
  ARRAY['2207111 ศิลปะการแสดงเบื้องต้น', '2207211 การกำกับการแสดง'],
  true
),
(
  'ผศ.ดร.ธนากร ภูมิศาสตร์วิทย์',
  'Asst. Prof. Dr. Thanakornd Phumisaftwit',
  'ภาควิชาภูมิศาสตร์',
  'อาคารมหาจุฬาลงกรณ์ ห้อง 620',
  ARRAY['ทุน JSPS Fellowship ปี 2563', 'รางวัลบทความดีเด่น วารสารภูมิศาสตร์ไทย ปี 2565'],
  ARRAY['Ph.D. Geography, Osaka University', 'M.Sc. Remote Sensing, AIT'],
  ARRAY['2208111 ภูมิศาสตร์กายภาพ', '2208211 ระบบสารสนเทศภูมิศาสตร์'],
  true
),
(
  'อ.ดร.เฮลมุท ชไนเดอร์',
  'Dr. Helmut Schneider',
  'ภาควิชาภาษาเยอรมัน',
  'อาคารบรมราชกุมารี ห้อง 220',
  ARRAY[]::TEXT[],
  ARRAY['Dr.phil. Germanistik, Humboldt-Universitaet zu Berlin', 'DAAD Scholarship Alumni'],
  ARRAY['2209111 ภาษาเยอรมันเบื้องต้น 1', '2209211 ภาษาเยอรมันระดับกลาง'],
  true
),
(
  'ผศ.ดร.คาร์ลอส โรดริเกซ',
  'Asst. Prof. Dr. Carlos Rodriguez',
  'ภาควิชาภาษาสเปน',
  'อาคารมหาจุฬาลงกรณ์ ห้อง 715',
  ARRAY['Instituto Cervantes Teaching Excellence Award 2022'],
  ARRAY['Ph.D. Hispanic Linguistics, Univ. Complutense de Madrid', 'DELE C2'],
  ARRAY['2210111 ภาษาสเปนเบื้องต้น 1', '2210311 วรรณคดีสเปน'],
  true
),
(
  'รศ.ดร.อรุณี จิตวิทยา',
  'Assoc. Prof. Dr. Arunee Jitwittaya',
  'ภาควิชาจิตวิทยา',
  'อาคารมหาจุฬาลงกรณ์ ห้อง 830',
  ARRAY['รางวัลนักวิจัยดาวรุ่ง วช. ปี 2562', 'รางวัลอาจารย์ดีเด่นด้านการสอน คณะอักษรศาสตร์ ปี 2564'],
  ARRAY['Ph.D. Clinical Psychology, University of Michigan', 'M.Sc. Psychology, Mahidol University'],
  ARRAY['2211111 จิตวิทยาทั่วไป', '2211211 จิตวิทยาพัฒนาการ', '2211311 จิตวิทยาการปรึกษา'],
  true
);
