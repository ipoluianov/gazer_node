package unit_storage

import (
	_ "embed"
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/resources"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitStorage struct {
	units_common.Unit
}

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Computer.Storage.Watcher.Alfa"
	info.Category = "computer"
	info.DisplayName = "Storage"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
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
