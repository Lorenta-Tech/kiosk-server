package mail

import (
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/env"
	"github.com/resend/resend-go/v2"
)

type ResendClient struct {
	client *resend.Client
	from   string
}

func NewResendClient() (*ResendClient, error) {
	var (
		apiKey = env.GetString("RESEND_API_KEY","")
		from   = env.GetString("FROM_EMAIL","")
	)

	if apiKey == "" || from == "" {
		return nil, fmt.Errorf("missing resend environment variables")
	}

	client := resend.NewClient(apiKey)

	return &ResendClient{
		client: client,
		from:   from,
	}, nil
}

func (r *ResendClient) Send(to, subject, body string) error {

	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}