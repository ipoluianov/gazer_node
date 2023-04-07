package unit_network

import (
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/resources"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitNetwork struct {
	units_common.Unit
	addressesOfInterfaces map[int]string
}

var Image []byte

func init() {
	Image = resources.R_files_sensors_unit_computer_network_png
}

func New() common_interfaces.IUnit {
	var c UnitNetwork
	return &c
}

func (c *UnitNetwork) InternalUnitStart() error {
	c.SetString("TotalSpeed", "", "")
	c.SetMainItem("TotalSpeed")

	go c.Tick()
	return nil
}

func (c *UnitNetwork) InternalUnitStop() {
}

func (c *UnitNetwork) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	return meta.Marshal()
}

func (c *UnitNetwork) Tick() {
}
