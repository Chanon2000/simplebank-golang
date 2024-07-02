package db // ทำการ implement transaction เพื่อ create new user ใน file นี้

import "context"

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error // จะคือ callback function ซึ่งจะคือ function ที่ exected หลังจาก user ถูก inserted ใน transaction เดียวกันนี้
}

type CreateUserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams) // เก็บ user output จาก CreateUser ลง result.User
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.User) // AfterCreate จะ return error ซึ่งถ้ามัน return error ตัว execTx ก็จะทำการ rollback transaction นี้ทันที
	})

	return result, err
}
