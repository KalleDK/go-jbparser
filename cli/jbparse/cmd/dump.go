package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/KalleDK/go-jbparser/jbparser"
	"github.com/KalleDK/go-jbparser/jbparser/jbpage"
	"github.com/KalleDK/go-money/money"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var dumpReverse bool
var dumpTimes int

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump [file]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var account jbparser.AccountStatement
		if len(args) == 0 {
			account = stdinAccount()

		} else {
			account = loadAccount(args[0])
		}
		printTransactions(account, dumpTimes, dumpReverse)
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().BoolVarP(&dumpReverse, "reverse", "r", false, "reverse the sorting")
	dumpCmd.Flags().IntVarP(&dumpTimes, "times", "n", 0, "number of rows to show")
}

const (
	aLeft   = tablewriter.ALIGN_LEFT
	aRight  = tablewriter.ALIGN_RIGHT
	aCenter = tablewriter.ALIGN_CENTER
)

type tFormat struct {
	Title string
	Align int
}

var mfmt = money.Formatter{
	Decimal:  ',',
	Thousand: '.',
}

func fmtBool(b bool) string {
	if b {
		return "x"
	}
	return " "
}

func printTable(w io.Writer, formats []tFormat, d [][]string) {
	headers := []string{}
	align := []int{}
	for _, header := range formats {
		headers = append(headers, header.Title)
		align = append(align, header.Align)
	}

	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetHeader(headers)
	table.SetColumnAlignment(align)
	table.AppendBulk(d)
	table.Render()
}

func stdinAccount() jbparser.AccountStatement {
	fmt.Println("Paste html:")
	buffer := &bytes.Buffer{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		d := []byte(scanner.Text())
		lt := len(d)
		if lt < 1 {
			continue
		}
		if d[lt-1] == 24 {
			break
		}
		if lt >= 7 {
			/*
				if string(d[lt-7:]) == "</html>" {
					break
				}
			*/
			if strings.Contains(string(d), "</html>") {
				break
			}
		}
		buffer.Write(d)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	account, err := jbpage.Parse(buffer)
	if err != nil {
		log.Fatal(err)
	}

	return account
}

func loadAccount(path string) jbparser.AccountStatement {

	filePath := path
	fp, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	account, err := jbpage.Parse(fp)
	if err != nil {
		log.Fatal(err)
	}

	return account

}

func printTransactions(account jbparser.AccountStatement, n int, reverse bool) {

	tableFormat := []tFormat{
		{"Reg", aLeft},
		{"Number", aLeft},
		{"Account", aLeft},
		{"UseDate", aLeft},
		{"PostDate", aLeft},
		{"Text", aLeft},
		{"Category", aLeft},
		{"Amount", aRight},
		{"Balance", aRight},
		{"Reconciled", aCenter},
	}

	transactions := account.Transactions
	lt := len(transactions)
	if n > 0 && lt > n {
		transactions = transactions[:n]
	}
	if n < 0 && lt > -n {
		transactions = transactions[lt+n:]
	}

	rows := [][]string{}
	for _, t := range transactions {
		rows = append(rows, []string{
			fmt.Sprint(t.Account.Reg),
			fmt.Sprint(t.Account.Number),
			t.Account.Name,
			t.UseDate.Format("Mon, 02 Jan 2006 15:04:05"),
			t.PostDate.Format("Mon, 02 Jan 2006"),
			t.Text,
			t.Category,
			mfmt.Sprint(t.Amount),
			mfmt.Sprint(t.Balance),
			fmtBool(t.Reconciled),
		})
	}

	if reverse {
		l := len(rows)
		for i := 0; i < l/2; i = i + 1 {
			rows[i], rows[l-i-1] = rows[l-i-1], rows[i]
		}
	}

	printTable(os.Stdout, tableFormat, rows)
}
