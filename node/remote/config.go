package remote

import (
	"fmt"
)

// Details represents a node details for a remote node
type Details struct {
	RPC *RPCConfig `yaml:"rpc"`
}

func NewDetails(rpc *RPCConfig) *Details {
	return &Details{
		RPC: rpc,
	}
}

func DefaultDetails() *Details {
	return NewDetails(DefaultRPCConfig())
}

// Validate implements node.Details
func (d *Details) Validate() error {
	if d.RPC == nil {
		return fmt.Errorf("rpc config cannot be null")
	}

	return nil
}

// --------------------------------------------------------------------------------------------------------------------

// RPCConfig contains the configuration for the RPC endpoint
type RPCConfig struct {
	ClientName     string `yaml:"client_name"`
	Address        string `yaml:"address"`
	MaxConnections int    `yaml:"max_connections"`
}

// NewRPCConfig allows to build a new RPCConfig instance
func NewRPCConfig(clientName, address string, maxConnections int) *RPCConfig {
	return &RPCConfig{
		ClientName:     clientName,
		Address:        address,
		MaxConnections: maxConnections,
	}
}

// DefaultRPCConfig returns the default instance of RPCConfig
func DefaultRPCConfig() *RPCConfig {
	return NewRPCConfig("juno", "http://localhost:26657", 20)
}
