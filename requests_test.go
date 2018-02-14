package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func makePostUser(user User) (*http.Response, error) {
	userString := make([]byte, 2038)
	userString, err := json.Marshal(&user)
	if err != nil {
		return nil, fmt.Errorf("Testing internal error: %v", err)
	}
	resp, err := http.Post("http://127.0.0.1:8080/user", "Content-Type: application/json",
		bytes.NewReader(userString))
	if err != nil {
		return nil, fmt.Errorf("There was error at making the request: %v", err)
	}
	return resp, nil
}

func makeDeleteUser(id string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "http://127.0.0.1:8080/user/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("Testing error at making request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("There was an error at making the request: %v", err)
	}
	return resp, nil
}

func TestGetUsers(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:8080/users")
	if err != nil {
		t.Errorf("There was error at making the request: %v\n", err)
	}
	body := resp.Body
	bodyBuffer := make([]byte, 2048)
	nBytes, err := body.Read(bodyBuffer)
	if err != nil && err != io.EOF {
		t.Errorf("Cannot read body: %v\n", err)
	}
	bodyBuffer = bodyBuffer[:nBytes]

	var users = new(struct {
		Users []string
	})
	err = json.Unmarshal(bodyBuffer, &users)
	switch err := err.(type) {
	case nil:
		//Continue
	case *json.UnmarshalTypeError:
		t.Errorf("The received list %v doesn't match the wanted one\n", bodyBuffer)
		fmt.Printf("Cannot decode the json data, %v\n", err)

	default:
		t.Errorf("Unexpected error: %v\n", err)
	}
}

func TestPostUser(t *testing.T) {
	user := User{Name: "Javier", ID: "702390421", Lastname: "Rivas",
		Address: Address{Provincia: "Limón", Canton: "Limón", Distrito: "Limón"},
		Phones:  []int{84139034, 27585124}}

	resp, err := makePostUser(user)
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("Cannot add the user, wanted status code %d, received %d\n",
			201, resp.StatusCode)
	}
	if _, err = makeDeleteUser("702390421"); err != nil {
		fmt.Printf("The created user wasn't deleted in TestPostUser")
	}
}

func TestIncompletePostUser(t *testing.T) {
	user := User{Name: "Javier", ID: "702390421",
		Address: Address{Provincia: "Limón", Canton: "Limón", Distrito: "Limón"},
		Phones:  []int{84139034, 27585124}}
	resp, err := makePostUser(user)
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("The server aren't handling the error correctly, wanted status code %d"+
			", received %d\n", 400, resp.StatusCode)
	}
}

func TestDeleteExistingUser(t *testing.T) {
	user := User{Name: "Javier", ID: "702390421", Lastname: "Rivas",
		Address: Address{Provincia: "Limón", Canton: "Limón", Distrito: "Limón"},
		Phones:  []int{84139034, 27585124}}
	resp, err := makePostUser(user)
	if err != nil {
		t.Errorf("Internal testing fail: %v\n", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("Internal testing fail: Cannot create a new user\n")
	}

	resp, err = makeDeleteUser("702390421")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if resp.StatusCode != 204 {
		t.Errorf("The existing user wasn't deleted\n")
	}
}

func TestDeleteNoExistingUser(t *testing.T) {
	resp, err := makeDeleteUser("702390421")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("The server isn't handling correctly the no existence of an user,"+
			"expected status %d, received %d\n", 404, resp.StatusCode)
	}
}
