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
)

// loginAuth implements the LOGIN authentication mechanism
type loginAuth struct {
	username, password string
}

// LoginAuth returns an Auth that implements the LOGIN authentication mechanism
func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:", "username:":
			return []byte(a.username), nil
		case "Password:", "password:":
			return []byte(a.password), nil
		default:
			return nil, errors.Errorf("unexpected server challenge: %s", string(fromServer))
		}
	}
	return nil, nil
}

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
		"Message-ID: %s\r\n"+
		"Date: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		receiver, config.SystemName, config.SMTPFrom, encodedSubject, messageId, time.Now().Format(time.RFC1123Z), content)

	// Use LOGIN auth instead of PLAIN auth
	auth := LoginAuth(config.SMTPAccount, config.SMTPToken)
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
