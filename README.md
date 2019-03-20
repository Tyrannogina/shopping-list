# Go Shopping

## Access your shopping list
This code is deployed and accessible here: https://shopping-list-234812.appspot.com/.

You can run the application locally by following this steps:
1. Prepare the environment: you will need python and golang installed in your system.
2. Clone this repository into your Go project folder.
3. Execute the following command `dev_appserver.py app.yaml`, which will allow you to access the project in `localhost:8080`.

## First time learning and interacting with
- Golang
- Google Cloud
- gcloud CLI
- Firebase
- Datastore

## TODO:
- Delete shopping lists.
- Add items.
- Delete items.
- Delete all items in a list.
- Create a minimal UI.

## Known issues/bugs
- User token validation fails when request is sent from localhost, which is quite particular since the user is able to log in and out normally.

## Developing process
The time I had available for this project has been very limited by other obligations, so I would appreciate if this is taken into account.
### Bad things
I grossly underestimated how different is Golang compared with the languages I know and how complicated I would find it. In retrospective, I would have chosen python or Java to work on this project, to minimise the time I spent stuck.
### Good things
I managed to have a basic application working and deployed in a couple evenings even though it was my first time using all the tools. Kudos to people writing the documentation for AppEngine, it's beautiful and accessible.

## Tools
- VSCode
- https://marketplace.visualstudio.com/items?itemName=ms-vscode.Go

## Resources
- https://cloud.google.com/appengine/docs/standard/go/building-app/creating-your-application
- https://tour.golang.org/welcome/1
- https://cloud.google.com/sdk/gcloud/reference/app/deploy
- https://godoc.org/google.golang.org/appengine
