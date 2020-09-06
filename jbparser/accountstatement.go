package jbparser

import (
	"time"

	"github.com/KalleDK/go-money/money"
)

// Amount is the common type to contain values
type Amount = money.Amount

// AccountStatement is Info about the account and some transactions
type AccountStatement struct {
	Info         AccountInfo
	Transactions []Transaction
}

// AccountInfo is basis info about the Account
type AccountInfo struct {
	Name   string
	Reg    uint64
	Number uint64
}

// Transaction contains all info about a transaction
type Transaction struct {
	UseDate    time.Time
	PostDate   time.Time
	Account    AccountInfo
	Text       string
	Amount     Amount
	Balance    Amount
	Category   string
	Reconciled bool
}
