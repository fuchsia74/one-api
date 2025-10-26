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

	// ---- Unified advanced client with STARTTLS support ----
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !config.ForceEmailTLSVerify,
		ServerName:         config.SMTPServer,
	}

	var (
		conn net.Conn
		err  error
	)

	// 465: implicit TLS, others: plain first (will try STARTTLS below)
	if config.SMTPPort == 465 {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
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

	// For non-465 ports, try STARTTLS if the server advertises it
	if config.SMTPPort != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(tlsConfig); err != nil {
				return errors.Wrap(err, "failed to start STARTTLS")
			}
			logger.Logger.Debug("SMTP connection upgraded via STARTTLS")
		} else {
			logger.Logger.Warn("SMTP server does not advertise STARTTLS; proceeding without TLS")
		}
	}

	// Authenticate if configured
	if shouldAuth() {
		if err = client.Auth(auth); err != nil {
			return errors.Wrap(err, "SMTP authentication failed")
		}
	}

	if err = client.Mail(config.SMTPFrom); err != nil {
		return errors.Wrap(err, "failed to set MAIL FROM")
	}
	for _, rcpt := range receiverEmails {
		if err = client.Rcpt(rcpt); err != nil {
			return errors.Wrapf(err, "failed to add recipient: %s", rcpt)
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

	// Try polite quit; ignore error since邮件已发送成功
	_ = client.Quit()
	return nil
}
