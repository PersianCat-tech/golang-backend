package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	fmt.Println(">>before:", account1.Balance, account2.Balance)

	n := 5	//为了调试并发，暂时将n修改为2
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		
		
		go func() {
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParms{
				FromAccountID: account1.ID,
				ToAccountID: account2.ID,
				Amount: amount, 
			})

			errs <- err
			results <- result
		}()
	}

	//check results outside
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		//check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//check entries 
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)	//账户转出
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
		
		//check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)
		
		//TODO: check accounts' balance
		fmt.Println(">>tx:", fromAccount.Balance, toAccount.Balance)

		diff1 := account1.Balance - fromAccount.Balance	//从account1中流出的金额
		diff2 := toAccount.Balance - account2.Balance	//增加到account2的金额
		require.Equal(t, diff1, diff2)	//事务正常执行则diff1与diff2相等
		require.True(t, diff1 > 0)
		require.True(t, diff1 % amount == 0)	//转出的金额以amount为基本单位

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)

		require.NotContains(t, existed, k)
		existed[k] = true
	}

	//check the final updated balances 
	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">>after:", updateAccount1.Balance, updateAccount2.Balance)

	require.Equal(t, account1.Balance - int64(n)*amount, updateAccount1.Balance)
	require.Equal(t, account2.Balance + int64(n)*amount, updateAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	fmt.Println(">>before:", account1.Balance, account2.Balance)

	n := 10	//5个事务由账户一向账户二转帐，5个事务由账户二向账户一转账
	amount := int64(10)

	errs := make(chan error)
	
	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID
		
		if i % 2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			ctx := context.Background()
			_, err := store.TransferTx(ctx, TransferTxParms{
				FromAccountID: fromAccountID,
				ToAccountID: toAccountID,
				Amount: amount, 
			})

			errs <- err
			//results <- result
		}()
	}

	//check results outside
	//existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		
	}

	//check the final updated balances 
	updateAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">>after:", updateAccount1.Balance, updateAccount2.Balance)

	require.Equal(t, account1.Balance, updateAccount1.Balance)
	require.Equal(t, account2.Balance, updateAccount2.Balance)
}