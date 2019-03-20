package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	firebase "firebase.google.com/go"
	"google.golang.org/appengine" // Required external App Engine library
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var (
	firebaseConfig = &firebase.Config{
		DatabaseURL:   "https://shopping-list-234812.firebaseio.com",
		ProjectID:     "shopping-list-234812",
		StorageBucket: "shopping-list-234812.appspot.com",
	}
	indexTemplate = template.Must(template.ParseFiles("index.html"))
)

// ShoppingItem : Item of the shopping list. Contains a name.
type ShoppingItem struct {
	UserID string
	Name   string
}

type templateParams struct {
	Notice   string
	ItemName string
	Items    []ShoppingItem
}

func main() {
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/", indexHandler)
	appengine.Main() // Starts the server to receive requests
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := appengine.NewContext(r)
	params := templateParams{}

	q := datastore.NewQuery("ShoppingItem").Limit(20)

	var keys []*datastore.Key
	var err error

	if keys, err = q.GetAll(ctx, &params.Items); err != nil {
		handleError(ctx, err, "datastore.GetAll: %v", "Couldn't get latest items. Refresh? %v!", params, w)
		return
	}

	if false {
		if err := datastore.DeleteMulti(ctx, keys); err != nil {
			handleError(ctx, err, "datastore.DeleteMulti: %v", "Items could not be deleted. Refresh? %v!", params, w)
			return
		}
	}

	if r.Method == "GET" {
		indexTemplate.Execute(w, params)
		return
	}

	itemName := r.FormValue("item")

	if (itemName) == "" {
		w.WriteHeader(http.StatusBadRequest)
		params.Notice = "Write a new item for the list before you submit"
		indexTemplate.Execute(w, params)
		return
	}

	// Create a new Firebase App.
	app, err := firebase.NewApp(ctx, firebaseConfig)
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "firebase.NewApp: %v", "Couldn't authenticate 1. Try logging in again? %v!", params, w)
		return
	}
	// Create a new authenticator for the app.
	authClient, err := app.Auth(ctx)
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "app.Auth: %v", "Couldn't authenticate 2. Try logging in again? %v!", params, w)
		return
	}
	// Verify the token passed in by the user is valid.
	tok, err := authClient.VerifyIDTokenAndCheckRevoked(ctx, r.FormValue("token"))
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "authClient.VerifyIDAndCheckRevoked: %v", "Couldn't authenticate 3. Try logging in again? %v!", params, w)
		return
	}
	// Use the validated token to get the user's information.
	user, err := authClient.GetUser(ctx, tok.UID)
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "auth.GetUser: %v", "Couldn't authenticate 4. Try logging in again? %v!", params, w)
		return
	}

	newItem := ShoppingItem{
		UserID: user.UID,
		Name:   itemName,
	}

	// Attach item to existing shopping list
	key := datastore.NewIncompleteKey(ctx, "ShoppingItem", nil)

	if _, err := datastore.Put(ctx, key, &newItem); err != nil {
		handleError(ctx, err, "datastore.Put: %v", "Couldn't add new item to the list. Try again? %v!", params, w)
		return
	}

	params.Items = append([]ShoppingItem{newItem}, params.Items...)

	indexTemplate.Execute(w, params)
}

func handleError(ctx context.Context, err error, logMessage string, noticeMessage string, params templateParams, w http.ResponseWriter) {
	log.Errorf(ctx, logMessage, err)
	w.WriteHeader(http.StatusInternalServerError)
	params.Notice = fmt.Sprintf(noticeMessage, err)
	indexTemplate.Execute(w, params)
}
