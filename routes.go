package main

import (
	"net/http"

	"github.com/patrickoliveros/bookings/api"
	"github.com/patrickoliveros/bookings/internal/config"
	"github.com/patrickoliveros/bookings/internal/pages"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	setMiddlewares(mux)
	getRequests(mux)
	postRequests(mux)

	setSecurePages(mux)
	setAPIEndpoints(mux)

	enableStaticFiles(mux)

	return mux
}

func setMiddlewares(mux *chi.Mux) {
	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)
}

func getRequests(mux *chi.Mux) {
	mux.Get("/", pages.Repo.HomePage)
	mux.Get("/about", pages.Repo.AboutPage)

	mux.Get("/reservations", pages.Repo.ReservationsPage)

	mux.Get("/make-reservation", pages.Repo.MakeReservationPage)
	mux.Get("/reservation-summary", pages.Repo.SummaryMakeReservationPage)
	mux.Get("/contact", pages.Repo.ContactPage)

	mux.Get("/rooms/generals-quarters", pages.Repo.RoomsGeneral)
	mux.Get("/rooms/majors-suite", pages.Repo.RoomsMajor)
	mux.Get("/choose-room/{id}", pages.Repo.ChooseRoom)
	mux.Get("/book-room", pages.Repo.BookRoom)

	mux.Get("/login", pages.Repo.LoginPage)
	mux.Get("/logout", pages.Repo.LogoutPage)

	mux.Get("/favicon.ico", pages.Repo.Favicon)
}

func postRequests(mux *chi.Mux) {
	mux.Post("/reservations", pages.Repo.PostReservationsPage)
	mux.Post("/make-reservation", pages.Repo.PostMakeReservationPage)
	mux.Post("/availability", pages.Repo.AvailabilityReservationsPage)

	mux.Post("/login", pages.Repo.PostLoginPage)
}

func setSecurePages(mux *chi.Mux) {
	mux.Route("/admin", func(mux chi.Router) {
		// mux.Use(helpers.Auth)

		adminGetPages(mux)
		adminPostPages(mux)
	})
}

func setAPIEndpoints(mux *chi.Mux) {
	mux.Route("/api", func(mux chi.Router) {
		// mux.Use(helpers.Auth)

		// adminGetPages(mux)
		// adminPostPages(mux)
		mux.Get("/articles", api.GetArticles)
	})
}

func adminGetPages(mux chi.Router) {
	mux.Get("/dashboard", pages.Repo.AdminDashBoard)
	mux.Get("/reservations-new", pages.Repo.AdminReservationsNew)
	mux.Get("/reservations-all", pages.Repo.AdminReservationsAll)
	mux.Get("/reservations-calendar", pages.Repo.AdminReservationsCalendar)
	mux.Get("/reservation/{id}", pages.Repo.AdminReservationsById)
	mux.Get("/process-reservation/{id}", pages.Repo.AdminProcessReservation)
}

func adminPostPages(mux chi.Router) {
	mux.Post("/dashboard", pages.Repo.AdminDashBoard)
	mux.Post("/reservations-new", pages.Repo.AdminReservationsNew)
	mux.Post("/reservations-all", pages.Repo.AdminReservationsAll)
	mux.Post("/reservations-calendar", pages.Repo.AdminPostReservationsCalendar)
	mux.Post("/reservation/{id}", pages.Repo.AdminPostReservationsById)
}

func enableStaticFiles(mux *chi.Mux) {
	mux.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
}
