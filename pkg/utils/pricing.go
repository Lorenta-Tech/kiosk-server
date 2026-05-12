package utils

import "math"

var rates = map[string]float64{
	"monochromatic": 1.00,
	"color":         4.00,
}

func CalculateFilePrice(
	numOfPages,
	copies,
	pageLayout int,
	printingMode,
	printingSide string,
) (price float64, sheets int) {

	// how many pages fit in one sheet
	pagesPerSheet := pageLayout
	if printingSide == "double_side" {
		pagesPerSheet = pageLayout * 2
	}

	// total sheets per copy
	sheetsPerCopy := int(math.Ceil(float64(numOfPages) / float64(pagesPerSheet)))
	sheets = sheetsPerCopy * copies

	// cost per printed side
	costPerSide := rateFor(printingMode)

	var totalSides int

	if printingSide == "single_side" {

		// each sheet uses only one side
		totalSides = sheetsPerCopy

	} else {

		// double side logic
		fullSheets := numOfPages / (pageLayout * 2)
		remainingPages := numOfPages % (pageLayout * 2)

		if remainingPages == 0 {

			totalSides = fullSheets * 2

		} else if remainingPages <= pageLayout {

			// remaining pages fit on one side
			totalSides = fullSheets*2 + 1

		} else {

			// remaining pages need both sides
			totalSides = fullSheets*2 + 2
		}
	}

	totalSides = totalSides * copies
	price = float64(totalSides) * costPerSide

	return price, sheets
}

func rateFor(printingMode string) float64 {
	if rate, ok := rates[printingMode]; ok {
		return rate
	}

	return rates["monochromatic"]
}
