package unit_processes

import (
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitSystemProcesses struct {
	units_common.Unit
	periodMs int
}

func New() common_interfaces.IUnit {
	var c UnitSystemProcesses
	return &c
}

func (c *UnitSystemProcesses) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "0")
	return meta.Marshal()
}

type ProcessInfo struct {
	Name string
	Id   int
	Info string
}
