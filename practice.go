package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/santhosh-tekuri/jsonschema"
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

func validRequiredFieldsJSON(structure interface{}, reader io.Reader) (bool, error) {
	schema, err := jsonschema.Compile("userSchema.json")
	if err != nil {
		return false, err
	}
	if err = schema.Validate(reader); err != nil {
		return false, err
	}
	return true, err

}

func checkErrorServer(context *gin.Context, err error) {
	if err != nil {
		context.AbortWithStatusJSON(400, gin.H{
			"status":  "Error",
			"message": err,
		})
	}
}

func uploadUser(context *gin.Context) {
	fileHeader, err := context.FormFile("file")
	checkErrorServer(context, err)

	//Read the uploaded file
	file, err := fileHeader.Open()
	checkErrorServer(context, err)
	defer file.Close()

	//Validates that the info follows the schema
	_, err = validRequiredFieldsJSON(User{}, file)
	checkErrorServer(context, err)

	//Obtain the ID of the user
	user := make(map[string]interface{})
	fileBytes := make([]byte, 4096)
	file, _ = fileHeader.Open()
	cantBytes, err := file.Read(fileBytes)
	if err != nil && err != io.EOF {
		log.Printf("Error: %v\n", err)
		return
	}
	fileBytes = fileBytes[:cantBytes]
	json.Unmarshal(fileBytes, &user)
	id := user["ID"].(string)

	//Save the uploaded file
	err = context.SaveUploadedFile(fileHeader, "UserProfiles/"+string(id)+".json")
	checkErrorServer(context, err)

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
