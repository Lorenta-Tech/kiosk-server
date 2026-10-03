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

// CalculateFilePrice calculates the printing price based on:
//   - selected pages
//   - number of copies
//   - page layout
//   - printing mode
//   - printing side
//
// Pricing rules:
//
// MONOCHROMATIC:
//
//	Single-side:
//	  Actual printed sides × ₹2
//
//	Double-side:
//	  Physical sheets × 2 sides × ₹1
//
// COLOR:
//
//	Single-side:
//	  Actual printed sides × ₹5
//
//	Double-side:
//	  Actual printed sides × ₹5
//
// Important:
// For monochromatic double-side printing, both sides of every
// physical sheet are charged, even if the final back side is empty.
//
// For color double-side printing, only sides containing actual
// printed pages are charged.
func CalculateFilePrice(
	numOfPages int,
	pageRanges []string,
	copies int,
	pageLayout int,
	printingMode string,
	printingSide string,
) (price float64, sheets int) {

	// Count only selected pages.
	selectedPages := countSelectedPages(
		pageRanges,
		numOfPages,
	)

	// If no page range is provided,
	// use all pages.
	if selectedPages == 0 {
		selectedPages = numOfPages
	}

	// Protect against invalid input.
	if selectedPages <= 0 ||
		copies <= 0 ||
		pageLayout <= 0 {
		return 0, 0
	}

	// ------------------------------------------------------------
	// Calculate pages that fit on one physical sheet.
	// ------------------------------------------------------------
	//
	// Single-side:
	//
	//   pageLayout pages per sheet
	//
	// Double-side:
	//
	//   pageLayout pages on front
	//   pageLayout pages on back
	//
	// Therefore:
	//
	//   pageLayout * 2 pages per sheet
	//
	pagesPerSheet := pageLayout

	if printingSide == "double_side" {
		pagesPerSheet = pageLayout * 2
	}

	// ------------------------------------------------------------
	// Calculate physical sheets required per copy.
	// ------------------------------------------------------------

	sheetsPerCopy := int(
		math.Ceil(
			float64(selectedPages) /
				float64(pagesPerSheet),
		),
	)

	// Total physical sheets for all copies.
	sheets = sheetsPerCopy * copies

	// ------------------------------------------------------------
	// Get cost per printed side.
	// ------------------------------------------------------------

	costPerSide := rateFor(
		printingMode,
		printingSide,
	)

	var totalSides int

	// ============================================================
	// SINGLE-SIDE
	// ============================================================

	if printingSide == "single_side" {

		// Only one side of each physical sheet is used.
		//
		// Example:
		//
		// pageLayout = 1
		//
		// 1 page -> 1 side
		// 2 pages -> 2 sides
		// 3 pages -> 3 sides
		//
		// pageLayout = 2
		//
		// 1 page -> 1 side
		// 2 pages -> 1 side
		// 3 pages -> 2 sides
		// 4 pages -> 2 sides
		totalSides = int(
			math.Ceil(
				float64(selectedPages) /
					float64(pageLayout),
			),
		)

	} else if printingMode == "monochromatic" {

		// ========================================================
		// MONOCHROMATIC + DOUBLE-SIDE
		// ========================================================
		//
		// Charge BOTH sides of every physical sheet.
		//
		// This accounts for physical paper consumption.
		//
		// pageLayout = 1:
		//
		// 1 page -> 1 sheet -> 2 sides -> ₹2
		// 2 pages -> 1 sheet -> 2 sides -> ₹2
		// 3 pages -> 2 sheets -> 4 sides -> ₹4
		// 4 pages -> 2 sheets -> 4 sides -> ₹4
		// 5 pages -> 3 sheets -> 6 sides -> ₹6
		//
		// pageLayout = 2:
		//
		// 1 page -> 1 sheet -> 2 sides -> ₹2
		// 2 pages -> 1 sheet -> 2 sides -> ₹2
		// 3 pages -> 1 sheet -> 2 sides -> ₹2
		// 4 pages -> 1 sheet -> 2 sides -> ₹2
		// 5 pages -> 2 sheets -> 4 sides -> ₹4
		//
		totalSides = sheetsPerCopy * 2

	} else {

		// ========================================================
		// COLOR + DOUBLE-SIDE
		// ========================================================
		//
		// Charge ONLY sides that actually contain printed pages.
		//
		// Do NOT charge the unused back side of the final sheet.
		//
		// pageLayout = 1:
		//
		// 1 page -> 1 printed side -> ₹5
		// 2 pages -> 2 printed sides -> ₹10
		// 3 pages -> 3 printed sides -> ₹15
		// 4 pages -> 4 printed sides -> ₹20
		// 5 pages -> 5 printed sides -> ₹25
		//
		// pageLayout = 2:
		//
		// 1 page -> 1 printed side -> ₹5
		// 2 pages -> 1 printed side -> ₹5
		// 3 pages -> 2 printed sides -> ₹10
		// 4 pages -> 2 printed sides -> ₹10
		// 5 pages -> 3 printed sides -> ₹15
		//
		totalSides = int(
			math.Ceil(
				float64(selectedPages) /
					float64(pageLayout),
			),
		)
	}

	// ------------------------------------------------------------
	// Apply number of copies.
	// ------------------------------------------------------------

	totalSides *= copies

	// ------------------------------------------------------------
	// Calculate final price.
	// ------------------------------------------------------------

	price = float64(totalSides) * costPerSide

	return price, sheets
}

// countSelectedPages calculates the number of unique selected
// pages from page ranges.
//
// Supported formats:
//
//	"1"       -> page 1
//	"1-3"     -> pages 1, 2, 3
//	"5-7"     -> pages 5, 6, 7
//
// Duplicate pages are counted only once.
func countSelectedPages(
	pageRanges []string,
	maxPages int,
) int {

	selected := make(map[int]bool)

	for _, r := range pageRanges {

		r = strings.TrimSpace(r)

		if r == "" {
			continue
		}

		// --------------------------------------------------------
		// Range such as "1-3"
		// --------------------------------------------------------

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

			// ----------------------------------------------------
			// Single page such as "5"
			// ----------------------------------------------------

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
	// Monochromatic single-side = ₹2 per side.
	if printingSide == "single_side" &&
		printingMode == "monochromatic" {

		return 2.00
	}

	// Normal rates.
	if rate, ok := rates[printingMode]; ok {
		return rate
	}

	// Default to monochromatic pricing.
	return rates["monochromatic"]
}
