package api

import (
	"encoding/json"
	"net/http"

	"github.com/patrickoliveros/bookings/models"
)

var Articles []models.Article

func GetArticles(w http.ResponseWriter, r *http.Request) {
	Articles = []models.Article{
		{Title: "Hello", Desc: "Article Description 1", Content: "Article Content 1"},
		{Title: "Hello 2", Desc: "Article Description 2", Content: "Article Content 2"},
	}

	json.NewEncoder(w).Encode(Articles)
}
