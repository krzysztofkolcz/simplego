package transactions_test

import (
	"testing"

	"simpleGo/transactions"
)

func TestTransaction000(t *testing.T) {

	t.Run("Should run transaction", func(*testing.T) {

		transactions.Transaction000()

	})

}
