package report

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jarismar/b3c-invoice-reader-lambda/constants"
	"github.com/jarismar/b3c-invoice-reader-lambda/utils"
	"github.com/jarismar/b3c-service-entities/entity"
)

type ConsoleReport struct {
	invoice *entity.Invoice
}

func GetConsoleReport(invoice *entity.Invoice) *ConsoleReport {
	return &ConsoleReport{
		invoice: invoice,
	}
}

func (report *ConsoleReport) printTaxGroup(taxGroupName string, taxGroup *entity.TaxGroup) {
	fmt.Println(taxGroupName)
	fmt.Printf("TaxGroup.Id ...... : %d\n", taxGroup.Id)
	fmt.Printf("TaxGroup.Source .. : %s\n", taxGroup.Source)
	fmt.Printf("TaxGroup.ExtId ... : %d\n", taxGroup.ExternalId)
	fmt.Printf(
		"%2s %10s %10s %8s %10s %7s\n",
		"#",
		"Code",
		"Base",
		"Rate",
		"Value",
		"Id",
	)
	for key, taxInstance := range taxGroup.Taxes {
		fmt.Printf(
			"%02d %10s %10.4f %8.5f %10.4f %7d\n",
			key,
			taxInstance.Tax.Code,
			taxInstance.BaseValue,
			taxInstance.TaxRate,
			taxInstance.TaxValue,
			taxInstance.Id,
		)
	}
}

func getAssetType(cmp *entity.Company) string {
	if cmp.BDR {
		return "BDR"
	}

	if cmp.ETF {
		return "ETF"
	}

	return "SHR"
}

func (report *ConsoleReport) printInvoiceItems() {
	invoice := report.invoice
	fmt.Printf("Invoice.Items .... : %d\n", len(invoice.Items))
	fmt.Printf(
		"%3s %8s %4s %6s %10s %6s %7s\n",
		"Ord",
		"Tag",
		"Type",
		"Qty",
		"Price",
		"Debit",
		"Id",
	)

	for _, item := range invoice.Items {
		fmt.Printf(
			"%3d %8s %4s %6d %10.2f %6s %7d\n",
			item.Order,
			item.Company.Code,
			getAssetType(item.Company),
			item.Qty,
			item.Price,
			strconv.FormatBool(item.Debit),
			item.Id,
		)
	}
}

func (report *ConsoleReport) printItemBatch() {
	invoice := report.invoice
	items := invoice.Items
	taxTypes := constants.TaxTypes
	itemBatchList := make([]entity.InvoiceItem, 0, len(items))

	for _, item := range items {
		if item.Debit {
			itemBatchList = append(itemBatchList, item)
		}
	}

	fmt.Printf("Item.ItemBatch ... : %d\n", len(itemBatchList))

	if len(itemBatchList) == 0 {
		return
	}

	fmt.Printf(
		"%3s %8s %5s %8s %10s %8s %10s %7s %7s %7s %7s %7s %7s %7s\n",
		"Ord",
		"Tag",
		"Qty",
		"Raw",
		"Avg",
		"Taxes",
		"Total",
		"SETFEE",
		"EMLFEE",
		"BRKFEE",
		"ISSFEE",
		"TaxGrId",
		"BatchId",
		"ItemId",
	)

	for _, item := range itemBatchList {
		itemBatch := item.ItemBatch
		taxGroup := itemBatch.TaxGroup

		setFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.SETFEE)
		emlFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.EMLFEE)
		brkFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.BRKFEE)
		issFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.ISSSPFEE)

		fmt.Printf(
			"%3d %8s %5d %8.2f %10.4f %8.4f %10.4f %7.4f %7.4f %7.4f %7.4f %7d %7d %7d\n",
			item.Order,
			item.Company.Code,
			item.Qty,
			itemBatch.RawPrice,
			itemBatch.AvgPrice,
			itemBatch.TotalTaxes,
			(itemBatch.AvgPrice * float64(itemBatch.Qty)),
			setFee,
			emlFee,
			brkFee,
			issFee,
			taxGroup.Id,
			item.Id,
			itemBatch.Id,
		)
	}
}

