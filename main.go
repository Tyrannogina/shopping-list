package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

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

type templateParams struct {
	Notice        string
	ItemName      string
	ShoppingLists []ShoppingList
}

// User : owner of the shopping list, has a name and a UserID
type User struct {
	Name   string
	UserID string
}

// Item : Item of the shopping list. Contains a name and can be striked through.
type Item struct {
	Name    string
	Striked bool
}

// ShoppingList : allows to save several items. Has an author with their corresponding id, creation time, name and list of items.
type ShoppingList struct {
	//Owner     User,
	Name      string
	Items     []Item
	CreatedAt time.Time
}

func main() {
	http.HandleFunc("/", indexHandler)
	appengine.Main() // Starts the server to receive requests
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := appengine.NewContext(r)
	params := templateParams{}

	q := datastore.NewQuery("ShoppingList").Order("-CreatedAt").Limit(20)

	var keys []*datastore.Key
	var err error

	if keys, err = q.GetAll(ctx, &params.ShoppingLists); err != nil {
		log.Errorf(ctx, "Getting shopping lists: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		params.Notice = "Couldn't get latest shopping Lists. Refresh?"
		indexTemplate.Execute(w, params)
		return
	}

	for i := 0; i < len(keys); i++ {
		params.Notice += fmt.Sprintf("Keys: %v", keys)
	}

	if false {
		if err := datastore.DeleteMulti(ctx, keys); err != nil {
			log.Errorf(ctx, "Shopping lists could not be deleted: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			params.Notice = "Shopping lists could not be deleted. Refresh?"
			indexTemplate.Execute(w, params)
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
		params.Notice = "No message provided"
		indexTemplate.Execute(w, params)
		return
	}
	/*
		// Create a new Firebase App.
		app, err := firebase.NewApp(ctx, firebaseConfig)
		if err != nil {
			params.Notice = "Couldn't authenticate 1. Try logging in again?"
			log.Errorf(ctx, "firebase.NewApp: %v", err)
			params.ItemName = itemName // Preserve their input so they can try again.
			indexTemplate.Execute(w, params)
			return
		}
		// Create a new authenticator for the app.
		authClient, err := app.Auth(ctx)
		if err != nil {
			params.Notice = "Couldn't authenticate 2. Try logging in again?"
			log.Errorf(ctx, "app.Auth: %v", err)
			params.ItemName = itemName // Preserve their input so they can try again.
			indexTemplate.Execute(w, params)
			return
		}

		// Verify the token passed in by the user is valid.
		tok, err := authClient.VerifyIDTokenAndCheckRevoked(ctx, r.FormValue("token"))
		if err != nil {
			params.Notice = fmt.Sprintf("Couldn't authenticate 3. Try logging in again? auth.VerifyIDAndCheckRevoked: %v", err)
			log.Errorf(ctx, "auth.VerifyIDAndCheckRevoked: %v", err)
			params.ItemName = itemName // Preserve their input so they can try again.
			indexTemplate.Execute(w, params)
			return
		}
		// Use the validated token to get the user's information.
		user, err := authClient.GetUser(ctx, tok.UID)
		if err != nil {
			params.ItemName = itemName // Preserve their input so they can try again.
			handleError(ctx, err, "auth.GetUser: %v", ""Couldn't authenticate 4. Try logging in again? %v!", params, w)
			return
		}
	*/
	newItem := Item{
		Name:    itemName,
		Striked: false,
	}
	// Insert new shopping list if one does not exist already
	if len(params.ShoppingLists) == 0 {
		name := "My groceries"

		shoppingList := ShoppingList{
			Name:      name,
			Items:     []Item{newItem},
			CreatedAt: time.Now(),
		}

		key := datastore.NewIncompleteKey(ctx, "ShoppingList", nil)

		if _, err := datastore.Put(ctx, key, &shoppingList); err != nil {
			handleError(ctx, err, "datastore.Put: %v", "Couldn't add new shopping list list. Try again? %v!", params, w)
			return
		}

		params.ShoppingLists = append([]ShoppingList{shoppingList}, params.ShoppingLists...)
	} else {
		// Attach item to existing shopping list
		key := datastore.NewIncompleteKey(ctx, "Item", keys[0])
		params.Notice += fmt.Sprintf("OMG whats the new item key %v!", key)
		var currentShoppingList ShoppingList

		if err := datastore.Get(ctx, keys[0], &currentShoppingList); err != nil {
			handleError(ctx, err, "datastore.Put: %v", "Couldn't add new item to the list. Try again? %v!", params, w)
			return
		}

		if _, err := datastore.Put(ctx, key, &newItem); err != nil {
			handleError(ctx, err, "datastore.Put: %v", "Couldn't add new item to the list. Try again? %v!", params, w)
			return
		}

		params.Notice += fmt.Sprintf("OMG whats the new item %v!", newItem)

	}

	params.Notice += fmt.Sprintf("Hey You! %v\n", params.ShoppingLists) //user.DisplayName)
	indexTemplate.Execute(w, params)
}

func handleError(ctx context.Context, err error, logMessage string, noticeMessage string, params templateParams, w http.ResponseWriter) {
	log.Errorf(ctx, logMessage, err)
	w.WriteHeader(http.StatusInternalServerError)
	params.Notice = fmt.Sprintf(noticeMessage, err)
	indexTemplate.Execute(w, params)
}
