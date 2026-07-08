package mail

import (
	"fmt"
	"github.com/Lorenta-Tech/kiosk-server/internal/models"
)

type EmailTemplate struct {
	Subject string
	Body    string
}

func BuildPrinterErrorTemplate(
	errType models.PrinterError,
	printerID string,
	sessionID string,
) (EmailTemplate, error) {

	switch errType {

	case models.PrinterErrorPaperJam:
		return EmailTemplate{
			Subject: "Printer Alert - Paper Jam",
			Body: fmt.Sprintf(`
				<h2>Paper Jam Detected</h2>
				<p><strong>Printer:</strong> %s</p>
				<p><strong>Session:</strong> %s</p>
				<p>Please check the printer immediately.</p>
			`, printerID, sessionID),
		}, nil

	case models.PrinterErrorNoInk:
		return EmailTemplate{
			Subject: "Printer Alert - No Ink",
			Body: fmt.Sprintf(`
				<h2>Ink Empty</h2>
				<p><strong>Printer:</strong> %s</p>
				<p><strong>Session:</strong> %s</p>
				<p>Printer ink needs replacement.</p>
			`, printerID, sessionID),
		}, nil

	case models.PrinterErrorPaperOutOfBounds:
		return EmailTemplate{
			Subject: "Printer Alert - Paper Out Of Bounds",
			Body: fmt.Sprintf(`
				<h2>Paper Alignment Error</h2>
				<p><strong>Printer:</strong> %s</p>
				<p><strong>Session:</strong> %s</p>
				<p>The paper alignment is incorrect.</p>
			`, printerID, sessionID),
		}, nil
	}

	return EmailTemplate{}, fmt.Errorf("unsupported printer error")
}
