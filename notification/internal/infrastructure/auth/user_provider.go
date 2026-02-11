package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"notification/internal/domain"
	"notification/internal/port"
	"strconv"
	"strings"
)

type UserProvider struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewUserProvider(baseURL, token string) *UserProvider {
	return &UserProvider{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{},
	}
}

func (p *UserProvider) GetByIDs(ctx context.Context, ids []int) ([]domain.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	url := fmt.Sprintf("%s/users?ids=%s", p.baseURL, strings.Join(parts, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users fetch status: %d", resp.StatusCode)
	}

	var raw []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(raw))
	for _, item := range raw {
		id, _ := item["id"].(float64)
		email, _ := item["email"].(string)
		name, _ := item["name"].(string)
		users = append(users, domain.User{
			ID:    int(id),
			Email: email,
			Name:  name,
		})
	}
	return users, nil
}

var _ port.UserProvider = (*UserProvider)(nil)
