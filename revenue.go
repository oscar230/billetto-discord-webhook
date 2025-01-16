package main

import (
	"fmt"
	"sort"
)

func GetRevenue(priceList []Price, currency string, sales int) string {
	// Step 1: Sort the price list by price ascending
	sort.Slice(priceList, func(i, j int) bool {
		return priceList[i].Price < priceList[j].Price
	})

	totalSales := float32(0)
	totalSalesText := "Prislista;"
	for _, entry := range priceList {
		if sales <= 0 {
			break
		}

		// Determine how much we can sell at this price
		sellAmount := min(sales, entry.Amount)
		totalSales += float32(sellAmount) * entry.Price
		totalSalesText += fmt.Sprintf("\n* %d st à %.2f %s/st", sellAmount, entry.Price, currency)
		if sellAmount == entry.Amount {
			totalSalesText += " (slutsålt)"
		} else {
			totalSalesText += fmt.Sprintf(" (%d st kvar)", entry.Amount-sellAmount)
		}
		sales -= sellAmount
	}
	totalSalesText = fmt.Sprintf("%.2f %s\n%s\n*Priset baseras på antagandet att alla köper den billigaste biljetten tillgänglig.*", totalSales, currency, totalSalesText)

	if sales > 0 {
		totalSalesText += fmt.Sprintf("\n⚠️ **Fel**: Det har sålts %d fler biljetter än vad som är definierade i prislistan! Inget är fel på Billetto, bara denna övervakare.", sales)
	}

	return totalSalesText
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
