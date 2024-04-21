package main

import (
	"log"
	"net"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr string `yaml:"listen_addr"`
	ConnectionRedisSync struct {
		Addr string `yaml:"addr"`
	} `yaml:"connection_redis_sync"`
	StreamRedis struct {
		Addr string `yaml:"addr"`
	} `yaml:"stream_redis"`
	MaxMessageSizeKb int `yaml:"max_message_size_kb,omitempty"`
	InternalEndpointWhitelist []string `yaml:"internal_endpoint_access_whitelist"`
	InternalEndpointWhitelistInet []net.IPNet
	Endpoints []struct {
		Path string `yaml:"path"`
		SetTags map[string]string `yaml:"set_tags"`
		StreamIncludeTags []string `yaml:"stream_include_tags"`
		MaxMessageSizeKb int `yaml:"max_message_size_kb,omitempty"`
	} `yaml:"endpoints"`
	Prometheus struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"prometheus"`
}

func loadConfig(strConf string) (*Config, error) {
	newConfig := &Config{}

	err := yaml.Unmarshal([]byte(strConf), newConfig)
	if err != nil {
		return nil, err
	}

	cleanConfigErr := cleanConfig(newConfig)
	if cleanConfigErr != nil {
		return nil, cleanConfigErr
	}

	return newConfig, nil
}

func loadConfigFromFile(filePath string) (*Config, error) {
    rawFile, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    return loadConfig(string(rawFile))
}

// Ensures the required fields are set and sets any defaults
func cleanConfig(config *Config) error {
	if strings.Trim(config.ListenAddr, " \t") == "" {
		config.ListenAddr = "0.0.0.0:8080"
	}
	if config.MaxMessageSizeKb == 0 {
		// 128mb default
		config.MaxMessageSizeKb = 1024*128
	}

	// Only allow localhost access to internal endpoints by default
	if len(config.InternalEndpointWhitelist) == 0 {
		config.InternalEndpointWhitelist = []string{"127.0.0.1/8", "::1/128"}
	}

	config.InternalEndpointWhitelistInet = []net.IPNet{}
	for _, cidr := range config.InternalEndpointWhitelist {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("Error parsing internal endpoint whitelist range %s: %s", cidr, err)
			continue
		}
		config.InternalEndpointWhitelistInet = append(config.InternalEndpointWhitelistInet, *ipNet)
	}

	for _, endpoint := range config.Endpoints {
		if endpoint.Path == "" {
			endpoint.Path = "/ws"
		}
		if endpoint.SetTags == nil {
			endpoint.SetTags = map[string]string{}
		}
		if endpoint.StreamIncludeTags == nil {
			endpoint.StreamIncludeTags = []string{}
		}
	}

	return nil
}
