package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	baseURL   = "https://api.github.com"
	org       = "vishwa-test-2"
	repo      = "sample"
	authToken = "eyJhbGciOiJFUzI1NiIsImVudiI6InByb2QiLCJraWQiOiJnX0hsUlpnWGRzMlNFcUVFUGVzZmdiUDBaRWxIV2tiaFJ5SUd4alRKMWNrIiwidHlwIjoiSldUIiwidmVyIjoiMS4wIn0.eyJjbGllbnRfaWQiOiJnaXQudWJlcmludGVybmFsLmNvbSIsImVtYWlsIjoidnBhdGlsMTZAZXh0LnViZXIuY29tIiwiZXhwIjoxNzQxMjc1MTAyLCJpYXQiOjE3NDEyMDI4MDIsImlzcyI6InNwaWZmZTovL3Vzc28udXBraS5jYSIsImp0aSI6IjQwY2Q2MDI1LWNhYmMtNGQ5ZS1iMjZmLTIzMDJjMWQ4MmQ4ZiIsInBsY3kiOiJ0K1U5TkFITWJuTENRUmxWSFZOdS9WdktlaHlkdXJSblhYMWpqZGt4bzRUQVFPcVJPQkVVQkQyUGk3V2VTVVlJeUdQcndDa0JWTGJWZ05ZTkYwU0YvZjF0Ui9EYlhhUWNycCs2YXpwSTRDSGQrNlhMTlMzY09XY3JQRFFKSHNCY1I5YW9FdEZXOUxxeVdKV0YzWmtKUFpHcEJmT2l6Qmd3RlpiT0M1WT0iLCJwbGN5X2tleSI6ImtleS11c3NvLXBsY3ktMTEwOTE4LnBlbSIsInN1YiI6InNwaWZmZTovL3BlcnNvbm5lbC51cGtpLmNhL2VpZC85OTkwMDA1MDYzMTAiLCJ0ZW5hbmN5IjoidWJlci9wcm9kdWN0aW9uIiwidHlwZSI6Im9mZmxpbmUiLCJ1dWlkIjoiNjg2YzBmOWYtMzRmMy00OTgwLWE0NGYtMmU5ODVmMzY0MTg0In0.mhW0U3fW9J9HVSLjgfTfCgmgB4WF7E-C32IJ78jK5hzEK6iEE9Ng-V4-iKT0yKNq2n8-g6uG2T17-oUndeQW9w"
)

type Issue struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func createIssue(title, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", baseURL, org, repo)

	issue := Issue{
		Title: title,
		Body:  body,
	}

	jsonData, err := json.Marshal(issue)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create issue: %s", resp.Status)
	}

	fmt.Println("Issue created successfully")
	return nil
}

func main2() {
	title := "Sample Issue Title"
	body := "This is a sample issue body."

	err := createIssue(title, body)
	if err != nil {
		fmt.Printf("Error creating issue: %v\n", err)
	}
}
