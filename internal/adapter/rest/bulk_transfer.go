package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/angusgmorrison/hexagonal/internal/service"
	"github.com/gin-gonic/gin"
)

// TransactionService describes the methods a handler expects to call to perform
// transfers. Using an interface allows us to mock services when
// testing. This interface should follow the concrete service type.
type TransactionService interface {
	BulkTransaction(ctx context.Context, bt service.BulkTransaction) error
}

// Statically verify that the interface and concrete type remain in sync.
var _ TransactionService = (*service.TransactionService)(nil)

// bulkTransferRequest represents an incoming bulk transfer payload.
type bulkTransferRequest struct {
	OrganizationName string          `json:"organization_name" binding:"required"`
	OrganizationBIC  string          `json:"organization_bic" binding:"required"`
	OrganizationIBAN string          `json:"organization_iban" binding:"required"`
	CreditTransfers  creditTransfers `json:"credit_transfers" binding:"min=1"`
}

func (btr bulkTransferRequest) toDomain() service.BulkTransaction {
	bt := service.BulkTransaction{
		Account: service.BankAccount{
			OrganizationName: btr.OrganizationName,
			OrganizationBIC:  btr.OrganizationBIC,
			OrganizationIBAN: btr.OrganizationIBAN,
		},
		Transactions: btr.CreditTransfers.toDomain(),
	}

	return bt
}

type creditTransfers []creditTransfer

func (cts creditTransfers) toDomain() service.Transactions {
	transactions := make(service.Transactions, 0, len(cts))

	for _, transfer := range cts {
		transactions = append(transactions, transfer.toDomain())
	}

	return transactions
}

type creditTransfer struct {
	Amount           json.Number `json:"amount" binding:"required,numeric"`
	Currency         string      `json:"currency" binding:"required"`
	CounterpartyName string      `json:"counterparty_name" binding:"required"`
	CounterpartyBIC  string      `json:"counterparty_bic" binding:"required"`
	CounterpartyIBAN string      `json:"counterparty_iban" binding:"required"`
	Description      string      `json:"description" binding:"required"`
}

func (ct creditTransfer) toDomain() service.Transaction {
	return service.Transaction{
		AmountCents:      ct.amountCents(),
		Currency:         ct.Currency,
		CounterpartyName: ct.CounterpartyName,
		CounterpartyBIC:  ct.CounterpartyBIC,
		CounterpartyIBAN: ct.CounterpartyIBAN,
		Description:      ct.Description,
	}
}

func (ct creditTransfer) amountCents() int64 {
	// We validate that ct.Amount is numeric when binding, so we can safely
	// ignore the error from ParseFloat.
	f, _ := ct.Amount.Float64()

	return int64(f * 100)
}

// BulkTransfer receives bulk transfer requests over HTTP and executes them.
func (s *Server) handleCreateBulkTransfer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var btr bulkTransferRequest
		if err := c.ShouldBind(&btr); err != nil {
			s.logger.Printf("Failed to parse bulk transfer request: %s", err)
			c.AbortWithStatus(http.StatusBadRequest)

			return
		}

		if err := s.transactionService.BulkTransaction(c, btr.toDomain()); err != nil {
			s.logger.Printf("Bulk transfer failed: %s", err)
			c.AbortWithStatus(http.StatusUnprocessableEntity)

			return
		}

		c.Status(http.StatusCreated)
	}
}
