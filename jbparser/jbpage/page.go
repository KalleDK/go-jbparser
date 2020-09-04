package jbpage

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/KalleDK/go-jbparser/jbparser"
	"github.com/KalleDK/go-money/money"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	usetimeFormat  = "2006-01-02-15.04.05.999999"
	posttimeFormat = "02.01.2006"
)

var (
	loc          *time.Location = time.UTC
	usetimeRegex                = regexp.MustCompile(`\\'id\\':\\'(\d+-\d+-\d+-\d+\.\d+\.\d+\.\d+)\\'`)
)

func parseTransaction(node *html.Node, info jbparser.AccountInfo) (t jbparser.Transaction, err error) {

	t.UseDate, err = getUseDate(node)
	if err != nil {
		return
	}

	t.PostDate, err = getPostDate(node)
	if err != nil {
		return
	}

	t.Account = info
	//t.Account, err = getAccount(node)
	if err != nil {
		return
	}

	t.Text, err = getText(node)
	if err != nil {
		return
	}

	t.Amount, err = getAmount(node)
	if err != nil {
		return
	}

	t.Balance, err = getBalance(node)
	if err != nil {
		return
	}

	t.Reconciled, err = getReconciled(node)
	if err != nil {
		return
	}

	t.Category, err = getCategory(node)
	if err != nil {
		return
	}

	return t, nil
}

// Parse reads a webpage stream and parses it
func Parse(r io.Reader) (jbparser.AccountStatement, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return jbparser.AccountStatement{}, err
	}

	return ParseHTML(doc)
}

// ParseHTML parses a webpage
func ParseHTML(node *html.Node) (jbparser.AccountStatement, error) {

	info, err := getAccountInfo(node)
	if err != nil {
		return jbparser.AccountStatement{}, err
	}

	// Find the div that contains the transactions
	transactions := []jbparser.Transaction{}
	for _, elem := range getTransactions(node) {
		transaction, err := parseTransaction(elem, info)
		if err != nil {
			return jbparser.AccountStatement{}, err
		}
		transactions = append(transactions, transaction)
	}

	return jbparser.AccountStatement{
		Info:         info,
		Transactions: transactions,
	}, nil
}

// #region Parse Page

func getAccountInfo(node *html.Node) (jbparser.AccountInfo, error) {
	//account-selector-chosen
	elem := htmlquery.FindOne(node, ".//div[contains(@class, 'account-selector-chosen')]")
	if elem == nil {
		return jbparser.AccountInfo{}, fmt.Errorf("can't locate master")
	}

	name := strings.TrimSpace(htmlquery.InnerText(elem))

	//account-selector-name-and-number
	elems := htmlquery.Find(node, ".//div[contains(@class, 'account-selector-name-and-number')]")
	if len(elems) < 1 {
		return jbparser.AccountInfo{}, fmt.Errorf("can't locate master")
	}
	for _, el := range elems {
		if strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(el, ".//div[contains(@class, 'account-selector-account-name')]"))) == name {
			num := strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(el, ".//div[contains(@class, 'account-selector-account-number')]")))
			nums := strings.Split(num, " ")
			reg, err := strconv.ParseUint(nums[0], 10, 64)
			if err != nil {
				return jbparser.AccountInfo{}, fmt.Errorf("can't locate master")
			}
			account, err := strconv.ParseUint(nums[1], 10, 64)
			if err != nil {
				return jbparser.AccountInfo{}, fmt.Errorf("can't locate master")
			}

			return jbparser.AccountInfo{
				Name:   name,
				Reg:    reg,
				Number: account,
			}, nil
		}
	}

	return jbparser.AccountInfo{}, fmt.Errorf("can't locate master")
}

func getTransactions(node *html.Node) []*html.Node {
	return htmlquery.Find(node, "(.//ul[contains(@class, 'old-postings')])[1]/div")
}

// #endregion

// #region Parse Transaction Row

