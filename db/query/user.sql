-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  full_name,
  email
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
  -- วิธีที่ 1 ในการทำให้ UpdateUser สามารถเลือก update แค่บาง fields ได้
  -- hashed_password = CASE -- CASE นี้ เพื่อจะได้สามารถ update แค่ field ใด field นึงต่อครั้งได้ -- ศึกษาเพิ่มเติมได้ที่ doc ของ sqlc
  --   WHEN @set_hashed_password::boolean = TRUE THEN @hashed_password -- ใส่ sqlc.arg(set_hashed_password)::boolean เพื่อกำหนด name และ type (ถ้าใส่เป็น $1 มันจะกำหนดชื่อให้เอา)เป็น set_hashed_password กับ boolean ซึ่งเขียน sqlc.arg() ย่อได้เป็น @
  --   -- ถ้าไม่กำหนด type (::boolean) ใน go code จะทำให้ parameter นี้กลายเป็น interface{}
  --   -- @hashed_password หลัง THEN คือกำหนดชื่อ parameter เช่นกัน
  --   ELSE hashed_password -- แต่ hashed_password หลัง ELSE นั้นคือ hashed_password column นะ ไม่ใช่การกำหนดชื่อ parameter
  -- END,
  -- full_name = CASE
  --   WHEN @set_full_name::boolean = TRUE THEN @full_name
  --   ELSE full_name
  -- END,
  -- email = CASE
  --   WHEN @set_email::boolean = TRUE THEN @email
  --   ELSE email
  -- END

  -- วิธีที่ 2 : ใช้ COALESCE วิธีที่ 1 (อ่านเพิ่มเติมของ COALESCE ที่ doc ของ sqlc)
  hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
  password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at), -- narg ย่อมาจาก nullable argument
  full_name = COALESCE(sqlc.narg(full_name), full_name),
  email = COALESCE(sqlc.narg(email), email)
WHERE
  username = @username -- เนื่องจากถ้าเราใช้ @ ใช้การกำหนด parameter เราต้องใช้มันในทุก parameter เลย ทำให้ตรงนี้เราเลยกำหนดเป็น @username ด้วยนั้นเอง
RETURNING *;

-- ทุกครั้งที่รัน make sqlc อาจลืมรัน make mock ต่อด้วย