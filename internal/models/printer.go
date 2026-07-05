package models

type PrinterError string

const (
	PrinterErrorPaperOutOfBounds PrinterError = "paper_out_of_bounds"
	PrinterErrorPaperJam         PrinterError = "paper_jam"
	PrinterErrorNoInk            PrinterError = "no_ink"
)
