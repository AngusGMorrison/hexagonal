// Package handler provides http handlers.
package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/angusgmorrison/hexagonal/internal/app/transferdomain"
	"github.com/gin-gonic/gin"
)

// TransferHandler specifies endpoints for handling bulk transfer
type TransferHandler struct {
	logger  *log.Logger
	service *transferdomain.Service
}

func NewTransferHandler(logger *log.Logger, service *transferdomain.Service) *TransferHandler {
	return &TransferHandler{
		logger:  logger,
		service: service,
	}
}

// BulkTransfer receives bulk transfer requests over HTTP and executes them.
func (th *TransferHandler) BulkTransfer(c *gin.Context) {
	var btr bulkTransferRequest
	if err := c.ShouldBind(&btr); err != nil {
		th.logger.Printf("failed to parse bulk transfer request: %s", err)
		c.AbortWithStatus(http.StatusUnprocessableEntity)

		return
	}

	if err := th.service.PerformBulkTransfer(
		c, btr.ToDomain(), transferdomain.ValidateBulkTransfer,
	); err != nil {
		th.logger.Printf("bulk transfer failed: %s", err)
		c.AbortWithStatus(http.StatusUnprocessableEntity)

		return
	}

	c.Status(http.StatusCreated)
}

// bulkTransferRequest represents an incoming bulk transfer payload.
type bulkTransferRequest struct {
	OrganizationName string          `json:"organization_name" binding:"required"`
	OrganizationBIC  string          `json:"organization_bic" binding:"required"`
	OrganizationIBAN string          `json:"organization_iban" binding:"required"`
	CreditTransfers  creditTransfers `json:"credit_transfers" binding:"required"`
}

func (btr bulkTransferRequest) ToDomain() transferdomain.BulkTransfer {
	bt := transferdomain.BulkTransfer{
		Account: transferdomain.BankAccount{
			OrganizationName: btr.OrganizationName,
			OrganizationBIC:  btr.OrganizationBIC,
			OrganizationIBAN: btr.OrganizationIBAN,
		},
		CreditTransfers: btr.CreditTransfers.toDomain(),
	}

	return bt
}

type creditTransfers []creditTransfer

func (cts creditTransfers) toDomain() []transferdomain.CreditTransfer {
	domainTransfers := make([]transferdomain.CreditTransfer, len(cts))

	for i, transfer := range cts {
		domainTransfers[i] = transfer.toDomain()
	}

	return domainTransfers
}

type creditTransfer struct {
	Amount           json.Number `json:"amount" binding:"required,numeric"`
	Currency         string      `json:"currency" binding:"required"`
	CounterpartyName string      `json:"counterparty_name" binding:"required"`
	CounterpartyBIC  string      `json:"counterparty_bic" binding:"required"`
	CounterpartyIBAN string      `json:"counterparty_iban" binding:"required"`
	Description      string      `json:"description" binding:"required"`
}

func (ct creditTransfer) toDomain() transferdomain.CreditTransfer {
	return transferdomain.CreditTransfer{
		AmountCents:      ct.amountCents(),
		Currency:         ct.Currency,
		CounterpartyName: ct.CounterpartyName,
		CounterpartyBIC:  ct.CounterpartyBIC,
		CounterpartyIBAN: ct.CounterpartyIBAN,
		Description:      ct.Description,
	}
}

//nolint:gomnd
func (ct creditTransfer) amountCents() int64 {
	// We validate that ct.Amount is numeric when binding, so we can safely
	// ignore the error from ParseFloat.
	f, _ := ct.Amount.Float64()

	return int64(f * 100)
}
