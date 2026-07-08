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
		apiKey = env.GetString("RESEND_API_KEY", "re_Y5tBWYjJ_2KYvgXRZsVkAvSNpkqnpQfRq")
		from   = env.GetString("FROM_EMAIL", "info@lorentatechnologies.com")
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
func (r *ResendClient) Send(
	to []string,
	subject string,
	body string,
) error {

	params := &resend.SendEmailRequest{
		From:    r.from,
		To:      to,
		Subject: subject,
		Html:    body,
	}

	_, err := r.client.Emails.Send(params)
	return err
}
