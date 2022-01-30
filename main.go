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
	DOB         time.Time `json:"dob" fauna"dob"`
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

// getUser responds to GET requests with `id` parameter to API
func getUser(ctx *gin.Context) {
	var id string = ctx.Param("id")

	result, err := adminClient.Query(
		f.Get(f.Ref(f.Collection("Users"), id)))

	// In case user not found
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// Shape the retrived data into <User> type
	var user User
	if err = result.At(f.ObjKey("data")).Get(&user); err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to fetch user details"})
		return
	}

	// Respond to client on read
	ctx.IndentedJSON(http.StatusOK, user)
}

// createUser responds to POST requests to API
func createUser(ctx *gin.Context) {

	var newUser User
	if err := ctx.BindJSON(&newUser); err != nil {
		ctx.IndentedJSON(
			http.StatusNotAcceptable,
			gin.H{"message": "invalid JSON data sent"},
		)
		return
	}

	// Get new id for document
	id, err := newID()
	if err != nil {
		ctx.IndentedJSON(
			http.StatusInternalServerError,
			gin.H{"message": "unable to generate an id for the user"},
		)
		return
	}

	// Adding the user's id & ctime
	newUser.ID = id
	newUser.CreatedAt = time.Now()

	// Commiting user details to DB
	_, err = adminClient.Query(
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

	if err != nil {
		ctx.IndentedJSON(
			http.StatusInternalServerError,
			gin.H{"message": "unable to create document"},
		)
		return
	}

	// Respond to client on create
	ctx.IndentedJSON(http.StatusCreated, newUser)
}

// updateUser responds to PATCH requests to API
func updateUser(ctx *gin.Context) {

	var id string = ctx.Param("id")
	var updatedUser User

	if err := ctx.BindJSON(&updatedUser); err != nil {
		return
	}

	/// Similar to getUser func ///
	result, err := adminClient.Query(
		f.Get(f.Ref(f.Collection("Users"), id)))

	// Incase user not found
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	var user User
	if err = result.At(f.ObjKey("data")).Get(&user); err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to fetch user details"})
		return
	}
	//////////////////////////////

	checkString := func(f1, f2 string) string {
		if f1 != "" {
			return f1
		}
		return f2
	}

	// ID, DOB & CreatedAt fields can't be changed in future after registration
	result, err = adminClient.Query(
		f.Update(
			f.Ref(f.Collection("Users"), id),
			f.Obj{
				"data": f.Obj{
					"name":        checkString(updatedUser.Name, user.Name),
					"address":     checkString(updatedUser.Address, user.Address),
					"description": checkString(updatedUser.Description, user.Description),
				},
			},
		))

	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// Fetch the updated version of user details
	if err = result.At(f.ObjKey("data")).Get(&updatedUser); err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "unable to fetch user details"})
		return
	}

	// Respond the updated details to client
	ctx.IndentedJSON(http.StatusPartialContent, updatedUser)
}

// deleteUser responds to DELETE requests to API
func deleteUser(ctx *gin.Context) {
	var id string = ctx.Param("id")

	_, err := adminClient.Query(
		f.Delete(f.Ref(f.Collection("Users"), id)))

	// In case user not found
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// Respond to client on delete
	ctx.IndentedJSON(http.StatusAccepted, gin.H{"message": id + ": user successfully deleted"})
}

func main() {

	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()

	router.GET("/api/users/:id", getUser)
	router.POST("/api/users", createUser)
	router.PATCH("/api/users/:id", updateUser)
	router.DELETE("/api/users/:id", deleteUser)

	router.Run(":3000")

}
