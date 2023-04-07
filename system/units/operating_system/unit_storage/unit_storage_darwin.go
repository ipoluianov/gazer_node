package unit_storage

import (
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/resources"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitStorage struct {
	units_common.Unit
}

var Image []byte

func init() {
	Image = resources.R_files_sensors_unit_computer_storage_png
}

func New() common_interfaces.IUnit {
	var c UnitStorage
	return &c
}

func (c *UnitStorage) InternalUnitStart() error {
	c.SetString("UsedPercents", "", "")
	c.SetMainItem("UsedPercents")

	go c.Tick()
	return nil
}

func (c *UnitStorage) InternalUnitStop() {
}

func (c *UnitStorage) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	return meta.Marshal()
}

func (c *UnitStorage) Tick() {
}