func getUseDate(node *html.Node) (time.Time, error) {
	// Find element <a>
	elem := htmlquery.FindOne(node, ".//a")
	if elem == nil {
		return time.Time{}, fmt.Errorf("can't locate use date")
	}

	// Find the OnClick attribute
	attr := htmlquery.SelectAttr(elem, "onclick")

	// Search the OnClick attribute for timestamp
	m := usetimeRegex.FindStringSubmatch(attr)
	if m == nil {
		return time.Time{}, fmt.Errorf("can't locate use date")
	}

	// Parse the timestamp to correct format
	t, err := time.ParseInLocation(usetimeFormat, m[1], loc)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func getReconciled(node *html.Node) (bool, error) {
	// Find element <input class="... js-table-checkbox ...">
	elem := htmlquery.FindOne(node, ".//input[contains(@class, 'js-table-checkbox')]")
	if elem == nil {
		return false, fmt.Errorf("can't locate reconciled")
	}

	// Parse the value attribute
	attr := htmlquery.SelectAttr(elem, "checked")

	if attr == "checked" {
		return true, nil
	}

	return false, nil
}

func getPostDate(node *html.Node) (time.Time, error) {
	// Find element <div class="... posting-date-compact ...">
	elem := htmlquery.FindOne(node, ".//div[contains(@class, 'posting-date-compact')]")
	if elem == nil {
		return time.Time{}, fmt.Errorf("can't locate post date")

	}

	// Get the innerText and remove newlines / spaces
	text := strings.TrimSpace(htmlquery.InnerText(elem))

	// Parse the timestamp to correct format
	t, err := time.ParseInLocation(posttimeFormat, text, loc)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func getText(node *html.Node) (string, error) {
	// Find element <input class="... posting-text ...">
	elem := htmlquery.FindOne(node, ".//div[contains(@class, 'posting-text')]")
	if elem == nil {
		return "", fmt.Errorf("can't locate text")
	}

	return strings.TrimSpace(htmlquery.InnerText(elem)), nil
}

func getCategory(node *html.Node) (string, error) {
	// Find element <input class="... posting-category ...">
	elem := htmlquery.FindOne(node, ".//div[contains(@class, 'posting-category')]")
	if elem == nil {
		return "", fmt.Errorf("can't locate category")
	}

	return strings.TrimSpace(htmlquery.InnerText(elem)), nil
}

func getAccount(node *html.Node) (string, error) {
	// Find element <input class="... posting-account ...">
	elem := htmlquery.FindOne(node, ".//div[contains(@class, 'posting-account')]")
	if elem == nil {
		return "", fmt.Errorf("can't locate account")
	}

	return strings.TrimSpace(htmlquery.InnerText(elem)), nil
}

func getAmount(node *html.Node) (money.Amount, error) {
	// Find element <input class="... posting-amount ...">
	elem := htmlquery.FindOne(node, ".//div[contains(@class, 'posting-amount')]")
	if elem == nil {
		return 0, fmt.Errorf("can't locate amount")
	}

	s := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(htmlquery.InnerText(elem)), ",", ""), ".", "")

	v, err := strconv.ParseInt(s, 10, 64)
	return money.FromCenti(v), err

}

func getBalance(node *html.Node) (money.Amount, error) {
	// Find element <input class="... posting-balance ...">
	elem := htmlquery.FindOne(node, ".//div[contains(@class, 'posting-balance')]")
	if elem == nil {
		return 0, fmt.Errorf("can't locate balance")
	}

	s := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(htmlquery.InnerText(elem)), ",", ""), ".", "")

	v, err := strconv.ParseInt(s, 10, 64)
	return money.FromCenti(v), err

}

/*
func calcUUID(t time.Time, seq uint16, nodeid [6]byte) (uuid.UUID, error) {

	l := uint64(t.UnixNano()/100) + g1582ns100

	var uid uuid.UUID

	timeLow := uint32(l & 0xffffffff)
	timeMid := uint16((l >> 32) & 0xffff)
	timeHi := uint16((l >> 48) & 0x0fff)
	timeHi |= 0x1000 // Version 1

	binary.BigEndian.PutUint32(uid[0:], timeLow)
	binary.BigEndian.PutUint16(uid[4:], timeMid)
	binary.BigEndian.PutUint16(uid[6:], timeHi)
	binary.BigEndian.PutUint16(uid[8:], seq)
	copy(uid[10:], nodeid[:])

	return uid, nil
}
*/

// #endregion
