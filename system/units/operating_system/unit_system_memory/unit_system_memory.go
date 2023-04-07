package unit_system_memory

import (
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/resources"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gazer_node/utilities/uom"
	"github.com/shirou/gopsutil/mem"
)

type UnitSystemMemory struct {
	units_common.Unit

	totalIsSet bool
}

var Image []byte

func init() {
	Image = resources.R_files_sensors_unit_computer_memory_png
}

func New() common_interfaces.IUnit {
	var c UnitSystemMemory
	return &c
}

func (c *UnitSystemMemory) InternalUnitStart() error {
	c.SetMainItem("UsedPercent")

	c.SetString("Total", "", "")
	c.SetString("Available", "", "")
	c.SetString("Used", "", "")

	c.SetString("UsedPercent", "", "")

	go c.Tick()
	return nil
}

func (c *UnitSystemMemory) InternalUnitStop() {
}

func (c *UnitSystemMemory) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	return meta.Marshal()
}

func (c *UnitSystemMemory) Tick() {
	defer func() {
		if r := recover(); r != nil {
			logger.Println("Panic in Unit")
			c.Started = false
			c.SetStringForAll("panic", "error")
		}
	}()
	dtBegin := time.Now()
	c.Started = true
	for !c.Stopping {
		for i := 0; i < 10; i++ {
			if c.Stopping {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		v, _ := mem.VirtualMemory()

		percents := (float64(v.Used) / float64(v.Total)) * 100.0

		// Common
		if !c.totalIsSet {
			c.SetUInt64("Total", v.Total/1048576, uom.MB)
			c.totalIsSet = true
		}
		c.SetUInt64("Available", v.Available/1048576, uom.MB)
		c.SetUInt64("Used", v.Used/1048576, uom.MB)
		c.SetFloat64("UsedPercent", percents, "%", 1)

		if time.Now().Sub(dtBegin).Seconds() > 5 {
			//panic("This is panic!")
		}
	}

	time.Sleep(1 * time.Millisecond)
	c.SetString("Total", "", "stopped")
	c.SetString("Available", "", "stopped")
	c.SetString("Used", "", "stopped")
	c.SetString("UsedPercent", "", "stopped")

	c.Started = false
}
