package raspberrypi_cpu_temperature_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitRaspberryPiCPUTemp struct {
	units_common.Unit
	periodMs int
}

func New() common_interfaces.IUnit {
	var c UnitRaspberryPiCPUTemp
	return &c
}

const (
	ItemNameResult = "Temperature"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "RaspberryPi.CPU.Temperature.Alfa"
	info.Category = "raspberry_pi"
	info.DisplayName = "Raspberry PI CPU temperature"
	info.Constructor = New
	info.ImgBytes = nil
	info.Description = ""
	return info
}

func (c *UnitRaspberryPiCPUTemp) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "")
	return meta.Marshal()
}

func (c *UnitRaspberryPiCPUTemp) InternalUnitStart() error {
	var err error
	c.SetString(ItemNameResult, "", "")
	c.SetMainItem(ItemNameResult)

	type Config struct {
		Period float64 `json:"period"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameResult, err.Error(), "error")
		return err
	}

	c.periodMs = int(config.Period)
	if c.periodMs < 100 {
		err = errors.New("wrong period")
		c.SetString(ItemNameResult, err.Error(), "error")
		return err
	}

	go c.Tick()
	return nil
}

func (c *UnitRaspberryPiCPUTemp) InternalUnitStop() {
}

func (c *UnitRaspberryPiCPUTemp) Tick() {
	c.Started = true
	dtOperationTime := time.Now().UTC()

	for !c.Stopping {
		for {
			if c.Stopping || time.Now().Sub(dtOperationTime) > time.Duration(c.periodMs)*time.Millisecond {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			break
		}
		dtOperationTime = time.Now().UTC()

		bs, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")

		if err == nil {
			valueAsString := strings.TrimSpace(string(bs))
			valueAsFloat, err := strconv.ParseFloat(valueAsString, 64)
			if err == nil {
				c.SetFloat64(ItemNameResult, valueAsFloat/1000.0, "°C", 1)
			}
		}

		if err != nil {
			c.SetString(ItemNameResult, err.Error(), "error")
			continue
		}
	}
	c.SetString(ItemNameResult, "", "stopped")
	c.Started = false
}
