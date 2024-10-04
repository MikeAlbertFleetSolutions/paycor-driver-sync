package paycor

import (
	"bytes"
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MikeAlbertFleetSolutions/go-paycor"
)

type DriverHomeAddress struct {
	EmployeeNumber string
	LastName       string
	FirstName      string
	Address1       string
	Address2       string
	City           string
	State          string
	ZIPCode        string
}

// Client is our type
type Client struct {
	paycor     *paycor.Client
	httpClient *http.Client
}

// NewClient creates a new client for the Paycor API
func NewClient(publicKey, privateKey, host string) (*Client, error) {
	// connect to paycor
	paycorClient := paycor.NewClient(publicKey, privateKey, host)

	client := &Client{
		paycor: paycorClient,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	return client, nil
}

// onlyNums returns only the numbers from a string as a string
func onlyNums(s string) string {
	bs := []byte(s)
	j := 0
	for _, b := range bs {
		if '0' <= b && b <= '9' {
			bs[j] = b
			j++
		}
	}
	return string(s[:j])
}

// GetDriverHomeAddresses gets the driver home addresses from Paycor
func (client *Client) GetDriverHomeAddresses(homeAddressesReport string) ([]DriverHomeAddress, error) {
	// get report from paycor
	report, err := client.paycor.GetReportByName(homeAddressesReport)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	// open csv
	r := csv.NewReader(bytes.NewReader(report))
	r.ReuseRecord = true

	numRows := 0
	var driverHomeAddresses []DriverHomeAddress
	for {
		// read row
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}
		numRows++

		// skip header
		if numRows == 1 {
			continue
		}

		driverHomeAddresses = append(driverHomeAddresses, DriverHomeAddress{
			EmployeeNumber: onlyNums(record[0]),
			LastName:       strings.TrimSpace(record[1]),
			FirstName:      strings.TrimSpace(record[2]),
			Address1:       strings.TrimSpace(record[3]),
			Address2:       strings.TrimSpace(record[4]),
			City:           strings.TrimSpace(record[5]),
			State:          strings.TrimSpace(record[6]),
			ZIPCode:        strings.TrimSpace(record[7]),
		})
	}

	return driverHomeAddresses, nil
}
