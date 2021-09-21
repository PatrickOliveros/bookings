package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patrickoliveros/bookings/internal/helpers"
	"github.com/patrickoliveros/bookings/internal/renders"
)

var functions = template.FuncMap{
	"humanDate":    helpers.HumanDate,
	"formatDate":   helpers.FormatDate,
	"iterate":      helpers.Iterate,
	"add":          helpers.Add,
	"calendarDate": helpers.CalendarDate,
}

func TestRun(t *testing.T) {
	_, err := runApplication()

	if err != nil {
		t.Errorf("failed runApplication()")
	}

}

func TestMain(m *testing.M) {

	registerModels()
	setupApplicationConfig()
	setupDependencies()
	setupSession()

	err := setupApplicationTestTemplates()
	log.Println(err)

	os.Exit(m.Run())
}

func setupApplicationTestTemplates() error {
	tc, err := CreateAllTestTemplatesCache()
	helpers.HandleFatalError(err, "cannot create template cache")

	app.TemplateCache = tc

	renders.NewRenderer(&app)

	return err
}

func CreateAllTestTemplatesCache() (map[string]*template.Template, error) {
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

type myHandler struct{}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
