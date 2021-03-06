package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

// Address : Struc de la direccion
type Address struct {
	Provincia string `json:"provincia" binding:"required"`
	Canton    string `json:"canton" binding:"required"`
	Distrito  string `json:"distrito" binding:"required"`
}

// User : Struc del usuario
type User struct {
	ID             string  `json:"ID" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Lastname       string  `json:"lastname" binding:"required"`
	Address        Address `json:"address" binding:"required"`
	Phones         []int   `json:"phones" binding:"required"`
	PoliticalParty string  `json:"politicalParty"`
}

// PoliticalParty : Describes the name and quantity of members
type PoliticalParty struct {
	Name            string
	QuantityMembers int
}

// readUserFile reads the user file of a specified user by an id
func readUserFile(userID string) ([]byte, error) {
	//Tried to open the file
	file, err := os.Open("UserProfiles/" + userID + ".json")
	if err != nil {
		log.Println(err)
		return nil, errors.New("Error, the searched user was not found")
	}
	defer file.Close()

	//Tried to read the file
	dataFile := make([]byte, 4096)
	nBytes, err := file.Read(dataFile)
	if err != nil {
		log.Println(err)
		return nil, errors.New("The searched user was not found")
	}

	return dataFile[:nBytes], nil
}

// structUserJSON obtain the user's info on the server and convert it on a user struct
func structUserJSON(userID string) (User, error) {
	var user User

	dataFile, err := readUserFile(userID)
	checkError(err, "Cannot read user file")

	err = json.Unmarshal(dataFile, &user)
	if err != nil {
		log.Println(err)
		return user, errors.New("Data not found")
	}

	return user, nil
}

// mapUserJSON obtain the user's info on the server and convert it on a map
func mapUserJSON(userID string) (map[string]*json.RawMessage, error) {
	userMap := make(map[string]*json.RawMessage)

	dataFile, err := readUserFile(userID)
	checkError(err, "Cannot read user file")

	err = json.Unmarshal(dataFile, &userMap)
	if err != nil {
		log.Println(err)
		return nil, errors.New("Data not found")
	}

	return userMap, nil
}

// getUser handles the get request of all the user's info
func getUser(context *gin.Context) {
	defer recoverServerError(context)

	id := context.Param("id")
	file, err := os.Open("UserProfiles/" + id + ".json")
	checkError(err, "Error, the searched user was not found")
	defer file.Close()

	userData := make([]byte, 1024)
	_, err = file.Read(userData)
	checkError(err, "The data from the file wasn't obtained")
	context.Data(200, "Content-Type: application/json", userData)
}

// getUserData handles the get request of specific user's data
func getUserData(context *gin.Context) {
	defer recoverServerError(context)

	id := context.Param("id")
	userMap, err := mapUserJSON(id)
	checkError(err, "Fail to get of data")

	//Try to obtain the wanted data
	data := context.Param("data")
	value := userMap[data]
	if value == nil {
		log.Println("Data not found")
		context.AbortWithStatusJSON(404, gin.H{
			"error": "Data not found",
		})
		return
	}
	//Return the obtained data
	context.JSON(200, gin.H{
		data: value,
	})
}

// deleteUser handles the delete request of an user
func deleteUser(context *gin.Context) {
	defer recoverServerError(context)

	id := context.Param("id")
	err := os.Remove("UserProfiles/" + id + ".json")
	if err != nil {
		log.Println(err)
		context.AbortWithStatus(404)
	}
	context.Data(204, gin.MIMEHTML, nil)
}

// searchUsersFiles searchs the user files on the UserProfiles directory
func searchUsersFiles() []string {
	file, err := os.Open("UserProfiles")
	checkError(err, "The user profile wasn't found")
	defer file.Close()

	filesInfo, err := file.Readdir(0)
	checkError(err, "Server error")

	//Obtain the names of the files and deletes the .json extension
	var fileNames []string
	for _, value := range filesInfo {
		fileName := value.Name()
		if fileName == "profileTemplate.json" || fileName == "profileTemplate.xml" {
			continue
		}
		fileName = strings.TrimSuffix(fileName, ".json")
		fileNames = append(fileNames, fileName)
	}
	return fileNames

}

// getUsers handles a get request to obtain the ids of all the users on the server
func getUsers(context *gin.Context) {
	defer recoverServerError(context)

	filesNames := searchUsersFiles()
	names := new(List)
	for _, value := range filesNames {
		names.Users = append(names.Users, value)
	}
	context.JSON(200, names)
}

// postUser handles a post request that it has in it's body the user info
// and saves it on the server
func postUser(context *gin.Context) {
	defer recoverServerError(context)

	var user User
	err := context.BindJSON(&user)
	checkError(err, "The send data wasn't correctly formatted")

	err = fillUserProfile(user)
	checkError(err, "The user wasn't correctly created")

	context.JSON(201, gin.H{
		"message": "Se ha creado el perfil de usuario",
	})
}

// postUserData handles a post request that updates a specific value of the user
func postUserData(context *gin.Context) {
	defer recoverServerError(context)

	//Obtain the user's saved data
	id := context.Param("id")
	userMap, err := mapUserJSON(id)
	checkError(err, "Wasn't able to get the saved data")

	//Obtain the data received from POST
	dataName := context.Param("data")
	dataReceive := make(map[string]*json.RawMessage)
	err = context.BindJSON(&dataReceive)
	checkError(err, "Wasn't able to get the requested change")

	//Makes the changes in the user map and convert it to json
	dataValue := dataReceive[dataName]
	userMap[dataName] = dataValue
	jsonString, err := json.Marshal(userMap)
	checkError(err, "There was a problem at saving the file")

	//Delete the old user file
	err = os.Remove("UserProfiles/" + id + ".json")
	checkError(err, "The file wasn't saved")

	//Save the user file with the new changes
	file, err := os.OpenFile("UserProfiles/"+id+".json", os.O_CREATE, 0755)
	checkError(err, "There was a problem at saving the file")
	defer file.Close()
	_, err = file.WriteString(string(jsonString))
	checkError(err, "There was a problem at saving the file")

	context.JSON(200, gin.H{
		"mensaje": "Se ha actualizado exitosamente el dato",
	})

}

// getUserXML return the user info and convert it to an xml file
func getUserXML(context *gin.Context) {
	defer recoverServerError(context)
	id := context.Param("id")

	type ListUsers struct {
		Users []User
	}

	//Obtain the user's data from the json file
	user, err := structUserJSON(id)
	checkError(err, "Fail to get of user data")

	//Load and define the user xml template
	list := new(ListUsers)
	list.Users = append(list.Users, user)
	tmpl, err := template.New("profileTemplate.xml").Funcs(template.FuncMap{
		"getPoliticalParties": getPoliticalParties,
		"getElectionsResult":  getElectionsResult,
	}).ParseFiles("UserProfiles/profileTemplate.xml")
	//Execute the user's data against the template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, list)
	checkError(err, "Cannot execute template")
	context.Data(200, "Content-Type: application/xml", buffer.Bytes())

}

// getPoliticalParties obtains all the political parties from a list of users
func getPoliticalParties(users []User) []PoliticalParty {
	var parties []PoliticalParty
	mapParties := make(map[string]int)
	for _, user := range users {
		if user.PoliticalParty != "" {
			mapParties[user.PoliticalParty]++
		}
	}
	//Make the map an array
	for key, value := range mapParties {
		party := PoliticalParty{key, value}
		parties = append(parties, party)
	}
	return parties
}

// getElectionsResult calc the winner of the elections
func getElectionsResult(parties []PoliticalParty) PoliticalParty {
	maximum := 0
	for index, party := range parties {
		if parties[maximum].QuantityMembers < party.QuantityMembers {
			maximum = index
		}
	}
	return parties[maximum]
}

// getUsersXML obtain all the users found in the UserProfiles directory
// Converts the info from json to xml and also perfoms an operation that
// shows the politic parties that the users are members
func getUsersXML(context *gin.Context) {
	defer recoverServerError(context)

	type ListUsers struct {
		Users []User
	}

	//Gets all the user's data from the directory
	filesNames := searchUsersFiles()
	list := new(ListUsers)
	for _, value := range filesNames {
		//Recuperar archivo json
		user, err := structUserJSON(value)
		checkError(err, "Fail to get of user data")

		list.Users = append(list.Users, user)
	}

	//Defines the template and its functions
	tmpl, err := template.New("profileTemplate.xml").Funcs(template.FuncMap{
		"getPoliticalParties": getPoliticalParties,
		"getElectionsResult":  getElectionsResult,
	}).ParseFiles("UserProfiles/profileTemplate.xml")
	checkError(err, "The template file wasn't loaded")

	//Executes the user's data against the template
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, list)
	checkError(err, "Cannot execute template")
	context.Data(200, "Content-Type: application/xml", buffer.Bytes())
}
