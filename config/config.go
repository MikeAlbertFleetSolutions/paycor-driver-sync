package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	msgMissingField = "required configuration missing %s"

	// unwrapped config values
	Paycor paycor
)

type configuration struct {
	Paycor paycor
}

func (c *configuration) validate() error {
	var err error

	err = c.Paycor.validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

type paycor struct {
	TenantID int32
	Endpoint string
	OAuth    struct {
		ClientID     string
		ClientSecret string
		RefreshToken string
	}
	APImSubscriptionKey string
}

func (p *paycor) validate() error {
	if p.TenantID == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor TenantID")
		log.Printf("%+v", err)
		return err
	}
	if len(p.Endpoint) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor Endpoint")
		log.Printf("%+v", err)
		return err
	}
	if len(p.OAuth.ClientID) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor OAuth ClientID")
		log.Printf("%+v", err)
		return err
	}
	if len(p.OAuth.ClientSecret) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor OAuth ClientSecret")
		log.Printf("%+v", err)
		return err
	}
	if len(p.OAuth.RefreshToken) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor OAuth RefreshToken")
		log.Printf("%+v", err)
		return err
	}
	if len(p.APImSubscriptionKey) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor APImSubscriptionKey")
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// FromFile reads the application configuration from file configFile
func FromFile(configFile string) error {
	// read config
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var c configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// validation
	err = c.validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	Paycor = c.Paycor

	return nil
}

// Write writes configuration to the file configFile
func Write(configFile string) error {
	// wrap
	c := configuration{
		Paycor: Paycor,
	}

	// make sure valid before proceeding
	err := c.validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// create YAML to write
	b, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// write out to file
	err = os.WriteFile(configFile, b, 0600)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}
