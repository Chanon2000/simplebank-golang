package db

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLStore provides all functions to execute SQL queries and transactions // เนื่องจาก Queries มันทำแค่ทีละ 1 operation ไม่ได้ support transaction ที่ต้องทำหลาย operation ดังนั้นเราเลยขยายความสามารถที่ store struct เพิ่ม embed Queries struct เข้ามาด้วย
type Store struct {
	*Queries // embed Queries struct ที่ struct นี้
	db *sql.DB
}

// NewStore creates a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error { // execTx ขึ้นต้นด้วย lowcase เพราะไม่จะใช้ func นี้แค่ใน package
	tx, err := store.db.BeginTx(ctx, nil) // ใส่ nil เพื่อไม่ให้ default level ของ database ถูกใช้
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		// ถ้า err ไม่ nil เราจะทำการ rollback transaction โดยเรียก tx.Rollback() โดย Rollback() จะ return rollback error เช่นกัน 
		if rbErr := tx.Rollback(); rbErr != nil {
			// ถ้า rbErr ไม่ nil ก็จะ report 2 error
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	// commit เมื่อทุก operation นั้น success ด้วย tx.Commit()
	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// var txKey = struct{}{} // {} ที่ 2 หมายความว่าเราสร้าง new empty object ให้กับ type นี้

// TransferTx performs a money transfer from one account to the other.
// It creates the transfer, add account entries, and update accounts' balance within a database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// txName := ctx.Value(txKey)

		// fmt.Println(txName, "create transfer") // กำหนด transfer transaction name และ print มันตรงนี้ เพื่อจะได้ debug เรื่อง deadlock ได้
		// creates the transfer
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID, // จะเห็นว่าเราใช้ arg variable ที่อยู่นอก callback function ได้ซึ่งนี้เรียกว่า closure
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// fmt.Println(txName, "create entry 1")
		// add account entries
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// fmt.Println(txName, "create entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}
		
		// fmt.Println(txName, "get account 1")
		// get account -> update its balance
		// account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		// if err != nil {
		// 	return err
		// }

		// fmt.Println(txName, "update account 1")
		// result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 	ID: arg.FromAccountID,
		// 	Balance: account1.Balance - arg.Amount,
		// })
		// if err != nil {
		// 	return err
		// }

		// fmt.Println(txName, "get account 2")
		// account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		// if err != nil {
		// 	return err
		// }

		// fmt.Println(txName, "update account 2")
		// result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		// 	ID: arg.ToAccountID,
		// 	Balance: account2.Balance + arg.Amount,
		// })
		// if err != nil {
		// 	return err
		// }

		// มีอีกวิธีที่ดีกว่าเพราะลด query นั้นคือใช้ AddAccountBalance
		// code ตรงนี้มันทำการ update balance ของ from-account และ to-account ซึ่งตอนรัน test คุณรันมันแบบ concurrent ซึ่งทำให้ถ้า 2 concurrent transactions นั้นเข้าไป update account เดียวกัน อาจทำให้เกิด deadlock ที่ postgres ได้
		// if arg.FromAccountID < arg.ToAccountID { // จากการรัน TestTransferTxDeadlock แล้วทำให้เกิด deadlock ที่ส่วนนี้นั้น คุณเลยเอา account id มาเพื่อให้แต่ละ transactions นั้น update account ในลำดับเดียวกัน เพื่อไม่ให้เกิด deadlock ใน case ของ TestTransferTxDeadlock นั้นเอง
		// 	result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		// 		ID: arg.FromAccountID,
		// 		Amount: -arg.Amount,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
	
		// 	result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		// 		ID: arg.ToAccountID,
		// 		Amount: arg.Amount,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
		// } else {
		// 	result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		// 		ID: arg.ToAccountID,
		// 		Amount: arg.Amount,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}

		// 	result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		// 		ID: arg.FromAccountID,
		// 		Amount: -arg.Amount,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		
		// ทำให้ code สั้นลง
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}
		
		return err
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return // เนื่องจากเราใส่ name parameter ที่ return type ด้วยเลยทำให้ เมื่อเราใส่แค่ return keyword ตรงนี้มันก็จะ return variable ตาม name ที่คุณกำหนด ที่ return type เลยนั้นเอง
		// คล้ายกับการเขียนแบบนี้ return account1, account2, err
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}