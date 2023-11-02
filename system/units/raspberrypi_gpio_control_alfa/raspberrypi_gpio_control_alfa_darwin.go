package raspberrypi_gpio_control_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitRaspberryPiGPIO struct {
	units_common.Unit
	periodMs int
}

func New() iunit.IUnit {
	var c UnitRaspberryPiGPIO
	return &c
}

const (
	ItemNameResult = "Result"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "RaspberryPi.GPIO.Control.Alfa"
	info.Category = "raspberry_pi"
	info.DisplayName = "Raspberry PI GPIO"
	info.Constructor = New
	info.ImgBytes = nil
	info.Description = ""
	return info
}

func (c *UnitRaspberryPiGPIO) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "")
	return meta.Marshal()
}

func (c *UnitRaspberryPiGPIO) InternalUnitStart() error {
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

func (c *UnitRaspberryPiGPIO) InternalUnitStop() {
}

func (c *UnitRaspberryPiGPIO) Tick() {
	var err error
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

		c.SetInt(ItemNameResult, 0, "")

		if err != nil {
			c.SetString(ItemNameResult, err.Error(), "error")
			continue
		}
	}
	c.SetString(ItemNameResult, "", "stopped")
	c.Started = false
}
