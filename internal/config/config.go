package config

import (
	"html/template"
	"log"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/patrickoliveros/bookings/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

type AppConfig struct {
	UseCache     bool
	UseSecure    bool
	InProduction bool

	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	PortNumber    string
	Session       *scs.SessionManager
	SiteSuffix    string
	MailChannel   chan models.MailData
	MailServer    *mail.SMTPServer
	RootDirectory string
}

type MailConfig struct {
	Host           string
	Port           int
	KeepAlive      bool
	ConnectTimeout time.Time
	SendTimeOut    time.Time
}
