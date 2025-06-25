package config

import (
	"errors"

	corecfg "github.com/Axway/agent-sdk/pkg/config"
)

// AgentConfig - represents the config for agent
type AgentConfig struct {
	CentralCfg corecfg.CentralConfig `config:"central"`
	GatewayCfg *GatewayConfig        `config:"gateway"`
}

// GatewayConfig - represents the config for gateway
type GatewayConfig struct {
	corecfg.IConfigValidator
	corecfg.IResourceConfigCallback
	LogFile string `config:"logFile"`
}

// ValidateCfg - Validates the gateway config
func (c *GatewayConfig) ValidateCfg() (err error) {
	if c.LogFile == "" {
		return errors.New("Invalid gateway configuration: logFile is not configured")
	}

	return
}
