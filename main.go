package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"firebase.google.com/go/auth"

	firebase "firebase.google.com/go"
	"google.golang.org/appengine" // Required external App Engine library
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var (
	firebaseConfig = &firebase.Config{
		DatabaseURL:      "https://shopping-list-234812.firebaseio.com",
		ProjectID:        "shopping-list-234812",
		StorageBucket:    "shopping-list-234812.appspot.com",
		ServiceAccountID: "firebase-adminsdk-w4g7i@shopping-list-234812.iam.gserviceaccount.com",
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
	itemName := r.FormValue("item")

	// If method is get, we just display the page.
	if r.Method == "GET" {
		indexTemplate.Execute(w, params)
		return
	}

	// Fetch all items
	q := datastore.NewQuery("ShoppingItem")

	var keys []*datastore.Key
	var err error
	if keys, err = q.GetAll(ctx, &params.Items); err != nil {
		handleError(ctx, err, "datastore.GetAll: %v", "Couldn't get latest items. Refresh? %v!", params, w)
		return
	}

	// If method post (it was not accepting delete) and hidden value delete, delete all items.
	if r.Method == "POST" && r.FormValue("delete") == "1" {
		if err := datastore.DeleteMulti(ctx, keys); err != nil {
			handleError(ctx, err, "datastore.DeleteMulti: %v", "Items could not be deleted. Refresh? %v!", params, w)
			return
		}
		params.Items = nil
		indexTemplate.Execute(w, params)
		return
	}

	// If method post but new item is empty, display a notice and render the page
	if (itemName) == "" {
		w.WriteHeader(http.StatusBadRequest)
		params.Notice = "Write a new item for the list before you submit"
		indexTemplate.Execute(w, params)
		return
	}

	user := retrieveUser(ctx, params, itemName, r.FormValue("token"), w)

	// Add the new item to the database.
	newItem := ShoppingItem{
		UserID: user.UID,
		Name:   itemName,
	}

	key := datastore.NewIncompleteKey(ctx, "ShoppingItem", nil)

	if _, err := datastore.Put(ctx, key, &newItem); err != nil {
		handleError(ctx, err, "datastore.Put: %v", "Couldn't add new item to the list. Try again? %v!", params, w)
		return
	}

	// Append new item to params
	params.Items = append([]ShoppingItem{newItem}, params.Items...)

	indexTemplate.Execute(w, params)
	return
}

func handleError(ctx context.Context, err error, logMessage string, noticeMessage string, params templateParams, w http.ResponseWriter) {
	log.Errorf(ctx, logMessage, err)
	w.WriteHeader(http.StatusInternalServerError)
	params.Notice += fmt.Sprintf(noticeMessage, err)
	indexTemplate.Execute(w, params)
}

func retrieveUser(ctx context.Context, params templateParams, itemName string, token string, w http.ResponseWriter) *auth.UserRecord {
	// Create a new Firebase App.
	app, err := firebase.NewApp(ctx, firebaseConfig)
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "firebase.NewApp: %v", "Couldn't authenticate 1. Try logging in again? %v!", params, w)
	}
	// Create a new authenticator for the app.
	authClient, err := app.Auth(ctx)
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "app.Auth: %v", "Couldn't authenticate 2. Try logging in again? %v!", params, w)
	}
	// Verify the token passed in by the user is valid.
	tok, err := authClient.VerifyIDTokenAndCheckRevoked(ctx, token)
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "authClient.VerifyIDAndCheckRevoked: %v", "Couldn't authenticate 3. Try logging in again? %v!", params, w)
	}
	// Use the validated token to get the user's information.
	user, err := authClient.GetUser(ctx, tok.UID)
	if err != nil {
		params.ItemName = itemName // Preserve their input so they can try again.
		handleError(ctx, err, "auth.GetUser: %v", "Couldn't authenticate 4. Try logging in again? %v!", params, w)
	}

	return user
}
