DROP TABLE IF EXISTS entries; -- เขียน sql ที่ revert กลับไป version ก่อนรัน up ใน init_schema migration นี้ -- ต้อง drop entries table ก่อนเพราะเหมือนมีแต่ relationship ปลายๆ
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS accounts;