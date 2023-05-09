package unit_ethereum_realtime_stat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/resources"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitEthereumRealTimeStat struct {
	units_common.Unit
	addr              string
	timeoutMs         int
	periodMs          int
	receivedVariables map[string]string
}

func New() common_interfaces.IUnit {
	var c UnitEthereumRealTimeStat
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

func (c *UnitEthereumRealTimeStat) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("addr", "Address", "localhost:445", "string", "", "", "")
	meta.Add("period", "Period, ms", "5000", "num", "0", "999999", "0")
	meta.Add("timeout", "Timeout, ms", "1000", "num", "0", "999999", "0")
	return meta.Marshal()
}

func (c *UnitEthereumRealTimeStat) InternalUnitStart() error {
	var err error

	type Config struct {
		Addr    string  `json:"addr"`
		Timeout float64 `json:"timeout"`
		Period  float64 `json:"period"`
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

func (c *UnitEthereumRealTimeStat) InternalUnitStop() {
}

func (c *UnitEthereumRealTimeStat) Tick() {
	// var err error
	c.Started = true
	dtLastTime := time.Now().UTC()

	lastBlock := uint64(0)

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
		block, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			c.SetString(ItemNameStatus, err.Error(), "error")
		}

		fSet := func(name string, value string, UOM string) {
			c.receivedVariables[name] = value
			c.SetString(name, value, UOM)
		}

		if block.Header().Number.Uint64() != lastBlock {
			lastBlock = block.Header().Number.Uint64()
			fSet("blockNumber", fmt.Sprint(block.Header().Number.Uint64()), "")
			fSet("transactionCount", fmt.Sprint(block.Transactions().Len()), "")
			gasPriceAvg := float64(0)
			gasPriceMin := float64(100000000000000)
			gasPriceMax := float64(0)
			value := float64(0)
			for _, tr := range block.Body().Transactions {
				gasPrice := float64(tr.GasPrice().Uint64())
				gasPriceAvg += gasPrice
				if gasPrice < gasPriceMin {
					gasPriceMin = gasPrice
				}
				if gasPrice > gasPriceMax {
					gasPriceMax = gasPrice
				}
				value += float64(tr.Value().Uint64())
			}
			gasPriceAvg = gasPriceAvg / float64(len(block.Body().Transactions))
			fSet("gasPriceAvg", fmt.Sprint(math.Round(gasPriceAvg/1000000000)), "gwei")
			fSet("gasPriceMin", fmt.Sprint(math.Round(gasPriceMin/1000000000)), "gwei")
			fSet("gasPriceMax", fmt.Sprint(math.Round(gasPriceMax/1000000000)), "gwei")
			fSet("totalValue", fmt.Sprint(math.Round(value/1000000000000000000)), "ETH")
		}

		client.Close()
	}

	for vName, _ := range c.receivedVariables {
		c.SetString(vName, "", "stopped")
	}

	c.SetString(ItemNameStatus, "", "stopped")
	c.Started = false
}
