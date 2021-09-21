package models

import "github.com/patrickoliveros/bookings/internal/forms"

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{} // interface is used to hold unknown data or anything
	CSRFToken       string
	Flash           string
	Warning         string
	Error           string
	PageTitle       string
	Form            *forms.Form
	IsAuthenticated bool
	UserDisplayName string
	SiteSuffix      string
}
