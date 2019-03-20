package main

import (
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

	if _, err := q.GetAll(ctx, &params.ShoppingLists); err != nil {
		log.Errorf(ctx, "Getting shopping lists: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		params.Notice = "Couldn't get latest shopping Lists. Refresh?"
		indexTemplate.Execute(w, params)
		return
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
		params.Notice = "Couldn't authenticate 4. Try logging in again?"
		log.Errorf(ctx, "auth.GetUser: %v", err)
		params.ItemName = itemName // Preserve their input so they can try again.
		indexTemplate.Execute(w, params)
		return
	}

	// Insert new shopping list if one does not exist already
	if len(params.ShoppingLists) == 0 {
		name := "My groceries"

		shoppingList := ShoppingList{
			Name:      name,
			CreatedAt: time.Now(),
		}

		key := datastore.NewIncompleteKey(ctx, "ShoppingList", nil)

		if _, err := datastore.Put(ctx, key, &shoppingList); err != nil {
			log.Errorf(ctx, "datastore.Put: %v", err)
			params.Notice = fmt.Sprintf("Thank you for your submission, %s!", user.DisplayName)
			w.WriteHeader(http.StatusInternalServerError)
			params.Notice = "Couldn't add new shopping list. Try again?"
			indexTemplate.Execute(w, params)
			return
		}

		params.ShoppingLists = append([]ShoppingList{shoppingList}, params.ShoppingLists...)
	}

	params.Notice = fmt.Sprintf("Hello %s", user.DisplayName)
	indexTemplate.Execute(w, params)
}
