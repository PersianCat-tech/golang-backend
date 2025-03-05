package db

import (
	"testing"

	"github.com/techschool/simplebank/db/util"
)


func createRandomEntry (t *testing.T, account Account) {
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount: util.RandomMoney(),
	}
}

func TestCreateEntry(t *testing.T) {
	
}