package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func GetUsersData(ids []int) ([]User, error) {
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	authToken := os.Getenv("AUTH_SERVICE_TOKEN")
	if len(ids) == 0 {
		return nil, fmt.Errorf("missing ids parameter")
	}

	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = strconv.Itoa(id)
	}

	url := fmt.Sprintf("%s/users?ids=%s", authServiceURL, strings.Join(idStrings, ","))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request")
	}

	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users data")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch users data, status: %d", resp.StatusCode)
	}

	var responseData []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to parse response")
	}

	return ParseUsersData(responseData)
}

func ParseUsersData(data []map[string]interface{}) ([]User, error) {
	var users []User

	for _, item := range data {
		id, ok := item["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid id format")
		}
		email, ok := item["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid email format")
		}
		name, ok := item["name"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid name format")
		}
		users = append(users, User{
			ID:    int(id),
			Email: email,
			Name:  name,
		})
	}

	return users, nil
}
