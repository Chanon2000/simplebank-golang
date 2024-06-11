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

// TransferTx performs a money transfer from one account to the other.
// It creates the transfer, add account entries, and update accounts' balance within a database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// creates the transfer
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID, // จะเห็นว่าเราใช้ arg variable ที่อยู่นอก callback function ได้ซึ่งนี้เรียกว่า closure
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// add account entries
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}
		
		// TODO: update account's balance ซึ่งสิ่งส่วนนี้ค่อนข้างซับซ้อน เพราะว่า มันอาจทำให้เกิด locking, deadlock เป็นต้น
		// ทำให้เรากลับมาทำมันทีหลังใน lecture อื่น

		return nil
	})


	return result, err
}