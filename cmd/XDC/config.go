// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"

	"gopkg.in/urfave/cli.v1"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/internal/debug"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(nodeFlags, rpcFlags...), whisperFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type ethstatsConfig struct {
	URL string
}

type account struct {
	Unlocks   []string
	Passwords []string
}
// whisper has been deprecated, but clients out there might still have [Shh]
// in their config, which will crash. Cut them some slack by keeping the
// config, and displaying a message that those config switches are ineffectual.
// To be removed circa Q1 2021 -- @gballet.
type whisperDeprecatedConfig struct {
	MaxMessageSize                        uint32  `toml:",omitempty"`
	MinimumAcceptedPOW                    float64 `toml:",omitempty"`
	RestrictConnectionBetweenLightClients bool    `toml:",omitempty"`
}

type gethConfig struct {
	Eth      eth.Config
	Shh      whisperDeprecatedConfig
	Node     node.Config
	Ethstats ethstatsConfig
}

type Bootnodes struct {
	Mainnet []string
	Testnet []string
}

type XDCConfig struct {
	Eth         eth.Config
	Shh         whisper.Config
	Node        node.Config
	Ethstats    ethstatsConfig
	Dashboard   dashboard.Config
	Account     account
	StakeEnable bool
	Bootnodes   Bootnodes
	Verbosity   int
	NAT         string
}

func loadConfig(file string, cfg *XDCConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "XDC.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, XDCConfig) {
	// Load defaults.
	cfg := XDCConfig{
		Eth:         eth.DefaultConfig,
		Shh:         whisper.DefaultConfig,
		Node:        defaultNodeConfig(),
		Dashboard:   dashboard.DefaultConfig,
		StakeEnable: true,
		Verbosity:   3,
		NAT:         "",
	}
	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}

		if cfg.Shh != (whisperDeprecatedConfig{}) {
			log.Warn("Deprecated whisper config detected. Whisper has been moved to github.com/ethereum/whisper")
		}
	}
	if ctx.GlobalIsSet(utils.StakingEnabledFlag.Name) {
		cfg.StakeEnable = ctx.GlobalBool(utils.StakingEnabledFlag.Name)
	}
	if !ctx.GlobalIsSet(debug.VerbosityFlag.Name) {
		debug.Glogger.Verbosity(log.Lvl(cfg.Verbosity))
	}

	if !ctx.GlobalIsSet(utils.NATFlag.Name) && cfg.NAT != "" {
		ctx.Set(utils.NATFlag.Name, cfg.NAT)
	}

	// Check testnet is enable.
	if ctx.GlobalBool(utils.XDCTestnetFlag.Name) {
		common.IsTestnet = true
	}

	// Check rollback hash exist.
	if rollbackHash := ctx.GlobalString(utils.RollbackFlag.Name); rollbackHash != "" {
		common.RollbackHash = common.HexToHash(rollbackHash)
	}

	// Check GasPrice
	common.MinGasPrice = common.DefaultMinGasPrice
	if ctx.GlobalIsSet(utils.StakerGasPriceFlag.Name) {
		if gasPrice := int64(ctx.GlobalInt(utils.StakerGasPriceFlag.Name)); gasPrice > common.DefaultMinGasPrice {
			common.MinGasPrice = gasPrice
		}
	}

	// read passwords from environment
	passwords := []string{}
	for _, env := range cfg.Account.Passwords {
		if trimmed := strings.TrimSpace(env); trimmed != "" {
			value := os.Getenv(trimmed)
			for _, info := range strings.Split(value, ",") {
				if trimmed2 := strings.TrimSpace(info); trimmed2 != "" {
					passwords = append(passwords, trimmed2)
				}
			}
		}
	}
	cfg.Account.Passwords = passwords

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetEthConfig(ctx, stack, &cfg.Eth)
	if ctx.GlobalIsSet(utils.EthStatsURLFlag.Name) {
		cfg.Ethstats.URL = ctx.GlobalString(utils.EthStatsURLFlag.Name)
	}
	utils.SetShhConfig(ctx, stack)

	return stack, cfg
}

func applyValues(values []string, params *[]string) {
	data := []string{}
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			data = append(data, trimmed)
		}
	}
	if len(data) > 0 {
		*params = data
	}

}

// enableWhisper returns true in case one of the whisper flags is set.
func checkWhisper(ctx *cli.Context) {
	for _, flag := range whisperFlags {
		if ctx.GlobalIsSet(flag.GetName()) {
			log.Warn("deprecated whisper flag detected. Whisper has been moved to github.com/ethereum/whisper")
		}
	}
}

// makeFullNode loads geth configuration and creates the Ethereum backend.
func makeFullNode(ctx *cli.Context) (*node.Node, ethapi.Backend) {
	stack, cfg := makeConfigNode(ctx)

	backend := utils.RegisterEthService(stack, &cfg.Eth)

	checkWhisper(ctx)
	// Configure GraphQL if requested
	if ctx.GlobalIsSet(utils.GraphQLEnabledFlag.Name) {
		utils.RegisterGraphQLService(stack, backend, cfg.Node)
	}
	// Add the Ethereum Stats daemon if requested.
	if cfg.Ethstats.URL != "" {
		utils.RegisterEthStatsService(stack, backend, cfg.Ethstats.URL)
	}
	return stack, backend
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.Eth.Genesis != nil {
		cfg.Eth.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}
