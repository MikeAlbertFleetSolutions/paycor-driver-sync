package paycor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type paycor struct {
	Endpoint            string
	ClientID            string
	ClientSecret        string
	RefreshToken        string
	APImSubscriptionKey string

	authorization string
}

// Client is our type
type Client struct {
	Paycor     paycor
	httpClient *http.Client
}

// NewClient creates a new client for the Paycor APIs, and gets a refresh authorization token
// Paycor.RefreshToken could be updated after a call to this function, caller needs to persist it for next call to NewClient
func NewClient(endpoint, clientid, clientsecret, refreshtoken, apimsubscriptionkey string) (*Client, error) {
	client := &Client{
		Paycor: paycor{
			Endpoint:            endpoint,
			ClientID:            clientid,
			ClientSecret:        clientsecret,
			RefreshToken:        refreshtoken,
			APImSubscriptionKey: apimsubscriptionkey,
		},
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	err := client.refreshToken()
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return client, nil
}

// makeRequest is a helper function to wrap making REST calls to the Paycor APIs
func (client *Client) makeRequest(method, path string, params url.Values, body io.Reader) ([]byte, error) {
	// form request url
	urlStr, err := url.JoinPath(client.Paycor.Endpoint, path)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	// create request
	request, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+client.Paycor.authorization)
	request.Header.Set("Ocp-Apim-Subscription-Key", client.Paycor.APImSubscriptionKey)

	// reqDump, err := httputil.DumpRequestOut(request, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("REQUEST:\n%s\n", string(reqDump))

	// make request, get response
	var response *http.Response
	response, err = client.httpClient.Do(request)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	defer response.Body.Close()

	// resDump, err := httputil.DumpResponse(response, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("RESPONSE:\n%s\n", string(resDump))

	// get body
	var b []byte
	if response.ContentLength != 0 {
		b, err = io.ReadAll(response.Body)
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}
	}

	// error?
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		err = fmt.Errorf("%s call to %s returned status code %d: %s", method, path, response.StatusCode, string(b))
		log.Printf("%+v", err)
		return nil, err
	}

	return b, nil
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type refreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int32  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

func (client *Client) refreshToken() error {
	// APImSubscriptionKey as query parameter
	params := url.Values{
		"subscription-key": []string{client.Paycor.APImSubscriptionKey},
	}

	// refreshTokenRequest as body
	reqBody, err := json.Marshal(
		refreshTokenRequest{
			RefreshToken: client.Paycor.RefreshToken,
			ClientID:     client.Paycor.ClientID,
			ClientSecret: client.Paycor.ClientSecret,
		})
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	b, err := client.makeRequest("POST", "/authenticationsupport/retrieveAccessTokenWithRefreshToken", params, bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var r refreshTokenResponse
	err = json.Unmarshal(b, &r)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// set session variables
	client.Paycor.authorization = r.AccessToken
	client.Paycor.RefreshToken = r.RefreshToken

	return nil
}

type getEmployeesByTenantIDResponse struct {
	HasMoreResults       string `json:"HasMoreResults"`
	ContinuationToken    string `json:"ContinuationToken"`
	AdditionalResultsURL string `json:"AdditionalResultsUrl"`
	Records              []struct {
		ID             string `json:"Id"`
		EmployeeNumber int    `json:"EmployeeNumber"`
		FirstName      string `json:"FirstName"`
		MiddleName     string `json:"MiddleName"`
		LastName       string `json:"LastName"`
		Employee       struct {
			ID  string `json:"Id"`
			URL string `json:"Url"`
		} `json:"Employee"`
	} `json:"Records"`
}

func (client *Client) GetEmployeesByTenantID(tenantID int32) ([]byte, error) {
	b, err := client.makeRequest("GET", fmt.Sprintf("/tenants/%d/employees", tenantID), nil, nil)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	var r getEmployeesByTenantIDResponse
	err = json.Unmarshal(b, &r)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	log.Printf("%+v", r)

	return b, nil
}
