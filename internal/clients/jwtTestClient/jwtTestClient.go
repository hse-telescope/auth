package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "http://localhost:8080"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
	Token   string `json:"token"`
}

func main() {
	for {
		var username, password string
		fmt.Scan(&username, &password)
		registerUser(username, password)
		fmt.Scan(&username, &password)
		token := loginUser(username, password)
		getUsers(token)
		currTime := time.Now()
		waittime := time.Now().Add(1 * time.Minute).Add(100 * time.Millisecond)
		for currTime.Compare(waittime) != 1 {
			currTime = time.Now()
		}
		getUsers(token)
	}
}

func registerUser(username, password string) {
	url := baseURL + "/register"
	creds := Credentials{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(creds)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return
	}

	fmt.Println("Register Response:", response)
}

func loginUser(username, password string) string {
	url := baseURL + "/login"
	creds := Credentials{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(creds)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return ""
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return ""
	}

	fmt.Println("Login Response:", response)
	return response.Token
}

func getUsers(token string) {
	url := baseURL + "/users"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Println("Get Users Response:", string(body))
}
