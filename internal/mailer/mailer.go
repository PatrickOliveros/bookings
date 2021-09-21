package mailer

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/patrickoliveros/bookings/internal/config"
	"github.com/patrickoliveros/bookings/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

var app *config.AppConfig

func NewMailer(a *config.AppConfig) {
	app = a
}

func ListenForMail() {
	go func() {
		for {
			msg := <-app.MailChannel
			sendMessage(msg)
		}
	}()
}

func sendMessage(m models.MailData) {
	client, err := app.MailServer.Connect()
	if err != nil {
		log.Println(err)
	}

	log.Println(">> Trying to send a message...")

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	data, err := ioutil.ReadFile("./email-templates/basic.html")
	if err != nil {
		panic(err)
	}

	mailTemplate := string(data)
	msgToSend := strings.Replace(mailTemplate, "[%body%]", "hello, world", 1)
	email.SetBody(mail.TextHTML, msgToSend)

	err = email.Send(client)
	if err != nil {
		log.Println(">> Failed sending email:", err)
	} else {
		log.Println(">> Email sent!")
	}
}
