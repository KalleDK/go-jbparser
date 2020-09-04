package jbparser

import (
	"time"

	"github.com/KalleDK/go-money/money"
)

const (
	lillian    = 2299160          // Julian day of 15 Oct 1582
	unix       = 2440587          // Julian day of 1 Jan 1970
	epoch      = unix - lillian   // Days between epochs
	g1582      = epoch * 86400    // seconds between epochs
	g1582ns100 = g1582 * 10000000 // 100s of a nanoseconds between epochs
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
