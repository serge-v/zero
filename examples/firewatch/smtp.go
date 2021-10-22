package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
)

func sendEmail(from, to, subject, body string) error {
	subject = mime.QEncoding.Encode("utf-8", subject)

	msg := "From: " + from + "\n"
	msg += "To: " + to + "\n"
	msg += "Subject: " + subject + "\n\n"
	msg += body

	if err := smtpSend(from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func smtpSend(from string, to []string, message []byte) error {
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	if user == "" || pass == "" {
		return fmt.Errorf("invalid credentials")
	}

	auth := smtp.PlainAuth("", user, pass, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, message)
	if err != nil {
		return fmt.Errorf("sendmail: %w", err)
	}

	return nil
}

func formatMultipartMessage(plain, html io.Reader, zip string) (string, error) {
	var b bytes.Buffer

	mwr := multipart.NewWriter(&b)

	fmt.Fprintf(&b, "Content-Type: multipart/mixed; boundary=%s\n\n", mwr.Boundary())

	headers := make(textproto.MIMEHeader)
	headers.Add("Content-Type", "text/html")
	part, err := mwr.CreatePart(headers)
	if err != nil {
		return "", fmt.Errorf("create html part: %w", err)
	}
	fmt.Fprintln(part, html)

	//	headers = make(textproto.MIMEHeader)
	//	headers.Add("Content-Type", "text/plain")
	//	part, err = mwr.CreatePart(headers)
	//	if err != nil {
	//		return "", fmt.Errorf("create text part: %w", err)
	//	}
	//	fmt.Fprintln(part, plain)
	mwr.Close()

	return b.String(), nil
}