func (report *ConsoleReport) printTrade() {
	invoice := report.invoice
	items := invoice.Items
	taxTypes := constants.TaxTypes
	itemTradeList := make([]entity.InvoiceItem, 0, len(items))

	for _, item := range items {
		if !item.Debit {
			itemTradeList = append(itemTradeList, item)
		}
	}

	fmt.Printf("Item.Trade ....... : %d\n", len(itemTradeList))

	if len(itemTradeList) == 0 {
		return
	}

	fmt.Printf(
		"%3s %8s %5s %7s %7s %10s %7s %7s %7s %7s %7s %7s %7s %7s %7s\n",
		"Ord",
		"Tag",
		"Qty",
		"Avg",
		"Sell",
		"Res",
		"Taxes",
		"SETFEE",
		"EMLFEE",
		"BRKFEE",
		"ISSFEE",
		"IRFEE",
		"TaxGrId",
		"TradeId",
		"ItemId",
	)

	for _, item := range itemTradeList {
		trade := item.Trade
		taxGroup := trade.TaxGroup

		setFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.SETFEE)
		emlFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.EMLFEE)
		brkFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.BRKFEE)
		issFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.ISSSPFEE)
		irrFee := utils.GetTaxValueByGroup(taxGroup, taxTypes.IRRFFEE)

		fmt.Printf(
			"%3d %8s %5d %7.2f %7.2f %10.2f %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f %7d %7d %7d\n",
			item.Order,
			item.Company.Code,
			trade.Qty,
			trade.CompanyBatch.AvgPrice,
			trade.AvgPrice,
			(trade.RawResults - trade.TotalTax),
			trade.TotalTax,
			setFee,
			emlFee,
			brkFee,
			issFee,
			irrFee,
			trade.TaxGroup.Id,
			trade.Id,
			item.Id,
		)
	}
}

func (report *ConsoleReport) printTradeBatch() {
	items := report.invoice.Items

	var tradeBatch *entity.TradeBatch
	for _, item := range items {
		if item.Trade != nil {
			tradeBatch = item.Trade.TradeBatch
			break
		}
	}

	if tradeBatch == nil {
		return
	}

	fmt.Printf("Item.TradeBatch .. : \n")
	fmt.Printf(
		"%4s %8s %10s %10s %10s %8s %8s %8s\n",
		"Type",
		"Start Dt",
		"AccLoss",
		"Results",
		"Trades",
		"Taxes",
		"IRFEE",
		"TBatchId",
	)

	irfee := utils.GetTaxValueByGroup(
		tradeBatch.TaxGroup,
		constants.TaxTypes.IRFEE,
	)

	fmt.Printf(
		"%4s %8s %10.2f %10.2f %10.2f %8.2f %8.2f %8d\n",
		"SHR",
		tradeBatch.StartDate.Format("2006-01"),
		tradeBatch.Shr.AccLoss,
		tradeBatch.Shr.Results,
		tradeBatch.Shr.TotalTrade,
		tradeBatch.Shr.TotalTax,
		irfee,
		tradeBatch.Id,
	)

	fmt.Printf(
		"%4s %8s %10.2f %10.2f %10.2f %8.2f %8.2f %8d\n",
		"BDR",
		tradeBatch.StartDate.Format("2006-01"),
		tradeBatch.Bdr.AccLoss,
		tradeBatch.Bdr.Results,
		tradeBatch.Bdr.TotalTrade,
		tradeBatch.Bdr.TotalTax,
		irfee,
		tradeBatch.Id,
	)

	fmt.Printf(
		"%4s %8s %10.2f %10.2f %10.2f %8.2f %8.2f %8d\n",
		"ETF",
		tradeBatch.StartDate.Format("2006-01"),
		tradeBatch.Etf.AccLoss,
		tradeBatch.Etf.Results,
		tradeBatch.Etf.TotalTrade,
		tradeBatch.Etf.TotalTax,
		irfee,
		tradeBatch.Id,
	)
}

func (report *ConsoleReport) Run() error {
	invoice := report.invoice
	user := invoice.User

	fmt.Println("===== Report =====")
	fmt.Printf("User.name ........ : %s\n", user.UserName)
	fmt.Printf("User.Id .......... : %d\n", user.Id)
	fmt.Printf("User.UUID ........ : %s\n", user.UUID)
	fmt.Printf("Invoice.Id ....... : %d\n", invoice.Id)
	fmt.Printf("Invoice.Num ...... : %d\n", invoice.Number)
	fmt.Printf("Invoice.Date ..... : %s\n", invoice.MarketDate.Format(time.RFC3339))
	report.printTaxGroup("Invoice.TaxGroup . :", invoice.TaxGroup)
	report.printInvoiceItems()
	report.printItemBatch()
	report.printTrade()
	report.printTradeBatch()
	fmt.Println("==================")

	return nil
}
