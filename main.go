package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	f "github.com/fauna/faunadb-go/v4/faunadb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type User struct {
	ID          string    `json:"id,omitempty" fauna:"id"`
	Name        string    `json:"name" fauna:"name"`
	DOB         string    `json:"dob" fauna"dob"`
	Address     string    `json:"address" fauna:"address"`
	Description string    `json:"description" fauna:"description"`
	CreatedAt   time.Time `json:"createdAt,omitempty" fauna:"createdAt"`
}

// Global variable to access admin-client to faunaDB
var adminClient *f.FaunaClient

func init() {

	var err error = godotenv.Load()
	if err != nil {
		panic("[ERROR]: Unable to load <.env> file!")
	}

	// Get a new fauna client with access key
	adminClient = f.NewFaunaClient(os.Getenv("FAUNADB_ADMIN_SECRET"))

	// Create a persistent database node for the API
	result, err := adminClient.Query(
		f.CreateDatabase(f.Obj{"name": "golang-api"}))

	handleError(result, err)

	// Create a collection (table) inside the database
	result, err = adminClient.Query(
		f.CreateCollection(f.Obj{"name": "Users"}))

	handleError(result, err)

	// Create an index to access documents easily
	result, err = adminClient.Query(
		f.CreateIndex(
			f.Obj{
				"name":   "users_by_id",
				"source": f.Collection("Users"),
				"terms":  f.Arr{f.Obj{"field": f.Arr{"data", "id"}}},
			},
		))

	handleError(result, err)

}

// Handles the error and prompts the result accordingly
func handleError(result f.Value, err error) {
	if err != nil {
		fmt.Printf("[FAUNADB-WARN]: ")
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("[FAUNADB-DEBUG]: ")
		fmt.Println(result)
	}
}

// Fetch new ID for a document on call
func newID() (id string, err error) {
	result, err := adminClient.Query(f.NewId())
	if err != nil {
		return "", err
	}

	err = result.Get(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

// slice of <User>s - for temporary use only!
var users []User = []User{
	{
		ID:          "1",
		Name:        "Sreekesari Vangeepuram",
		DOB:         "27-01-2003",
		Address:     "6-3-787/R/306, Hyderabad, Telangana, India - 500016",
		Description: "Minimalist | Programmer | Explorer",
		CreatedAt:   time.Now(),
	},
	{
		ID:          "2",
		Name:        "Vemula Vamshi Krishna",
		DOB:         "20-06-2004",
		Address:     "Near AMB Mall, Rajendra Nagar, Kondapur, Hyderabad, Telangana, India",
		Description: "Intermediate Student",
		CreatedAt:   time.Now(),
	},
	{
		ID:          "3",
		Name:        "Chepuri Vikram Bharadwaj",
		DOB:         "16-12-2000",
		Address:     "Near Alpha Hotel, Secendrabad East, Hyderabad, Telangana, India",
		Description: "GOAT",
		CreatedAt:   time.Now(),
	},
}

// getUsers responds to GET requests to API
func getUsers(ctx *gin.Context) {
	// Change <users> part
	ctx.IndentedJSON(http.StatusOK, users)
}

// getUserByID responds to GET requests with `id` parameter to API
func getUserByID(ctx *gin.Context) {
	var id string = ctx.Param("id")

	// Change this part with call to DB
	for _, user := range users {
		if user.ID == id {
			ctx.IndentedJSON(http.StatusOK, user)
			return
		}
	}

	ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

// createUsers responds to POST requests to API
func createUsers(ctx *gin.Context) {

	var newUser User
	if err := ctx.BindJSON(&newUser); err != nil {
		ctx.IndentedJSON(
			http.StatusNotAcceptable,
			gin.H{"message": "Invalid JSON!"},
		)
		return
	}

	// Get new id for document
	id, err := newID()
	if err != nil {
		ctx.IndentedJSON(
			http.StatusInternalServerError,
			gin.H{"message": "Unable to generate an id for the user!"},
		)
		return
	}

	// Adding the user's id & ctime
	newUser.ID = id
	newUser.CreatedAt = time.Now()

	// Change the below line -> Upsert to DB
	users = append(users, newUser) // DEBUG
	result, err := adminClient.Query(
		f.Create(
			f.Ref(f.Collection("Users"), id),
			f.Obj{
				"data": f.Obj{
					"id":          newUser.ID,
					"name":        newUser.Name,
					"dob":         newUser.DOB,
					"address":     newUser.Address,
					"description": newUser.Description,
					"createdAt":   newUser.CreatedAt,
				},
			},
		))

	handleError(result, err)

	// Respond client
	ctx.IndentedJSON(http.StatusCreated, newUser)
}

// updateUsers responds to PATCH requests to API
func updateUsers(ctx *gin.Context) {

	var id string = ctx.Param("id")
	var updatedUser User

	if err := ctx.BindJSON(&updatedUser); err != nil {
		return
	}

	// [ID] and [CreatedAt] fields can't be either created or changed
	for _, user := range users {
		if user.ID == id {
			// Update `name`
			if updatedUser.Name != "" && updatedUser.Name != user.Name {
				user.Name = updatedUser.Name
			}

			// Update `dob`
			if updatedUser.DOB != "" && updatedUser.DOB != user.DOB {
				user.DOB = updatedUser.DOB
			}

			// Update `address`
			if updatedUser.Address != "" && updatedUser.Address != user.Address {
				user.Address = updatedUser.Address
			}

			// Update `description`
			if updatedUser.Description != "" && updatedUser.Description != user.Description {
				user.Description = updatedUser.Description
			}

			ctx.IndentedJSON(http.StatusPartialContent, user)
			return
		}
	}

	ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

// deleteUsers responds to DELETE requests to API
func deleteUsers(ctx *gin.Context) {
	var id string = ctx.Param("id")

	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			ctx.IndentedJSON(http.StatusAccepted, user)
			return
		}
	}

	ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

func main() {

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()

	router.GET("/api/users", getUsers)
	router.GET("/api/users/:id", getUserByID)
	router.POST("/api/users", createUsers)
	router.PATCH("/api/users/:id", updateUsers)
	router.DELETE("/api/users/:id", deleteUsers)

	router.Run(":3000")

}
