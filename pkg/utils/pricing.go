package utils

import (
	"math"
	"strconv"
	"strings"
)

var rates = map[string]float64{
	"monochromatic": 1.00,
	"color":         5.00,
}

// CalculateFilePrice calculates the printing price based on
// selected pages, copies, page layout, printing mode and side.
//
// Pricing rules:
//
//	Monochromatic single-side = ₹2 per printed side
//	Monochromatic double-side = ₹1 per side
//	Color = ₹5 per side
//
// For double-side printing:
//
//	Every physical sheet is charged for BOTH sides,
//	even when the second side is unused.
//
//	Example:
//	1 page  -> 1 sheet -> 2 sides charged
//	2 pages -> 1 sheet -> 2 sides charged
//	3 pages -> 2 sheets -> 4 sides charged
//	4 pages -> 2 sheets -> 4 sides charged
func CalculateFilePrice(
	numOfPages int,
	pageRanges []string,
	copies int,
	pageLayout int,
	printingMode string,
	printingSide string,
) (price float64, sheets int) {

	// Count only selected pages.
	selectedPages := countSelectedPages(pageRanges, numOfPages)

	// If no page range is provided, use all pages.
	if selectedPages == 0 {
		selectedPages = numOfPages
	}

	// Protect against invalid input.
	if selectedPages <= 0 || copies <= 0 || pageLayout <= 0 {
		return 0, 0
	}

	// Calculate how many pages can be printed on one physical sheet.
	//
	// Single-side:
	//     pageLayout pages per sheet
	//
	// Double-side:
	//     pageLayout pages on front
	//     pageLayout pages on back
	//
	// Therefore:
	//     pageLayout * 2 pages per sheet.
	pagesPerSheet := pageLayout

	if printingSide == "double_side" {
		pagesPerSheet = pageLayout * 2
	}

	// Calculate physical sheets required for ONE copy.
	sheetsPerCopy := int(
		math.Ceil(
			float64(selectedPages) / float64(pagesPerSheet),
		),
	)

	// Total physical sheets across all copies.
	sheets = sheetsPerCopy * copies

	// Get price per side.
	costPerSide := rateFor(printingMode, printingSide)

	var totalSides int

	if printingSide == "single_side" {

		// Single-side printing:
		//
		// Every physical sheet only uses one side.
		//
		// Example with pageLayout = 1:
		//
		// 1 page -> 1 side
		// 2 pages -> 2 sides
		// 3 pages -> 3 sides
		//
		// Example with pageLayout = 2:
		//
		// 1 page -> 1 side
		// 2 pages -> 1 side
		// 3 pages -> 2 sides
		// 4 pages -> 2 sides.
		totalSides = int(
			math.Ceil(
				float64(selectedPages) / float64(pageLayout),
			),
		)

	} else {

		// Double-side printing:
		//
		// IMPORTANT:
		// Every physical sheet consumes BOTH sides.
		//
		// Therefore:
		//
		// 1 sheet  = 2 charged sides
		// 2 sheets = 4 charged sides
		// 3 sheets = 6 charged sides
		//
		// This also correctly handles partially filled sheets.
		totalSides = sheetsPerCopy * 2
	}

	// Apply number of copies.
	totalSides *= copies

	// Calculate final price.
	price = float64(totalSides) * costPerSide

	return price, sheets
}

// countSelectedPages calculates the number of unique selected pages
// from page ranges.
//
// Supported:
//
//	"1"   -> page 1
//	"1-3" -> pages 1, 2, 3
//
// Duplicate pages are counted only once.
func countSelectedPages(pageRanges []string, maxPages int) int {

	selected := make(map[int]bool)

	for _, r := range pageRanges {

		r = strings.TrimSpace(r)

		if r == "" {
			continue
		}

		// Range such as "1-3".
		if strings.Contains(r, "-") {

			parts := strings.Split(r, "-")

			if len(parts) != 2 {
				continue
			}

			start, err1 := strconv.Atoi(
				strings.TrimSpace(parts[0]),
			)

			end, err2 := strconv.Atoi(
				strings.TrimSpace(parts[1]),
			)

			if err1 != nil || err2 != nil {
				continue
			}

			// Support reversed ranges such as "5-3".
			if start > end {
				start, end = end, start
			}

			for i := start; i <= end; i++ {

				if i >= 1 && i <= maxPages {
					selected[i] = true
				}
			}

		} else {

			// Single page such as "5".
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

// rateFor returns the price per printed side.
func rateFor(
	printingMode string,
	printingSide string,
) float64 {

	// Special pricing:
	//
	// Monochromatic single-side = ₹2.
	if printingSide == "single_side" &&
		printingMode == "monochromatic" {

		return 2.00
	}

	// Normal rates.
	if rate, ok := rates[printingMode]; ok {
		return rate
	}

	// Default to monochromatic.
	return rates["monochromatic"]
}
