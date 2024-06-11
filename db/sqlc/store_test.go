package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// การเขียน database transaction นั้นเป็นสิ่งที่คุณต้องระวัง ซึ่งมันง่ายที่จะเขียน แต่มันอาจกลายเป็นฝันร้ายได้ถ้าคุณไม่จัดการ concurrency อย่างระมัดระวัง
	// สิ่งที่จะทำให้มั้นใจว่า transaction นั้นทำงานได้ดี คือรันมันเป็นหลายๆ concurrent go routines
	// run n concurrent transfer transactions
	n := 5 // เราจะรัน 5 transactions
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() { // ใช้ go keyword เพื่อ start new routine
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: 	account1.ID,
				ToAccountID: 	account2.ID,
				Amount: 		amount,
			})

			errs <- err
			results <- result
		}() // ใส่ () เพื่อ call function ในทำงาน
	}

	// check results // จาก channel
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// TODO: check account's balance : เนื่องจากเรายังไม่ได้ implement ในส่วนของการ update accounts' balance 

	}
}