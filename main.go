package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/patrickoliveros/bookings/internal/config"
	"github.com/patrickoliveros/bookings/internal/driver"
	"github.com/patrickoliveros/bookings/internal/helpers"
	"github.com/patrickoliveros/bookings/internal/mailer"
	"github.com/patrickoliveros/bookings/internal/pages"
	"github.com/patrickoliveros/bookings/internal/renders"
	"github.com/patrickoliveros/bookings/models"
	mail "github.com/xhit/go-simple-mail/v2"

	"github.com/alexedwards/scs/v2"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger
var appConnectionString string

func main() {
	parseApplicationFlags()

	db, err := runApplication()
	if err != nil {
		panic(err)
	}

	defer db.SQL.Close()
	defer close(app.MailChannel)

	log.Printf(">>> Starting application on port %s...", app.PortNumber)

	srv := &http.Server{
		Addr:    app.PortNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func printConfiguration() {
	fmt.Println("\n\n-------------------------------------------")
	fmt.Println("Application Settings")
	fmt.Println("-------------------------------------------")
	fmt.Println("In Production -", app.InProduction)
	fmt.Println("Use Cache -", app.UseCache)
	fmt.Println("Use Secure -", app.UseSecure)
	fmt.Println("\n\n-------------------------------------------")
	fmt.Println("Database Settings")
	fmt.Println("-------------------------------------------")
	fmt.Println("Connection String -", appConnectionString)
	fmt.Println("")
}

func registerModels() {
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(models.MailData{})
	gob.Register(map[string]int{})
}

func runApplication() (*driver.DB, error) {

	registerModels()

	// pingDatabase()
	db, err := tryConnectDatabase()

	// setup default non-overridable values
	setupDefaultAppConfig()
	printConfiguration()
	setupSession()
	setupDependencies()
	setupApplicationTemplates()
	setupMailServer()
	setupMailChannel()
	setupRepo(db)

	return db, err
}

func setupMailChannel() {
	mailChannel := make(chan models.MailData)
	app.MailChannel = mailChannel

	log.Println(">>> Starting mail listener...")
	mailer.ListenForMail()
}

func setupMailServer() {
	app.MailServer.Host = "192.168.50.146"
	app.MailServer.Port = 1025
	app.MailServer.KeepAlive = false
	app.MailServer.ConnectTimeout = 10 * time.Second
	app.MailServer.SendTimeout = 10 * time.Second

	mailer.NewMailer(&app)
}

func tryConnectDatabase() (*driver.DB, error) {
	log.Println(">>> Connecting to database...")
	db, err := driver.ConnectSQL(appConnectionString)

	if err != nil {
		log.Fatal(">>> Cannot connect to database! Dying...")
	}

	log.Println(">>> Connected to DB using driver!...")

	return db, nil
}

// getDbConnectionString() sets the hardcoded dbvalues
func getDbConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s",
		"your-db-server", 5433, "your-db-user", "your-db-password", "your-db-name")
}

func setDbConnectionString(host, dbname, port, user, password, ssl string) {
	appConnectionString = fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, ssl)
}

// func pingDatabase() {
// 	conn, err := sql.Open("pgx", getDbConnectionString())
// 	if err != nil {
// 		log.Fatal(">>> Cannot connect to db...")
// 	}
// 	defer conn.Close()

// 	log.Println(">>> Connected to db...")

// 	err = conn.Ping()
// 	if err != nil {
// 		log.Fatal(">>> Cannot ping db...")
// 	}
// }

func parseFlags(dbName, dbUser, dbPassword, dbServer, dbPort, dbSsl string) {
	var list []string

	if strings.TrimSpace(dbName) == "" {
		list = append(list, "dbname")
	}
	if strings.TrimSpace(dbUser) == "" {
		list = append(list, "dbuser")
	}
	if strings.TrimSpace(dbPassword) == "" {
		list = append(list, "dbpassword")
	}
	if strings.TrimSpace(dbServer) == "" {
		list = append(list, "dbserver")
	}
	if strings.TrimSpace(dbPort) == "" {
		list = append(list, "dbport")
	}

	if len(list) > 0 {
		fmt.Println("Missing required flags: ")
		for _, x := range list {
			fmt.Println("\t>>", x)
		}
		os.Exit(1)
	}

	setDbConnectionString(dbServer, dbName, dbPort, dbUser, dbPassword, dbSsl)
}

// parseApplicationFlags processes items from the command line
func parseApplicationFlags() {
	appConfig := flag.String("config", "default", "Config Source?")
	inProduction := flag.Bool("production", false, "Application is running in production?")
	useCache := flag.Bool("cache", true, "Use template cache?")

	// configurable dbSettings
	dbName := flag.String("dbname", "", "Database name?")                                // empty string means required
	dbUser := flag.String("dbuser", "", "Database username?")                            // empty string means required
	dbPassword := flag.String("dbpassword", "", "Database password?")                    // empty string means required
	dbServer := flag.String("dbserver", "", "Database server?")                          // empty string means required
	dbPort := flag.String("dbport", "", "Database port?")                                // empty string means required
	dbSSL := flag.String("dbssl", "disable", "Database SSL (disable, prefer, require)?") // empty string means required

	flag.Parse()

	// configurable appSettings
	app.InProduction = *inProduction
	app.UseCache = *useCache

	switch *appConfig {
	case "flags":
		parseFlags(*dbName, *dbUser, *dbPassword, *dbServer, *dbPort, *dbSSL)
	case "json":
	case "default":
		// will use json, but for this purpose we will set things by default
		// todo: implement reading config from an input file
		setupApplicationConfig()
	default:
		log.Fatal("Missing configuration source. Please specify if `config` values would be 'default', 'flags', or 'json'")
	}
}

func setupDefaultAppConfig() {
	app.PortNumber = ":8080"
	app.SiteSuffix = "Sample Go Web Application"
	app.MailServer = mail.NewSMTPClient()
	app.RootDirectory, _ = os.Getwd()
	app.UseSecure = false

	infoLog = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog
}

func setupApplicationConfig() {
	app.InProduction = false
	app.UseCache = false

	appConnectionString = getDbConnectionString()
}

func setupApplicationTemplates() {
	tc, err := renders.CreateAllTemplatesCache()
	helpers.HandleFatalError(err, "cannot create template cache")

	app.TemplateCache = tc

	renders.NewRenderer(&app)
}

// setupDependencies bootstraps references appConfig to other packages that needs it
func setupDependencies() {
	helpers.AppConfig = &app
}

func setupRepo(db *driver.DB) {
	repo := pages.NewRepo(&app, db)
	pages.NewPageHandlers(repo)
}

func setupSession() {
	session = scs.New()
	session.Lifetime = 23 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session
}
