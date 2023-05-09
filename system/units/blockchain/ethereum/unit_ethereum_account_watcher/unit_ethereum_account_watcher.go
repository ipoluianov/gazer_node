package unit_ethereum_account_watcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/resources"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitEthereumAccountWatcher struct {
	units_common.Unit
	addr              string
	ethAddress        string
	timeoutMs         int
	periodMs          int
	receivedVariables map[string]string
}

func New() common_interfaces.IUnit {
	var c UnitEthereumAccountWatcher
	c.receivedVariables = make(map[string]string)
	return &c
}

const (
	ItemNameStatus = "Status"
)

var Image []byte

func init() {
	Image = resources.R_files_sensors_unit_network_json_requester_png
}

func (c *UnitEthereumAccountWatcher) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("addr", "Address", "localhost:445", "string", "", "", "")
	meta.Add("ethAddress", "ethAddress", "", "string", "", "", "")
	meta.Add("period", "Period, ms", "5000", "num", "0", "999999", "0")
	meta.Add("timeout", "Timeout, ms", "1000", "num", "0", "999999", "0")
	return meta.Marshal()
}

func (c *UnitEthereumAccountWatcher) InternalUnitStart() error {
	var err error

	type Config struct {
		Addr       string  `json:"addr"`
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

	c.addr = config.Addr
	if c.addr == "" {
		err = errors.New("wrong address")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.ethAddress = config.EthAddress
	if c.ethAddress == "" {
		err = errors.New("wrong eth address")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.timeoutMs = int(config.Timeout)
	if c.timeoutMs < 100 {
		err = errors.New("wrong timeout (<100)")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}
	if c.timeoutMs > 10000 {
		err = errors.New("wrong timeout (>10000)")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.periodMs = int(config.Period)
	if c.periodMs < 100 {
		err = errors.New("wrong period")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	if c.periodMs < c.timeoutMs {
		err = errors.New("wrong period (<timeout)")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}
	if c.periodMs < 100 {
		err = errors.New("wrong period (<100)")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}
	if c.periodMs > 60000 {
		err = errors.New("wrong period (>60000)")
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
	dtLastTime := time.Now().UTC()

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

		client, err := ethclient.DialContext(context.Background(), c.addr)
		if err != nil {
			c.SetString(ItemNameStatus, err.Error(), "error")
		}
		balance, err := client.BalanceAt(context.Background(), common.HexToAddress(c.ethAddress), nil)
		if err != nil {
			c.SetString(ItemNameStatus, err.Error(), "error")
		}

		fSet := func(name string, value string, UOM string) {
			c.receivedVariables[name] = value
			c.SetString(name, value, UOM)
		}

		fSet("address", fmt.Sprint(c.ethAddress), "")
		fSet("balance", fmt.Sprint(float64(balance.Uint64())/1000000000000000000), "ETH")

		client.Close()
	}

	for vName, _ := range c.receivedVariables {
		c.SetString(vName, "", "stopped")
	}

	c.SetString(ItemNameStatus, "", "stopped")
	c.Started = false
}
