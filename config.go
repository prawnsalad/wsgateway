package main

import "gopkg.in/yaml.v3"

type Config struct {
	ListenAddr string `yaml:"listen_addr"`
	ConnectionRedisSync struct {
		Addr string `yaml:"addr"`
	} `yaml:"connection_redis_sync"`
	StreamRedis struct {
		Addr string `yaml:"addr"`
	} `yaml:"stream_redis"`
	Endpoints []struct {
		Path string `yaml:"path"`
		SetTags map[string]string `yaml:"set_tags"`
		StreamIncludeTags []string `yaml:"stream_include_tags"`
	} `yaml:"endpoints"` 
}

func loadConfig(strConf string) (*Config, error) {
	newConfig := &Config{}

	err := yaml.Unmarshal([]byte(strConf), newConfig)
	if err != nil {
		return nil, err
	}

	// TODO: Validate this config before it gets returned
	return newConfig, nil
}
