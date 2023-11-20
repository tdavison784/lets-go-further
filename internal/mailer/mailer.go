package mailer

import (
	"bytes"
	"embed"
	"github.com/go-mail/mail/v2"
	"html/template"
	"time"
)

// Below we declare a new variable with the type embed.FS (embedded file system)
// to hold our email templates. This has a comment directive in the format: `//go:embed <path>`
// IMMEDIATELY ABOVE IT, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.

//go:embed "templates"
var templateFS embed.FS

// Mailer struct which contains a mail.Dialer instance (used to connect to an SMTP server)
// and the sender information for your emails (name, and address you want the email to be from,
// such as "Alice Smith <alice@example.com>")
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// Init a new mail.Dialer instance with the given SMTP server settings.
	// We also configure this to use a 5-second timeout whenever we send an email.
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	// return a Mailer instance containing the dialer and sender information
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send method on the Mailer type. This takes the recipient email address
// as the first parameter, the name of the file containing the templates,
// and any dynamic data for the templates as an any parameter
func (m Mailer) Send(recipient, templateFile string, data any) error {

	//  Use the ParseFS() method to parse the required template file from the embedded
	// file system
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Follow the same pattern to execute the "plainBody" template and store the result
	// in the plainBody variable
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// And likewise with the "htmlBody" template
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a new mail.Message instance.
	// Then we use he SetHeader() method to set the email recipient, sender, and subject
	// headers, the SetBody() method is set to the plain-text body, and the AddAlternative()
	// method to set the HTML body. It's important to note that the AddAlternative() should
	// always be called *after* SetBody()
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Call the DialAndSend() method on the dialer, passing in the message to send
	// This opens a connection to the SMTP server, sends the message, then closes the
	// connection. If there is a timeout, it will return a "dial tcp: i/o timeout" error.
	// adding simple retry logic here to try 3 times to send the message and sleep for
	// five seconds in-between efforts
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}
