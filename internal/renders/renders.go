package renders

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/justinas/nosurf"
	"github.com/patrickoliveros/bookings/internal/config"
	"github.com/patrickoliveros/bookings/internal/helpers"
	"github.com/patrickoliveros/bookings/models"
)

var functions = template.FuncMap{
	"humanDate":    helpers.HumanDate,
	"formatDate":   helpers.FormatDate,
	"iterate":      helpers.Iterate,
	"add":          helpers.Add,
	"calendarDate": helpers.CalendarDate,
}

var app *config.AppConfig

func NewRenderer(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	td.IsAuthenticated = app.Session.Exists(r.Context(), "user_id")
	td.SiteSuffix = app.SiteSuffix

	if app.Session.Exists(r.Context(), "user_id") {
		td.UserDisplayName = app.Session.Get(r.Context(), "user_displayname").(string)
	}

	return td
}

func RenderPageWithTemplate(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {

	var ts map[string]*template.Template

	var page = fmt.Sprintf("%s.page.html", tmpl)

	if app.UseCache {
		ts = app.TemplateCache
	} else {
		ts, _ = CreateAllTemplatesCache()
	}

	parsedTemplate, ok := ts[page]

	if !ok {
		log.Fatal("Could not get template")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = parsedTemplate.Execute(buf, td)
	_, err := buf.WriteTo(w)

	if err != nil {
		log.Fatal("Could not write template to browser")
	}
}

func CreateAllTemplatesCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := WalkFiles("./templates", ".page.html")

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			log.Fatal(err)
			return myCache, err
		}

		matches, err := WalkFiles("./templates/layouts", ".layout.html")
		if err != nil {
			log.Fatal(err)
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/layouts/*.layout.html")
			if err != nil {
				log.Fatal(err)
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}

func WalkFiles(rootDirectory, extension string) ([]string, error) {

	var list []string

	err := filepath.Walk(rootDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if fileNameWithoutExtension(path) == extension {
				list = append(list, path)
			}

			return nil
		})

	if err != nil {
		log.Println(err)
	}

	return list, err
}

func fileNameWithoutExtension(fileName string) string {
	if pos := strings.Index(fileName, `.`); pos != -1 {
		return fileName[pos:]
	}
	return fileName
}

func CreateTempAllTemplatesCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := WalkFiles("./templates", ".page.html")

	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			log.Fatal(err)
			return myCache, err
		}

		matches, err := WalkFiles("./templates/layouts", ".layout.html")
		if err != nil {
			log.Fatal(err)
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/layouts/*.layout.html")
			if err != nil {
				log.Fatal(err)
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
