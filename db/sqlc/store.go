package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) 
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

// SQLStore provides all functions to execute SQL queries and transactions // real db
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store { 
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

// ในความเป็นจริงหรือ production เรามักต้องการ pool of connections เพื่อจัดการกับ multiple request ใน parallet แทนการแค่จะสร้างแค่ 1 single connection ซึ่ง pgxpool package ทำได้