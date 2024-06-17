-- เขียน down-script เองโดยก็คือต้องทำการเขียน sql ที่ revert กลับนั้นเอง
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";

ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey"; -- accounts_owner_fkey เอาชื่อ foreign key มาจาก postgres

DROP TABLE IF EXISTS "users";
