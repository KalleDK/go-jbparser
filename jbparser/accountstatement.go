package jbparser

import (
	"time"

	"github.com/KalleDK/go-money/money"
)

const (
	usetimeFormat  = "2006-01-02-15.04.05.999999"
	posttimeFormat = "02.01.2006"
	lillian        = 2299160          // Julian day of 15 Oct 1582
	unix           = 2440587          // Julian day of 1 Jan 1970
	epoch          = unix - lillian   // Days between epochs
	g1582          = epoch * 86400    // seconds between epochs
	g1582ns100     = g1582 * 10000000 // 100s of a nanoseconds between epochs
)

type Amount = money.Amount

type AccountStatement struct {
	Info         AccountInfo
	Transactions []Transaction
}

type AccountInfo struct {
	Name   string
	Reg    uint64
	Number uint64
}

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
