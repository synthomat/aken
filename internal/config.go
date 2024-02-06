package internal

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
	"log"
)

type AzureConfig struct {
	TenantId     string `hcl:"tenant_id"`
	ClientId     string `hcl:"client_id"`
	ClientSecret string `hcl:"client_secret"`
}

type ChecksConfig struct {
	ThresholdDays []int  `hcl:"threshold_days"`
	ScheduleCron  string `hcl:"schedule_cron"`
}

type Provider struct {
	Type   string               `hcl:"type,label"`
	Values map[string]cty.Value `hcl:",remain"`
}
type NotificationsConfig struct {
	Providers []Provider `hcl:"provider,block"`
}

type Config struct {
	Azure         AzureConfig          `hcl:"azure,block"`
	Checks        ChecksConfig         `hcl:"checks,block"`
	Notifications *NotificationsConfig `hcl:"notifications,block"`
}

func MustLoadConfig(filePath string) *Config {
	var config Config
	err := hclsimple.DecodeFile(filePath, nil, &config)

	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
		return nil // will never return
	}

	return &config
}
