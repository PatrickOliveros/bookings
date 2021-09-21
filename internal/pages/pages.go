package pages

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/patrickoliveros/bookings/internal/config"
	"github.com/patrickoliveros/bookings/internal/driver"
	"github.com/patrickoliveros/bookings/internal/forms"
	"github.com/patrickoliveros/bookings/internal/helpers"
	"github.com/patrickoliveros/bookings/internal/logging"
	"github.com/patrickoliveros/bookings/internal/renders"
	"github.com/patrickoliveros/bookings/internal/repository"
	"github.com/patrickoliveros/bookings/internal/repository/dbrepo"
	"github.com/patrickoliveros/bookings/models"
)

//region repo

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

func (r *Repository) AddSessionError(req *http.Request, message string) {
	r.App.Session.Put(req.Context(), "error", strings.Title(message))
}

func (r *Repository) AddFlashMessage(req *http.Request, message string) {
	r.App.Session.Put(req.Context(), "flash", strings.Title(message))
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostGresRepo(db.SQL, a),
	}
}

func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewPageHandlers(r *Repository) {
	Repo = r
}

//endregion

//region static pages
func (m *Repository) HomePage(w http.ResponseWriter, r *http.Request) {
	pageTemplate := "home"

	renders.RenderPageWithTemplate(w, r, pageTemplate, &models.TemplateData{
		PageTitle: strings.Title(pageTemplate),
	})
}

func (m *Repository) Favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/favicon.ico")
}

func (m *Repository) AboutPage(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["remote_ip"] = r.RemoteAddr
	stringMap["local_ip"] = string(helpers.GetOutboundIP())

	pageTemplate := "about"

	renders.RenderPageWithTemplate(w, r, pageTemplate, &models.TemplateData{
		PageTitle: strings.Title(pageTemplate),
		StringMap: stringMap,
	})
}

func (m *Repository) ContactPage(w http.ResponseWriter, r *http.Request) {
	pageTemplate := "contact"

	renders.RenderPageWithTemplate(w, r, pageTemplate, &models.TemplateData{
		PageTitle: strings.Title(pageTemplate),
	})
}

