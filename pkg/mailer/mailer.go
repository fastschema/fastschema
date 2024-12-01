package mailer

import (
	"fmt"

	"github.com/fastschema/fastschema/fs"
)

func NewMailersFromConfig(config *fs.MailConfig) ([]fs.Mailer, error) {
	mailers := make([]fs.Mailer, len(config.Clients))
	for i, c := range config.Clients {
		name := fs.MapValue(c, "name", "")
		driver := fs.MapValue(c, "driver", "")
		if name == "" || driver == "" {
			return nil, fmt.Errorf("mailer %d: name and driver are required", i)
		}

		switch driver {
		case "smtp":
			smtpConfig := &SMTPConfig{
				DefaultSenderName: config.SenderName,
				DefaultSenderMail: config.SenderMail,
				Host:              fs.MapValue(c, "host", ""),
				Port:              int(fs.MapValue(c, "port", float64(587))),
				Username:          fs.MapValue(c, "username", ""),
				Password:          fs.MapValue(c, "password", ""),
				Insecure:          fs.MapValue(c, "insecure", false),
			}

			m, err := NewSMTPMailer(name, smtpConfig)
			if err != nil {
				return nil, fmt.Errorf("mailer %s: %w", name, err)
			}
			mailers[i] = m
		default:
			return nil, fmt.Errorf("mailer %s: unknown driver %s", name, driver)
		}
	}

	return mailers, nil
}
