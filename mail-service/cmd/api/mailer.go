package main

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	simpleMail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (mail *Mail) SendSmtpMessage(message Message) error {
	if message.From == "" {
		message.From = mail.FromAddress
	}

	if message.FromName == "" {
		message.FromName = mail.FromName
	}

	data := map[string]any{
		"message": message.Data,
	}

	message.DataMap = data
	formattedMessage, err := mail.buildHtmlMessage(message)

	if err != nil {
		return err
	}

	plainMessage, err := mail.buildPlainTextMessage(message)

	if err != nil {
		return err
	}

	server := simpleMail.NewSMTPClient()

	server.Host = mail.Host
	server.Port = mail.Port
	server.Username = mail.Username
	server.Password = mail.Password
	server.Encryption = mail.getEncryption(mail.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()

	if err != nil {
		return err
	}

	email := simpleMail.NewMSG()

	email.SetFrom(message.From).AddTo(message.To).SetSubject(message.Subject)
	email.SetBody(simpleMail.TextPlain, plainMessage)
	email.AddAlternative(simpleMail.TextHTML, formattedMessage)

	if len(message.Attachments) > 0 {
		for _, attachment := range message.Attachments {
			email.AddAttachment(attachment)
		}
	}

	err = email.Send(smtpClient)

	if err != nil {
		return err
	}

	return nil
}

func (mail *Mail) buildHtmlMessage(message Message) (string, error) {
	templateToRender := "./templates/mail.html.gohtml"
	t, err := template.New("email-html").ParseFiles(templateToRender)

	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	if err = t.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = mail.inlineCss(formattedMessage)

	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (mail *Mail) buildPlainTextMessage(message Message) (string, error) {
	templateToRender := "./templates/mail.plain.gohtml"
	t, err := template.New("email-plain").ParseFiles(templateToRender)

	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	if err = t.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (mail *Mail) inlineCss(formattedMessage string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(formattedMessage, &options)

	if err != nil {
		return "", err
	}

	html, err := prem.Transform()

	if err != nil {
		return "", err
	}

	return html, nil
}

func (mail *Mail) getEncryption(encryption string) simpleMail.Encryption {
	switch encryption {
	case "tls":
		return simpleMail.EncryptionSTARTTLS
	case "ssl":
		return simpleMail.EncryptionSSLTLS
	case "none", "":
		return simpleMail.EncryptionNone
	default:
		return simpleMail.EncryptionSTARTTLS
	}
}
