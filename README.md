# Golang  - Bookings and Reservations Web Application
This is a sample web application that performs CRUD operations using Go and PostgreSQL. 
<p>&nbsp;</p>

## Key Technologies
* Go 1.17
* PostgreSQL
* HTML

<p>&nbsp;</p>

### Go Libraries used
* Session Management: [SCS: HTTP Session Management](https://pkg.go.dev/github.com/alexedwards/scs/v2@v2.4.0)
* Forms Validator: [Go Validator](github.com/asaskevich/govalidator)
* Routing: [Chi Router](github.com/go-chi/chi/v5)
* CSRF Protection: [No Surf](github.com/justinas/nosurf)
* PostgreSQL Driver [pgconn](github.com/jackc/pgconn)
* Mail Server [Go Simple Mail](github.com/xhit/go-simple-mail/v2)

<p>&nbsp;</p>

### Client Side Libraries Used
* Vanilla JS Datepicker
* Notie
* Sweet Aleart    
* Bootstrap 5

<p>&nbsp;</p>

## Running the Application
The application is designed to have default values filled up in the main.go file, but also accepts parameters from the command line using flags. 

<p>&nbsp;</p>
By default, you can run the following:

```
go run . 
```

The application will run based on the hardcoded values in the code. Alternatively, you can use the following to override the hardcoded *database* settings. 

```
go run . -config "flags" -dbname "<your-db-name>" -dbuser "<your-db-user>" -dbpassword "<your-db-password>" -dbserver "<your-db-server>" -dbport "<your-db-port>"
```

Additional flags can be used to set if application is in production mode or should the application should use a template cache.

```
-production "false" -cache "false"
```

<p>&nbsp;</p>

This application is based off the Udemy [Course](https://www.udemy.com/course/building-modern-web-applications-with-go/) but uses a different application structure and options. 
<p>&nbsp;</p>

### What's modified from the course?
---
* Application run checking the database from the start
* Page templates are arranged in separate folders for better organization. Individual pages are separated from the templates.
* Tests are missing from this version as it needs to be re-written because of the different application structure.