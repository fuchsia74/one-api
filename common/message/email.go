package message

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
)

func shouldAuth() bool {
	return config.SMTPAccount != "" || config.SMTPToken != ""
}

func SendEmail(subject string, receiver string, content string) error {
	if receiver == "" {
		return errors.Errorf("receiver is empty")
	}
	if config.SMTPFrom == "" { // for compatibility
		config.SMTPFrom = config.SMTPAccount
	}
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	// Extract domain from SMTPFrom with fallback
	domain := "localhost"
	parts := strings.Split(config.SMTPFrom, "@")
	if len(parts) > 1 && parts[1] != "" {
		domain = parts[1]
	}

	// Generate a unique Message-ID
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return errors.Wrap(err, "failed to generate random bytes for Message-ID")
	}
	messageId := fmt.Sprintf("<%x@%s>", buf, domain)

	mail := fmt.Appendf(nil, "To: %s\r\n"+
		"From: %s<%s>\r\n"+
		"Subject: %s\r\n"+
		"Message-ID: %s\r\n"+ // add Message-ID header to avoid being treated as spam, RFC 5322
		"Date: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		receiver, config.SystemName, config.SMTPFrom, encodedSubject, messageId, time.Now().Format(time.RFC1123Z), content)

	auth := smtp.PlainAuth("", config.SMTPAccount, config.SMTPToken, config.SMTPServer)
	addr := net.JoinHostPort(config.SMTPServer, fmt.Sprintf("%d", config.SMTPPort))

	// Clean up recipient addresses
	receiverEmails := []string{}
	for email := range strings.SplitSeq(receiver, ";") {
		email = strings.TrimSpace(email)
		if email != "" {
			receiverEmails = append(receiverEmails, email)
		}
	}

	if len(receiverEmails) == 0 {
		return errors.New("no valid recipient email addresses")
	}

	// Use advanced client for port 465 (implicit TLS) or when auth is not needed
	// Also use advanced client for other ports to support STARTTLS
	var conn net.Conn
	var err error

	// Add connection timeout
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	if config.SMTPPort == 465 {
		// Port 465: implicit TLS (SMTPS)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: !config.ForceEmailTLSVerify,
			ServerName:         config.SMTPServer,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		// Other ports: plain connection first, will use STARTTLS if available
		conn, err = dialer.Dial("tcp", addr)
	}

	if err != nil {
		return errors.Wrap(err, "failed to connect to SMTP server")
	}

	client, err := smtp.NewClient(conn, config.SMTPServer)
	if err != nil {
		return errors.Wrap(err, "failed to create SMTP client")
	}
	defer client.Close()

	// For non-465 ports, try to use STARTTLS if supported
	if config.SMTPPort != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: !config.ForceEmailTLSVerify,
				ServerName:         config.SMTPServer,
			}
			if err = client.StartTLS(tlsConfig); err != nil {
				return errors.Wrap(err, "failed to start TLS")
			}
		}
	}

	// Authenticate if credentials are provided
	if shouldAuth() {
		if err = client.Auth(auth); err != nil {
			return errors.Wrap(err, "SMTP authentication failed")
		}
	}

	if err = client.Mail(config.SMTPFrom); err != nil {
		return errors.Wrap(err, "failed to set MAIL FROM")
	}

	for _, receiver := range receiverEmails {
		if err = client.Rcpt(receiver); err != nil {
			return errors.Wrapf(err, "failed to add recipient: %s", receiver)
		}
	}

	w, err := client.Data()
	if err != nil {
		return errors.Wrap(err, "failed to create message data writer")
	}

	if _, err = w.Write(mail); err != nil {
		return errors.Wrap(err, "failed to write email content")
	}

	if err = w.Close(); err != nil {
		return errors.Wrap(err, "failed to close message data writer")
	}

	return nil
}
