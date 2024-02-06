package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type AzureTokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

type PasswordCredentials struct {
	App           Application
	DisplayName   string    `json:"displayName"`
	StartDateTime time.Time `json:"startDateTime"`
	EndDateTime   time.Time `json:"endDateTime"`
}

type Application struct {
	AppId               string                `json:"appId"`
	DisplayName         string                `json:"displayName"`
	Description         string                `json:"description"`
	PasswordCredentials []PasswordCredentials `json:"passwordCredentials"`
}

type Entity interface {
	Application
}

type ApiResponse[E Entity] struct {
	Context  string `json:"@odata.context"`
	NextLink string `json:"@odata.nextLink"`
	Value    []E    `json:"value"`
}

type AzureAuth struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
}

func SimpleAzureAuth(clientId string, clientSecret string) AzureAuth {
	return AzureAuth{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Scope:        "https://graph.microsoft.com/.default",
	}
}

type AzureClient struct {
	TenantId    string
	Auth        AzureAuth
	AccessToken string
}

func NewAzureClient(tenantId string, auth AzureAuth) *AzureClient {
	return &AzureClient{
		TenantId: tenantId,
		Auth:     auth,
	}
}

func (az *AzureClient) Authenticate() error {
	resp, err := http.PostForm(fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", az.TenantId),
		url.Values{
			"grant_type":    {"client_credentials"},
			"scope":         {az.Auth.Scope},
			"client_id":     {az.Auth.ClientId},
			"client_secret": {az.Auth.ClientSecret},
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return errors.New(string(body))
	}

	var token AzureTokenResponse

	if err := json.Unmarshal(body, &token); err != nil {
		return errors.New(fmt.Sprintf("Could not unmarshall authentication response: %s", err))
	}

	az.AccessToken = token.AccessToken

	return nil
}

func (az *AzureClient) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+az.AccessToken)

	return req, nil
}

func (az *AzureClient) ListApplications() ([]Application, error) {
	req, _ := az.newRequest("GET", "https://graph.microsoft.com/v1.0/applications/", nil)

	c := http.Client{}
	resp, err := c.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var apps ApiResponse[Application]
	json.Unmarshal(body, &apps)

	return apps.Value, nil
}

func FlattenAndSortSecrets(apps []Application) []PasswordCredentials {
	var secrets []PasswordCredentials
	for _, app := range apps {
		for _, sec := range app.PasswordCredentials {
			sec.App = app
			secrets = append(secrets, sec)
		}
	}

	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].EndDateTime.Before(secrets[j].EndDateTime)
	})

	return secrets
}
