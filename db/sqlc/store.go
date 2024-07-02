package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) 
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) // เอา CreateUserTx ใส่ลง Store ด้วยเพื่อให้ง่ายต่อการ mocked เมื่อต้องการทำ unit test เป็นต้น
}

// SQLStore provides all functions to execute SQL queries and transactions // real db
type SQLStore struct {
	db *sql.DB
	*Queries
}

// NewStore creates a new store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db: db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q) // มันรัน fn จนเสร็จก็ ซึ่งถ้าเป็นใน CreateUserTx ก็คือรัน CreateUser เรียบร้อย(แต่ยังไม่ได้ commit) และรัน AfterCreate ซึ่งก็คือ .taskDistributor.DistributeTaskSendVerifyEmail นั้นคือส่ง send email task เรียบร้อย
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	
	// time.Sleep(2 * time.Second) // sleep -> 2s // เมื่อกำหนดให้รอ 2s ก่อนค่อย commit โดยที่ worker จะเอา task มารันเลยโดยไม่มี delay นั้นทำให้ worker เอา task มา process ก่อน ซึ่ง send_email task นั้นมันต้องเข้าไปอ่านข้อมูล User แต่ ข้อมูล user มัน update หรือ commit เข้า db ไม่ทัน เลยทำให้เกิด error นั้นเอง เพราะ record ที่ task นั้นต้องการอ่านมัน commit เข้ามาไม่ทัน
	// เนื่องจาก transactions ในงานจริงนั้นมันไม่ได้ commit ในเวลาอันรวดเร็วเสมอไป เพราะบางครั้งมันก็ช้า ตามจำนวน traffic ที่เข้ามาใน system เรา เป็นต้น
	return tx.Commit()
}