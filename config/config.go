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
	Paycor     paycor
	MikeAlbert mikealbert
)

type configuration struct {
	Paycor     paycor
	MikeAlbert mikealbert
}

func (c *configuration) validate() error {
	var err error

	err = c.Paycor.validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = c.MikeAlbert.validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

type paycor struct {
	PublicKey           string
	PrivateKey          string
	Host                string
	HomeAddressesReport string
}

func (p *paycor) validate() error {
	if len(p.PublicKey) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor PublicKey")
		log.Printf("%+v", err)
		return err
	}
	if len(p.PrivateKey) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor PrivateKey")
		log.Printf("%+v", err)
		return err
	}
	if len(p.Host) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor Host")
		log.Printf("%+v", err)
		return err
	}
	if len(p.HomeAddressesReport) == 0 {
		err := fmt.Errorf(msgMissingField, "Paycor HomeAddressesReport")
		log.Printf("%+v", err)
		return err
	}

	return nil
}

type mikealbert struct {
	ClientId     string
	ClientSecret string
	Endpoint     string
}

func (m *mikealbert) validate() error {
	if len(m.ClientId) == 0 {
		err := fmt.Errorf(msgMissingField, "Mike Albert ClientId")
		log.Printf("%+v", err)
		return err
	}
	if len(m.ClientSecret) == 0 {
		err := fmt.Errorf(msgMissingField, "Mike Albert ClientSecret")
		log.Printf("%+v", err)
		return err
	}
	if len(m.Endpoint) == 0 {
		err := fmt.Errorf(msgMissingField, "Mike Albert Endpoint")
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
	MikeAlbert = c.MikeAlbert

	return nil
}

// Write writes configuration to the file configFile
func Write(configFile string) error {
	// wrap
	c := configuration{
		Paycor:     Paycor,
		MikeAlbert: MikeAlbert,
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
