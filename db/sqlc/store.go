package db

import (
	"context"
	"database/sql"
	"fmt"
)

//store provides all functions to execute db queries and transaction

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback Err: %v", err, rbErr)
		}
		return err //回滚成功则返回原始交易错误
	}

	return tx.Commit()
}

// TransferTxParms contains the input parameters of transfer transation
type TransferTxParms struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult is the result of the transfer transaction
type TransferTxResult struct {
	Transfer Transfer	`json:"transfer"`
	FromAccount Account `json:"from_account"`
	ToAccount Account 	`json:"to_account"`
	FromEntry Entry		`json:"from_entry"`
	ToEntry   Entry		`json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another account
// It creates a transfer record, add account entries, and update accounts' balance within a single transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParms) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error{
		var err error
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount,	//转出账户
		})

		if err != nil {
			return err
		}

		result.ToEntry, err =  q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})

		if err != nil {
			return err
		}

		//TODO:	update accounts' balance	需要防止潜在的死锁

		return nil
	})

	return result, err
}
