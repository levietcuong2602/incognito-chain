package config

import (
	"encoding/json"
	"log"
	"path/filepath"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/utils"
	"github.com/spf13/viper"
)

var unifiedToken map[uint64]map[common.Hash]map[common.Hash]Vault

func UnifiedToken() map[uint64]map[common.Hash]map[common.Hash]Vault {
	return unifiedToken
}

//AbortUnifiedToken use for unit test only
// DO NOT use this function for development process
func AbortUnifiedToken() {
	unifiedToken = make(map[uint64]map[common.Hash]map[common.Hash]Vault)
}

// SetUnifiedToken
// DO NOT use this function for development process
func SetUnifiedToken(UnifiedToken map[uint64]map[common.Hash]map[common.Hash]Vault) {
	unifiedToken = UnifiedToken
}

type Vault struct {
	ExternalDecimal uint8   `mapstructure:"external_decimal"`
	ExternalTokenID string `mapstructure:"external_token_id"`
	NetworkID       uint8   `mapstructure:"network_id"`
}

func LoadUnifiedToken(data []byte) {
	unifiedToken = make(map[uint64]map[common.Hash]map[common.Hash]Vault)
	temp := make(map[string]map[string]map[string]Vault)
	network := c.Network()
	//read config from file
	viper.SetConfigName(utils.GetEnv(UnifiedTokenFileKey, DefaultUnifiedTokenFile))           // name of config file (without extension)
	viper.SetConfigType(utils.GetEnv(ConfigFileTypeKey, DefaultUnifiedTokenFileType))         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(filepath.Join(utils.GetEnv(ConfigDirKey, DefaultConfigDir), network)) // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			if err != nil {
				panic(err)
			}
		} else { //if file not found
			log.Println("Using default unified token for " + network)
			json.Unmarshal(data, &temp)
		}
	} else {
		err = viper.Unmarshal(&temp)
		if err != nil {
			panic(err)
		}
	}
	if _, found := temp[common.PRVIDStr]; found {
		panic("Found PRV in list unified token config")
	}
	for beaconHeightStr, unifiedTokens := range temp {
		beaconHeight, err := strconv.ParseUint(beaconHeightStr, 10, 64)
		if err != nil {
			panic(err)
		}
		unifiedToken[beaconHeight] = make(map[common.Hash]map[common.Hash]Vault)
		for unifiedTokenID, vaults := range unifiedTokens {
			unifiedTokenHash, err := common.Hash{}.NewHashFromStr(unifiedTokenID)
			if err != nil {
				panic(err)
			}
			unifiedToken[beaconHeight][*unifiedTokenHash] = make(map[common.Hash]Vault)
			for tokenIDStr, vault := range vaults {
				tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
				if err != nil {
					panic(err)
				}
				unifiedToken[beaconHeight][*unifiedTokenHash][*tokenID] = vault
			}
		}
	}
}

var mainnetUnifiedToken = []byte{}
var testnet1UnifiedToken = []byte{}
var testnet2UnifiedToken = []byte{}
var localUnifiedToken = []byte{}
