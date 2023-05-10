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
	"github.com/ipoluianov/gazer_node/utilities/uom"
)

type UnitEthereumRealTimeStat struct {
	units_common.Unit
	rpcUrl            string
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
	meta.Add("rpcUrl", "RPC URL", "", "string", "", "", "")
	meta.Add("period", "Period, ms", "5000", "num", "0", "3600000", "0")
	return meta.Marshal()
}

func (c *UnitEthereumRealTimeStat) InternalUnitStart() error {
	var err error

	type Config struct {
		RpcUrl string  `json:"rpcUrl"`
		Period float64 `json:"period"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}

	c.rpcUrl = config.RpcUrl
	if c.rpcUrl == "" {
		err = errors.New("wrong rpc url")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}

	c.periodMs = int(config.Period)
	if c.periodMs < 100 {
		err = errors.New("wrong period")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}

	if c.periodMs < 100 {
		err = errors.New("wrong period (<100)")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}
	if c.periodMs > 3600000 {
		err = errors.New("wrong period (>3600000)")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}

	c.receivedVariables = make(map[string]string)

	c.SetMainItem(ItemNameStatus)

	c.SetString(ItemNameStatus, "", uom.NONE)
	go c.Tick()
	return nil
}

func (c *UnitEthereumRealTimeStat) InternalUnitStop() {
}

func (c *UnitEthereumRealTimeStat) Tick() {
	// var err error
	c.Started = true
	dtLastTime := time.Now().UTC().Add(-time.Duration(c.periodMs) * time.Millisecond)

	lastBlock := uint64(0)

	for !c.Stopping {
		for {
			if c.Stopping || time.Now().Sub(dtLastTime) > time.Duration(c.periodMs)*time.Millisecond {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			c.SetString(ItemNameStatus, "", uom.STOPPED)
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
		block, err := client.BlockByNumber(context.Background(), nil)
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

		if block != nil && block.Header() != nil {
			if block.Header().Number.Uint64() != lastBlock {
				lastBlock = block.Header().Number.Uint64()
				fSet("blockNumber", fmt.Sprint(block.Header().Number.Uint64()), "")
				fSet("transactionCount", fmt.Sprint(block.Transactions().Len()), "")
				if len(block.Body().Transactions) > 0 {
					gasPriceAvg := float64(0)
					gasPriceMin := float64(10000000000000000000)
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
			}
		}

		c.SetString(ItemNameStatus, "ok", uom.NONE)

		client.Close()
	}

	for vName, _ := range c.receivedVariables {
		c.SetString(vName, uom.NONE, uom.STOPPED)
	}

	c.SetString(ItemNameStatus, uom.NONE, uom.STOPPED)
	c.Started = false
}
