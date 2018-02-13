package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
)

// List : Struc de lista
type List struct {
	Users []string `binding:"required"`
}

func recoverServerError(context *gin.Context) {
	if r := recover(); r != nil {
		context.JSON(400, gin.H{
			"error": r,
		})
	}
}

func checkError(err error, message string) {
	if err != nil {
		log.Println(err)
		panic(message)
	}
}

func pingPong(context *gin.Context) {
	context.JSON(200, gin.H{
		"message": "pong",
	})
}

func fillUserProfile(user User) error {
	tmpl, err := template.ParseFiles("UserProfiles/profileTemplate.json")
	if err != nil {
		fmt.Println(err)
		return errors.New("The user profile was not made")
	}
	file, err := os.Create("UserProfiles/" + user.ID + ".json")
	defer file.Close()
	err = tmpl.Execute(file, user)

	if err != nil {
		fmt.Println(err)
		return errors.New("The given inputs not match the user profile")
	}
	return nil
}

func main() {
	router := gin.Default()

	router.GET("/ping", pingPong)
	router.GET("/users", getUsers)
	router.GET("/user/:id", getUser)
	router.GET("/user/:id/:data", getUserData)
	router.GET("/xml/user/:id", getUserXML)
	router.GET("/xml/users", getUsersXML)
	router.POST("/user", postUser)
	router.POST("/user/:id/:data", postUserData)
	router.DELETE("/user/:id", deleteUser)
	router.Run(":8080") // listen and serve on 127.0.0.1:8080
}