func (m *Repository) LoginPage(w http.ResponseWriter, r *http.Request) {

	if m.App.Session.Exists(r.Context(), "user_id") {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	pageTemplate := "login"

	renders.RenderPageWithTemplate(w, r, pageTemplate, &models.TemplateData{
		PageTitle: strings.Title(pageTemplate),
		Form:      forms.New(nil),
	})
}

func (m *Repository) PostLoginPage(w http.ResponseWriter, r *http.Request) {

	pageTemplate := "login"

	_ = m.App.Session.RenewToken(r.Context())
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		renders.RenderPageWithTemplate(w, r, pageTemplate, &models.TemplateData{
			PageTitle: strings.Title(pageTemplate),
			Form:      form,
		})
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, displayName, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.AddSessionError(r, "Invalid login credentials")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "user_displayname", displayName)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) LogoutPage(w http.ResponseWriter, r *http.Request) {

	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	m.App.Session.Put(r.Context(), "flash", "Logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//endregion

func (m *Repository) RoomsGeneral(w http.ResponseWriter, r *http.Request) {
	renders.RenderPageWithTemplate(w, r, "generals-quarters", &models.TemplateData{
		PageTitle: "General's Quarters",
	})
}

func (m *Repository) RoomsMajor(w http.ResponseWriter, r *http.Request) {
	renders.RenderPageWithTemplate(w, r, "majors-suite", &models.TemplateData{
		PageTitle: "Major's Suite",
	})
}

func (m *Repository) ReservationsPage(w http.ResponseWriter, r *http.Request) {
	pageTemplate := "reservations"

	renders.RenderPageWithTemplate(w, r, pageTemplate, &models.TemplateData{
		PageTitle: strings.Title(pageTemplate),
	})
}

func (m *Repository) PostReservationsPage(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		m.AddSessionError(r, "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	start := r.Form.Get("start_date")
	end := r.Form.Get("end_date")

	startDate, _, err := helpers.HandleDate(start)
	if err != nil {
		m.AddSessionError(r, "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, _, err := helpers.HandleDate(end)
	if err != nil {
		m.AddSessionError(r, "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := m.DB.SearchAvailabilityByDates(startDate, endDate)
	if err != nil {
		m.AddSessionError(r, "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		// no availability
		m.AddSessionError(r, "No availability")
		http.Redirect(w, r, "/reservations", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	renders.RenderPageWithTemplate(w, r, "choose-room", &models.TemplateData{
		PageTitle: "Choose Room",
		Data:      data,
	})
}

func (m *Repository) AvailabilityReservationsPage(w http.ResponseWriter, r *http.Request) {
	// need to parse request body
	err := r.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate json
		resp := models.JsonReservationResponse{
			OK:      false,
			Message: "Internal server error",
		}

		outputJson(w, resp)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	startDate, _, _ := helpers.HandleDate(sd)
	endDate, _, _ := helpers.HandleDate(ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByDatesByRoom(startDate, endDate, roomID)
	if err != nil {
		// got a database error, so return appropriate json
		resp := models.JsonReservationResponse{
			OK:      false,
			Message: "Error querying database",
		}

		outputJson(w, resp)
		return
	}

	resp := models.JsonReservationResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	outputJson(w, resp)
}

func (m *Repository) MakeReservationPage(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.AddSessionError(r, "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.AddSessionError(r, "can't find room!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	pageTemplate := "make-reservation"

	renders.RenderPageWithTemplate(w, r, pageTemplate, &models.TemplateData{
		Data:      data,
		PageTitle: helpers.SanitizeString(pageTemplate),
		Form:      forms.New(nil),
		StringMap: stringMap,
	})
}

func (m *Repository) PostMakeReservationPage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		logging.ServerError(w, err)
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	startDate, _, err := helpers.HandleDate(sd)
	if err != nil {
		log.Fatal("Cannot parse start date")
	}

	endDate, _, err := helpers.HandleDate(ed)
	if err != nil {
		log.Fatal("Cannot parse end date")
	}

	roomID := 1
	// todo: fix this one
	// roomIDx, err := strconv.Atoi(r.Form.Get("room_id"))
	// if err != nil {
	// 	m.AddSessionError(r, "invalid data!")
	// 	http.Redirect(w, r, "/", http.StatusSeeOther)
	// 	return
	// }

	// room, err := m.DB.GetRoomByID(roomIDx)
	// if err != nil {
	// 	m.AddSessionError(r, "invalid data!")
	// 	http.Redirect(w, r, "/", http.StatusSeeOther)
	// 	return
	// }

	reservation := models.Reservation{
		FirstName:        r.Form.Get("first_name"),
		LastName:         r.Form.Get("last_name"),
		Phone:            r.Form.Get("phone"),
		Email:            r.Form.Get("email"),
		ReadableRoomName: r.Form.Get("room_name"),
		StartDate:        startDate,
		EndDate:          endDate,
		RoomID:           roomID,
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 5)
	form.MinLength("last_name", 5)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		renders.RenderPageWithTemplate(w, r, "make-reservation", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	reservation.Reference = helpers.GenerateGuid()

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		log.Println(err)
		return
	}

	// send notification to guest
	msg := models.MailData{
		To:   reservation.Email,
		From: "there@domain.com",
		Subject: fmt.Sprintf("Reservation Confirmation #%s - %s, %s",
			reservation.Reference, reservation.LastName, reservation.FirstName),
		Content: "hello! <strong>world</strong>!",
	}

	m.App.MailChannel <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) SummaryMakeReservationPage(w http.ResponseWriter, r *http.Request) {

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.AddSessionError(r, "Can't get reservation from session.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	renders.RenderPageWithTemplate(w, r, "reservation-summary", &models.TemplateData{
		PageTitle: "Major's Suite",
		Data:      data,
	})
}

// ChooseRoom displays list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.AddSessionError(r, "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.AddSessionError(r, "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL parameters, builds a sessional variable, and takes user to make res screen
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	startDate, _, _ := helpers.HandleDate(sd)
	endDate, _, _ := helpers.HandleDate(ed)

	var res models.Reservation

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.AddSessionError(r, "Can't get room from db!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func outputJson(w http.ResponseWriter, resp models.JsonReservationResponse) {
	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

/*
https://stackoverflow.com/questions/41109708/cant-define-receiver-from-another-package-in-go

You can only define methods on a type defined in that same package.
Your DB type, in this case, is defined within your dbconfig package, so your entity package can't define methods on it.

In this case, your options are to make GetContracts a function instead of a method and hand it the *dbconfig.DB as an argument,
or to invert the dependency by importing your entity package in dbconfig and
write GetContracts there (as a method or function, works either way).

This second one may actually be the better option, because, from a design perspective,
it breaks abstraction to have packages other than your database package creating SQL query strings.
*/

func (m *Repository) AdminDashBoard(w http.ResponseWriter, r *http.Request) {

	renders.RenderPageWithTemplate(w, r, "admin", &models.TemplateData{
		PageTitle: "Admin Home",
	})
}

func (m *Repository) AdminReservationsNew(w http.ResponseWriter, r *http.Request) {

	reservations, err := m.DB.GetAllReservations(true)
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	pageTitle := "reservations-new"
	renders.RenderPageWithTemplate(w, r, pageTitle, &models.TemplateData{
		PageTitle: helpers.SanitizeString(pageTitle),
		Data:      data,
	})
}

func (m *Repository) AdminReservationsAll(w http.ResponseWriter, r *http.Request) {

	reservations, err := m.DB.GetAllReservations(false)
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	pageTitle := "reservations"
	renders.RenderPageWithTemplate(w, r, "reservations-all", &models.TemplateData{
		PageTitle: helpers.SanitizeString(pageTitle),
		Data:      data,
	})
}

func (m *Repository) AdminReservationsById(w http.ResponseWriter, r *http.Request) {

	exploded := strings.Split(r.RequestURI, "/")
	// 0 -
	// 1 - admin
	// 2 - reservation
	// 3 - id

	reservationId, err := strconv.Atoi(exploded[3])
	if err != nil {
		m.AddSessionError(r, "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservations, err := m.DB.GetReservationById(reservationId)
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	page := "reservations-show"
	pageTitle := fmt.Sprintf("Reservation #%s", reservations.Reference)

	renders.RenderPageWithTemplate(w, r, page, &models.TemplateData{
		PageTitle: pageTitle,
		Data:      data,
	})
}

func (m *Repository) AdminPostReservationsById(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.IsEmail("email")

	exploded := strings.Split(r.RequestURI, "/")
	// 0 -
	// 1 - admin
	// 2 - reservation
	// 3 - id

	reservationId, err := strconv.Atoi(exploded[3])
	if err != nil {
		m.AddSessionError(r, "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if !form.Valid() {
		m.AddSessionError(r, "invalid form values")
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusTemporaryRedirect)
		return
	}

	reservations, err := m.DB.GetReservationById(reservationId)
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	// update the reservation
	reservations.FirstName = r.Form.Get("first_name")
	reservations.LastName = r.Form.Get("last_name")
	reservations.Email = r.Form.Get("email")
	reservations.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(reservations)
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	page := "reservations-show"
	pageTitle := fmt.Sprintf("Reservation #%s", reservations.Reference)

	m.App.Session.Put(r.Context(), "flash", "Updated Reservation!")

	renders.RenderPageWithTemplate(w, r, page, &models.TemplateData{
		PageTitle: pageTitle,
		Data:      data,
	})
}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {

	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	lastMonth, lastMonthYear,
		thisMonth, thisMonthYear,
		nextMonth, nextMonthYear := helpers.GetCalendarDates(now)

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear
	stringMap["this_month"] = thisMonth
	stringMap["this_month_year"] = thisMonthYear

	firstOfMonth, lastOfMonth := helpers.GetMonthBoundaries(now)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.GetAllRooms()
	if err != nil {
		logging.ServerError(w, err)
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		// we need to get all the restrictions
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			logging.ServerError(w, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				for d := y.StartDate; !d.After(y.EndDate); d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else {
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}

		}

		reservationMapKey := fmt.Sprintf("reservation_map_%d", x.ID)
		blockMapKey := fmt.Sprintf("block_map_%d", x.ID)

		data[reservationMapKey] = reservationMap
		data[blockMapKey] = blockMap

		m.App.Session.Put(r.Context(), blockMapKey, blockMap)
	}

	pageTitle := "reservations-calendar"
	renders.RenderPageWithTemplate(w, r, pageTitle, &models.TemplateData{
		PageTitle: helpers.SanitizeString(pageTitle),
		StringMap: stringMap,
		IntMap:    intMap,
		Data:      data,
	})
}

func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		logging.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	// process blocks
	rooms, err := m.DB.GetAllRooms()
	if err != nil {
		logging.ServerError(w, err)
	}

	form := forms.New(r.PostForm)

	for _, x := range rooms {
		// Get the block map from the session.
		// Loop through the entire map,
		// if we have an entry in the map that's not in posted data, and restriction id > 0, remove block

		currentMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)

		var idsForDeletion []string

		for name, value := range currentMap {
			// ok will be false if the value is not in the map
			if val, ok := currentMap[name]; ok {
				// only pay attention to values > 0, and that are not in the form post
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						// delete the restriction by id
						idsForDeletion = append(idsForDeletion, strconv.Itoa(value))
					}
				}
			}
		}

		if len(idsForDeletion) > 0 {
			// now delete the string
			err := m.DB.DeleteBlocksForRoom(x.ID, strings.Join(idsForDeletion, ", "))
			if err != nil {
				logging.ServerError(w, err)
				return
			}
		}
	}

	// now handle new blocks
	for name := range r.PostForm {
		// if the checkbox is not checked, the form won't be passed
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			// insert a new block
			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				logging.ServerError(w, err)
				return
			}
		}
	}

	m.AddFlashMessage(r, "changes saved!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}

func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	_ = m.DB.MarkProcessedReservation(id, 1)

	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed!")
	http.Redirect(w, r, "/admin/reservations-all", http.StatusSeeOther)
}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	_ = m.DB.DeleteReservation(id)

	m.App.Session.Put(r.Context(), "flash", "Reservation deleted!")
	http.Redirect(w, r, "/admin/reservations-all", http.StatusSeeOther)
}
