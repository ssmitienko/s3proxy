package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"regexp"
	"time"
)

type locationConfiguration struct {
	Prefix                 string
	StorageEndpoint        string
	StorageAccessKey       string
	StorageSecretAccessKey string
	StorageBucketName      string
	StorageBucketLocation  string
	StorageUseSSL          bool
	RegExpMatch            string //URL rewrting logic
	RegExpSub              string
	Expires                string //time to add to current time for expires
	CacheControl           string //Cache control header
	DropParams             bool   //Drop ?qweqweqw=qweq
	prefix                 *regexp.Regexp
	translation            *regexp.Regexp
	expires                time.Duration
}

type serviceConfiguration struct {
	Locations []locationConfiguration
}

var configuration serviceConfiguration

func configInit(configFile string) error {

	file, err := os.Open(configFile)
	if err != nil {
		log.Println("Failed to read configuration:", err)
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Println("Failed to parse configuration", err)
		return err
	}

	if len(configuration.Locations) < 1 {
		log.Println("No locations")
		return errors.New("no locatoins")
	}

	for i := 0; i < len(configuration.Locations); i++ {

		configuration.Locations[i].prefix, err = regexp.Compile(configuration.Locations[i].Prefix)
		if err != nil {
			log.Println("Failed to compile regexp", configuration.Locations[i].Prefix, err)
			return err
		}

		if len(configuration.Locations[i].RegExpMatch) > 0 {
			configuration.Locations[i].translation, err = regexp.Compile(configuration.Locations[i].RegExpMatch)
			if err != nil {
				log.Println("Failed to compile regexp", configuration.Locations[i].RegExpMatch, err)
				return err
			}
		}

		if len(configuration.Locations[i].Expires) > 0 {
			configuration.Locations[i].expires, err = time.ParseDuration(configuration.Locations[i].Expires)
			if err != nil {
				log.Println("Failed to parse time duration", configuration.Locations[i].Expires, err)
				return err
			}
		}
	}

	return nil
}
