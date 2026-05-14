package utils

import (
	"math"
	"strconv"
	"strings"
)

var rates = map[string]float64{
	"monochromatic": 1.00,
	"color":         4.00,
}

// CalculateFilePrice calculates price based on selected page ranges
func CalculateFilePrice(
	numOfPages int,
	pageRanges []string,
	copies int,
	pageLayout int,
	printingMode string,
	printingSide string,
) (price float64, sheets int) {

	// count only selected pages
	selectedPages := countSelectedPages(pageRanges, numOfPages)

	// fallback to all pages if no range provided
	if selectedPages == 0 {
		selectedPages = numOfPages
	}

	// how many pages fit in one sheet
	pagesPerSheet := pageLayout
	if printingSide == "double_side" {
		pagesPerSheet = pageLayout * 2
	}

	// total sheets per copy
	sheetsPerCopy := int(math.Ceil(float64(selectedPages) / float64(pagesPerSheet)))
	sheets = sheetsPerCopy * copies

	// cost per printed side
	costPerSide := rateFor(printingMode)

	var totalSides int

	if printingSide == "single_side" {

		// each sheet uses only one side
		totalSides = sheetsPerCopy

	} else {

		// double side logic
		fullSheets := selectedPages / (pageLayout * 2)
		remainingPages := selectedPages % (pageLayout * 2)

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

// countSelectedPages calculates total selected pages from ranges
func countSelectedPages(pageRanges []string, maxPages int) int {

	selected := make(map[int]bool)

	for _, r := range pageRanges {

		r = strings.TrimSpace(r)

		// range like 1-3
		if strings.Contains(r, "-") {

			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				continue
			}

			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

			if err1 != nil || err2 != nil {
				continue
			}

			if start > end {
				start, end = end, start
			}

			for i := start; i <= end; i++ {

				if i >= 1 && i <= maxPages {
					selected[i] = true
				}
			}

		} else {

			// single page like 5
			page, err := strconv.Atoi(r)
			if err != nil {
				continue
			}

			if page >= 1 && page <= maxPages {
				selected[page] = true
			}
		}
	}

	return len(selected)
}

func rateFor(printingMode string) float64 {
	if rate, ok := rates[printingMode]; ok {
		return rate
	}

	return rates["monochromatic"]
}