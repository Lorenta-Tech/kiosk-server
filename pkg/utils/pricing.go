package utils

var rates = map[string]map[string]float64{
	"monochromatic": {
		"single_side": 1.00,
		"double_side": 2.00,
	},
	"color": {
		"single_side": 5.00,
		"double_side": 10.00,
	},
}


func CalculateFilePrice(numOfPages, copies, pageLayout int, printingMode, printingSide string) (price float64, sheets int) {
	pagesPerSheet := pageLayout
	if printingSide == "double_side" {
		pagesPerSheet = pageLayout * 2
	}

	sheetsPerCopy := (numOfPages + pagesPerSheet - 1) / pagesPerSheet
	sheets = sheetsPerCopy * copies

	rate := rateFor(printingMode, printingSide)
	price = float64(sheets) * rate
	return price, sheets
}

func rateFor(printingMode, printingSide string) float64 {
	if modeRates, ok := rates[printingMode]; ok {
		if rate, ok := modeRates[printingSide]; ok {
			return rate
		}
	}
	return rates["monochromatic"]["single_side"]
}