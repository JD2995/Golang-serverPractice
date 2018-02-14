package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
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

func validRequiredFieldsJSON(structure interface{}, jsonBytes []byte) (map[string]*json.RawMessage, error) {
	userMap := make(map[string]*json.RawMessage)
	err := json.Unmarshal(jsonBytes, &userMap)
	if err != nil {
		return nil, err
	}

	t := reflect.TypeOf(structure)
	for i := 0; i < t.NumField(); i++ {
		tagMap := make(map[string]string)
		nameField := t.Field(i).Name
		rawTag := strings.Replace(string(t.Field(i).Tag), "\"", "", -1)
		tags := strings.Split(rawTag, " ")
		for _, tag := range tags {
			tagElements := strings.Split(tag, ":")
			tagMap[tagElements[0]] = tagElements[1]

		}

		if tagMap["json"] != "" {
			nameField = tagMap["json"]
		}
		//Doesn't matter if the value is not found
		if tagMap["binding"] != "required" {
			continue
		}

		//If the Required field is nil
		if userMap[nameField] == nil {
			return nil, fmt.Errorf("Required %s field not given", nameField)
		}
	}
	return userMap, nil
}

func uploadUser(context *gin.Context) {
	fileHeader, err := context.FormFile("file")
	if err != nil {
		log.Printf("The file wasn't uploaded: %v\n", err)
	}

	//Read the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("Cannot read the file: %v\n", err)
	}
	buffer := make([]byte, 2048)
	nBytes, err := file.Read(buffer)
	if err != nil {
		log.Printf("Cannot read the file: %v\n", err)
	}
	buffer = buffer[:nBytes]

	user, err := validRequiredFieldsJSON(User{}, buffer)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	id := make([]byte, 2048)
	user["ID"].UnmarshalJSON(id)
	fmt.Printf("%s\n", user)
	err = context.SaveUploadedFile(fileHeader, "UserProfiles/"+string(id)+".json")
	if err != nil {
		log.Printf("The file wasn't saved: %v\n", err)
	}

	context.JSON(201, gin.H{
		"URI": "/user/" + string(id),
	})
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
	router.POST("/upload/user", uploadUser)
	router.Run(":8080") // listen and serve on 127.0.0.1:8080
}
