-- name: CreateAccount :one
INSERT INTO accounts (
  owner,
  balance,
  currency
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE; -- ถ้าเป็น GetAccountForUpdate มันจะ block transaction อื่นจนกว่าจะ commit หรือทำงานเสร็จ แต่ว่าถ้าเป็น GetAccount มันจะไม่ block ทำให้อาจเกิดการ get value เก่าก่อน update ได้ นั้นเอง
-- เติม NO KEY เพื่อบอก postgres ไม่ต้องไป update key หรือ ID column ของ account table ซึ่งแก้ปัญหา deadlock ตอนรัน TestTransferTx test ด้วย (ที่เกิดจาก FOREIGN KEY ระหว่าง table)

-- name: ListAccounts :many
SELECT * FROM accounts
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: AddAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount) -- ใส่ sqlc.arg(amount) แทนใส่เป็น $2 เพื่อให้ชื่อ arg ตอน generate go code นั้นชื่อ Amount แทนนั้นเอง
WHERE id = sqlc.arg(id) -- แทนใส่เป็น $1 เพื่อบอก sqlc ให้ generate parameter name เป็น ID 
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;