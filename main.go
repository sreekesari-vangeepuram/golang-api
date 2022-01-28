package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	DOB         string    `json:"dob"`
	Address     string    `json:"address"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
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
		return
	}

	// Adding/Updating the user's ctime
	newUser.CreatedAt = time.Now()

	// Change the below line -> Upsert to DB
	users = append(users, newUser)
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
