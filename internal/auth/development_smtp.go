package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/plutoframework/pluto-auth/methods/password"
)

// DevelopmentSMTP sends ONLY to a numeric loopback capture service. TLS and
// SMTP AUTH are deliberately absent on this disposable local connection. This
// is not a production mail adapter. It never logs notifications/provider errors.
type DevelopmentSMTP struct{ address string }

func NewDevelopmentSMTP(address string) (*DevelopmentSMTP, error) {
	host, port, err := net.SplitHostPort(address)
	ip := net.ParseIP(host)

	number, portErr := strconv.Atoi(port)
	if err != nil || ip == nil || !ip.IsLoopback() || portErr != nil ||
		number < 1 || number > 65535 {
		return nil, errors.New(
			"development SMTP requires numeric loopback host:port",
		)
	}

	return &DevelopmentSMTP{address: address}, nil
}

func (s *DevelopmentSMTP) Send(
	ctx context.Context,
	n password.Notification,
) error {
	canonical, err := password.CanonicalEmail(n.Email)
	if err != nil || canonical != n.Email ||
		!strings.HasSuffix(n.Email, "@example.test") ||
		(n.Purpose != password.VerifyEmail && n.Purpose != password.RecoverPassword) {
		return password.ErrDelivery
	}

	// Validate the library's token before formatting it into the local message.
	if _, err := password.ChallengeDigest(n.Purpose, n.Token); err != nil {
		return password.ErrDelivery
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", s.address)
	if err != nil {
		return password.ErrDelivery
	}
	defer connection.Close()

	stop := context.AfterFunc(ctx, func() { _ = connection.Close() })
	defer stop()

	deadline, _ := ctx.Deadline()
	if err := connection.SetDeadline(deadline); err != nil {
		return password.ErrDelivery
	}

	client, err := smtp.NewClient(connection, "localhost")
	if err != nil {
		return password.ErrDelivery
	}
	defer client.Close()

	if err := client.Mail("pluto@example.test"); err != nil {
		return password.ErrDelivery
	}

	if err := client.Rcpt(n.Email); err != nil {
		return password.ErrDelivery
	}

	body, err := client.Data()
	if err != nil {
		return password.ErrDelivery
	}

	// No URL is built from Host, forwarded headers or untrusted return input.
	_, err = fmt.Fprintf(
		body,
		"From: pluto@example.test\r\nTo: %s\r\n"+
			"Subject: Pluto DEVELOPMENT %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n"+
			"Synthetic local walkthrough only. Explicitly POST the token from this message.\r\n"+
			"Purpose: %s\r\nToken: %s\r\nExpires: %s\r\n",
		n.Email,
		n.Purpose,
		n.Purpose,
		n.Token,
		n.ExpiresAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return password.ErrDelivery
	}

	if err := body.Close(); err != nil {
		return password.ErrDelivery
	}

	if err := client.Quit(); err != nil {
		return password.ErrDelivery
	}

	return ctx.Err()
}
