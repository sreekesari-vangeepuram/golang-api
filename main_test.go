package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	//"net/http/httptest"
	"testing"
)

const (
	PORT string = "3000"
	HOST string = "localhost"
	URI  string = "/api/users"
	URL  string = "http://" + HOST + ":" + PORT + URI
	ID   string = "322219098691863116" // <- Testing an entry
/*  ^^ PICK A NEW ID FROM DB AND CHANGE THIS CONSTANT DURING EVERT TEST */
)

// Testing GET request on an entry
func TestGetUser(t *testing.T) {
	req, err := http.NewRequest("GET", URL+"/"+ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("HTTP [GET] request failed: %v\n", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET request unsuccessful - %v\n", res.StatusCode)
	}
}

// Testing POST request
func TestCreateUser(t *testing.T) {
	var data = map[string]interface{}{
		"id":          nil,
		"name":        "Test User",
		"dob":         "1999-03-09T05:08:06.880755794+05:30",
		"address":     "Canada",
		"description": "I am a test user",
		"createdAt":   nil,
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")

	// Send request to API - for testing purposes
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		t.Fatalf("HTTP [POST] request failed: %v\n", err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("POST request unsuccessful - %v\n", res.StatusCode)
	}
}

// Testing PATCH request
func TestUpdateUser(t *testing.T) {

	var data = map[string]interface{}{
		"name":        "Anonymous User",
		"address":     "Anonymous Place",
		"description": "I am a Golang Developer",
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")

	// Send request to API - for testing purposes
	req, err := http.NewRequest("PATCH", URL+"/"+ID, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		t.Fatalf("HTTP [PATCH] request failed: %v\n", err)
	}

	if res.StatusCode != http.StatusPartialContent {
		t.Fatalf("PATCH request unsuccessful - %v\n", res.StatusCode)
	}
}

// Testing DELETE request on an entry
func TestDeleteUser(t *testing.T) {
	req, err := http.NewRequest("DELETE", URL+"/"+ID, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("HTTP [DELETE] request failed: %v\n", err)
	}

	if res.StatusCode != http.StatusAccepted {
		t.Fatalf("DELETE request unsuccessful - %v\n", res.StatusCode)
	}
}
