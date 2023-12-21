package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config parameters
type Config struct {
	Address        string        `json:"address"`
	DriverName     string        `json:"driverName"`
	StoragePath    string        `json:"storagePath"`
	InitSQL        string        `json:"initSQL"`
	MockDataSQL    string        `json:"mockDataSQL"`
	MaxHeaderBytes int           `json:"maxHeaderBytes"`
	CertFile       string        `json:"certFile"`
	KeyFile        string        `json:"keyFile"`
	IdleTimeout    time.Duration `json:"idleTimeout"`
	ReadTimeout    time.Duration `json:"readTimeout"`
	WriteTimeout   time.Duration `json:"writeTimeout"`
	RateLimit      int           `json:"rateLimit"`
}

func ReadConfig() (*Config, error) {
	configFile, err := os.ReadFile("config/config.json")
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	err = os.RemoveAll("images")
	if err != nil {
		err = os.Mkdir("images", os.ModePerm)
		if err != nil {
			return nil, err
		}
	} else {
		err = os.Mkdir("images", os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return &config, nil
}
