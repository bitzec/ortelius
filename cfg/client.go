// (c) 2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cfg

import (
	"github.com/ava-labs/gecko/ids"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	configKeysFilter    = "filter"
	configKeysKafka     = "kafka"
	configKeysChainID   = "chainID"
	configKeysDataType  = "dataType"
	configKeysIPCURL    = "ipcURL"
	configKeysNetworkID = "networkID"
)

type FilterConfig struct {
	Min uint32
	Max uint32
}

// ClientConfig manages configuration data for the client app
type ClientConfig struct {
	Context string
	Filter  FilterConfig

	ServiceConfig
	Kafka *kafka.ConfigMap

	// Chain-specific configs. These tie the client to single chain.
	ChainID   ids.ID
	DataType  string
	NetworkID uint32
	IPCURL    string
}

// NewClientConfig returns a *ClientConfig populated with data from the given file
func NewClientConfig(context string, file string) (*ClientConfig, error) {
	// Parse config file with viper and set defaults
	v, err := getConfigViper(file, map[string]interface{}{
		configKeysIPCURL:    "",
		configKeysChainID:   "",
		configKeysDataType:  "avm",
		configKeysNetworkID: 12345,
		configKeysFilter: map[string]interface{}{
			"max": 1073741824,
			"min": 2147483648,
		},
		configKeysKafka: map[string]interface{}{
			"client.id":          "avm",
			"enable.idempotence": true,
			"bootstrap.servers":  "127.0.0.1:9092",
		},
	})
	if err != nil {
		return nil, err
	}

	// Parse chainID string
	chainID, err := ids.FromString(v.GetString(configKeysChainID))
	if err != nil {
		return nil, err
	}

	// Get services config
	serviceConf, err := getServiceConfig(v)
	if err != nil {
		return nil, err
	}

	filterConf := getSubViper(v, configKeysFilter)

	// Collect config data into a ClientConfig object
	return &ClientConfig{
		Context: context,
		Filter: FilterConfig{
			Min: filterConf.GetUint32("min"),
			Max: filterConf.GetUint32("min"),
		},

		ServiceConfig: serviceConf,
		Kafka:         getKafkaConf(v.GetStringMap(configKeysKafka)),

		ChainID:   chainID,
		DataType:  v.GetString(configKeysDataType),
		IPCURL:    v.GetString(configKeysIPCURL),
		NetworkID: v.GetUint32(configKeysNetworkID),
	}, nil
}

func getKafkaConf(conf map[string]interface{}) *kafka.ConfigMap {
	kc := kafka.ConfigMap{}
	for k, v := range conf {
		kc[k] = v
	}
	return &kc
}
