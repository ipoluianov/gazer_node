package blockchain_ethereum_balance_alfa

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/uom"
)

type UnitEthereumAccountWatcher struct {
	units_common.Unit
	rpcUrl            string
	uom               string
	ethAddress        string
	periodMs          int
	receivedVariables map[string]string
}

func New() iunit.IUnit {
	var c UnitEthereumAccountWatcher
	c.receivedVariables = make(map[string]string)
	return &c
}

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Blockchain.Ethereum.Balance.Alfa"
	info.Category = "blockchain"
	info.DisplayName = "ETH Balance"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = "ETH Balance"
	return info
}

const (
	ItemNameStatus = "Status"
)

//go:embed "image.png"
var Image []byte

func (c *UnitEthereumAccountWatcher) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("rpcUrl", "RPC URL", "", "string", "", "", "")
	meta.Add("uom", "Currency", "ETH", "string", "", "", "")
	meta.Add("ethAddress", "ETH Address (0x...)", "", "string", "", "", "")
	meta.Add("period", "Period, ms", "5000", "num", "0", "3600000", "0")
	return meta.Marshal()
}

func (c *UnitEthereumAccountWatcher) InternalUnitStart() error {
	var err error

	type Config struct {
		RpcUrl     string  `json:"rpcUrl"`
		UOM        string  `json:"uom"`
		EthAddress string  `json:"ethAddress"`
		Timeout    float64 `json:"timeout"`
		Period     float64 `json:"period"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.rpcUrl = config.RpcUrl
	if c.rpcUrl == "" {
		err = errors.New("wrong rpc url")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.uom = config.UOM

	c.ethAddress = config.EthAddress
	if c.ethAddress == "" {
		err = errors.New("wrong eth address")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.periodMs = int(config.Period)
	if c.periodMs < 100 {
		err = errors.New("wrong period")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	if c.periodMs < 100 {
		err = errors.New("wrong period (<100)")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}
	if c.periodMs > 3600000 {
		err = errors.New("wrong period (>3600000)")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.receivedVariables = make(map[string]string)

	c.SetMainItem(ItemNameStatus)

	c.SetString(ItemNameStatus, "", "")
	go c.Tick()
	return nil
}

func (c *UnitEthereumAccountWatcher) InternalUnitStop() {
}

func (c *UnitEthereumAccountWatcher) Tick() {
	// var err error
	c.Started = true
	dtLastTime := time.Now().UTC().Add(-time.Duration(c.periodMs) * time.Millisecond)

	for !c.Stopping {
		for {
			if c.Stopping || time.Now().Sub(dtLastTime) > time.Duration(c.periodMs)*time.Millisecond {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			c.SetString(ItemNameStatus, "stopped", "")
			break
		}
		dtLastTime = time.Now().UTC()

		client, err := ethclient.DialContext(context.Background(), c.rpcUrl)
		if err != nil {
			c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
			for vName, _ := range c.receivedVariables {
				c.SetString(vName, "", uom.ERROR)
			}
			continue
		}
		balance, err := client.BalanceAt(context.Background(), common.HexToAddress(c.ethAddress), nil)
		if err != nil {
			c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
			for vName, _ := range c.receivedVariables {
				c.SetString(vName, "", uom.ERROR)
			}
			continue
		}

		fSet := func(name string, value string, UOM string) {
			c.receivedVariables[name] = value
			c.SetString(name, value, UOM)
		}

		fSet("address", fmt.Sprint(c.ethAddress), "")
		fSet("balance", fmt.Sprint(float64(balance.Uint64())/1000000000000000000), c.uom)

		c.SetString(ItemNameStatus, "ok", "")

		client.Close()
	}

	for vName, _ := range c.receivedVariables {
		c.SetString(vName, "", uom.STOPPED)
	}

	c.SetString(ItemNameStatus, "", uom.STOPPED)
	c.Started = false
}
