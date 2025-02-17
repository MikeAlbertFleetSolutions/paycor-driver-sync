package mikealbert

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Address struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	PostCode string `json:"postCode"`
}

type Driver struct {
	Address        Address `json:"address"`
	DriverId       *int    `json:"drvId,omitempty"`
	EmployeeNumber *string `json:"employeeNumber,omitempty"`
}

// Client is our type
type Client struct {
	clientId     string
	clientSecret string
	endpoint     string

	httpClient     *http.Client
	authentication struct {
		accessToken string
		expires     time.Time
	}

	prevRequest time.Time
}

// return first n characters of a string
func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

// NewClient creates a new mikealbert client
func NewClient(clientId, clientSecret, endpoint string) (*Client, error) {
	client := &Client{
		clientId:     clientId,
		clientSecret: clientSecret,
		endpoint:     endpoint,
		httpClient: &http.Client{
			Timeout: time.Second * 60,
		},
	}

	err := client.authenticate(clientId, clientSecret)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return client, nil
}

// makeRequest is a helper function to wrap making REST calls to mike albert
func (client *Client) makeRequest(method, url string, body io.Reader) ([]byte, error) {
	// mike albert does rate limiting
	// make sure it has been at least 500 milliseconds since the last call
	elapse := time.Since(client.prevRequest)
	if elapse < 500*time.Millisecond {
		time.Sleep(500 * time.Millisecond)
	}
	defer func() { client.prevRequest = time.Now() }()

	// create request
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	request.Header.Set("Accept", "application/json")

	// set content-type only on requests that send some content
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	// add authentication token
	if len(client.authentication.accessToken) > 0 {
		// need to re-authenticate?
		if !client.authentication.expires.After(time.Now().UTC().Add(5 * time.Minute)) {
			err := client.authenticate(client.clientId, client.clientSecret)
			if err != nil {
				log.Printf("%+v", err)
				return nil, err
			}
		}

		if len(client.authentication.accessToken) > 0 {
			request.Header.Add("Authorization", client.authentication.accessToken)
		}
	}

	// make request, get response
	var response *http.Response
	response, err = client.httpClient.Do(request)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	defer response.Body.Close()

	// get body for caller, if there is something
	var data []byte
	if response.ContentLength != 0 {
		data, err = io.ReadAll(response.Body)
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}
	}

	// error?
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		var r ErrorResponse
		if len(data) > 0 {
			err = json.Unmarshal(data, &r)
			if err != nil {
				log.Printf("%+v", err)
				return nil, err
			}
		} else {
			r.Message = "<no message>"
		}

		err = fmt.Errorf("%s call to %s returned status code %d, message: %s", method, url, response.StatusCode, r.Message)
		log.Printf("%+v", err)
		return nil, err
	}

	return data, nil
}

// helper function to authenticate against mikealbert API
func (client *Client) authenticate(clientId, clientSecret string) error {
	req := struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}

	ab, err := json.Marshal(req)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	u, err := url.JoinPath(client.endpoint, "token")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	b, err := client.makeRequest("POST", u, strings.NewReader(string(ab)))
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var resp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	client.authentication.accessToken = resp.TokenType + " " + resp.AccessToken
	client.authentication.expires = time.Now().UTC().Add(time.Duration(resp.ExpiresIn) * time.Second)

	return nil
}

// Find drivers by employeeNumber
func (client *Client) FindDrivers(employeeNumber string) ([]Driver, error) {
	req := struct {
		EmployeeNumber string `json:"employeeNumber"`
	}{
		EmployeeNumber: employeeNumber,
	}

	ab, err := json.Marshal(req)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	u, err := url.JoinPath(client.endpoint, "driver-management/driver/find")
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	b, err := client.makeRequest("POST", u, strings.NewReader(string(ab)))
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	var resp []Driver
	err = json.Unmarshal(b, &resp)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return resp, nil
}

// Update driver by driver ID
func (client *Client) UpdateDriver(driverId int, address1, address2, postCode string) (*Driver, error) {
	req := Driver{
		Address: Address{
			Address1: address1,
			Address2: address2,
			PostCode: firstN(postCode, 5),
		},
	}

	ab, err := json.Marshal(req)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	u, err := url.JoinPath(client.endpoint, "driver-management/driver", strconv.Itoa(driverId))
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	b, err := client.makeRequest("POST", u, strings.NewReader(string(ab)))
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	var resp Driver
	err = json.Unmarshal(b, &resp)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return &resp, nil
}
