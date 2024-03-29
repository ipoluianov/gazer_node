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
	config   Config
}

type ConfigItem struct {
	Name    string `json:"name"`
	Index   string `json:"index"`
	Mode    string `json:"mode"`
	Default string `json:"default"`
}

type Config struct {
	Period float64       `json:"period"`
	Pins   []*ConfigItem `json:"pins"`
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

	t1 := meta.Add("pins", "Pins", "", "table", "", "", "")
	t1.Add("name", "Name", "pin_name", "string", "", "", "")
	t1.Add("index", "GPIO#", "0", "string", "", "", "raspberry-pi-gpio")
	t1.Add("mode", "Mode", "input", "string", "", "", "gpio-mode")
	t1.Add("default", "Default", "0", "string", "", "", "")
	return meta.Marshal()
}

func (c *UnitRaspberryPiGPIO) InternalUnitStart() error {
	var err error
	c.SetString(ItemNameResult, "", "")
	c.SetMainItem(ItemNameResult)

	err = json.Unmarshal([]byte(c.GetConfig()), &c.config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameResult, err.Error(), "error")
		return err
	}

	c.periodMs = int(c.config.Period)
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
	//var err error
	c.Started = true
	dtOperationTime := time.Now().UTC()

	for _, item := range c.config.Pins {
		c.SetString(item.Name, item.Mode, item.Index)
	}

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

		/*st, err := c.IDataStorage().GetItem(c.Id() + "/name")
		if err != nil {
			if st.Value.Value == "1" {
				// High
			} else {
				// Low
			}
		}

		if err != nil {
			c.SetString(ItemNameResult, err.Error(), "error")
			continue
		}*/
	}
	c.SetString(ItemNameResult, "", "stopped")
	c.Started = false
}
