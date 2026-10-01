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

func BuildDBDownTemplate(
	serviceName string,
	host string,
	downSince string,
	errorMsg string,
	downtime string,
	failures int,
) (EmailTemplate, error) {

	return EmailTemplate{
		Subject: fmt.Sprintf("[CRITICAL] Database Down - %s", serviceName),
		Body: fmt.Sprintf(`
				<h2>Database Unreachable</h2>
				<p>The database backing <strong>%s</strong> is not responding.</p>
				<p><strong>Host:</strong> %s</p>
				<p><strong>Down since:</strong> %s</p>
				<p><strong>Downtime:</strong> %s</p>
				<p><strong>Consecutive failures:</strong> %d</p>
				<p><strong>Error:</strong> %s</p>
				<p>All database dependent requests are failing until connectivity is restored.</p>
			`, serviceName, host, downSince, downtime, failures, errorMsg),
	}, nil
}

func BuildDBRecoveredTemplate(
	serviceName string,
	host string,
	downSince string,
	downtime string,
) (EmailTemplate, error) {

	return EmailTemplate{
		Subject: fmt.Sprintf("[RESOLVED] Database Recovered - %s", serviceName),
		Body: fmt.Sprintf(`
				<h2>Database Recovered</h2>
				<p>The database backing <strong>%s</strong> is healthy again.</p>
				<p><strong>Host:</strong> %s</p>
				<p><strong>Down since:</strong> %s</p>
				<p><strong>Total downtime:</strong> %s</p>
			`, serviceName, host, downSince, downtime),
	}, nil
}
