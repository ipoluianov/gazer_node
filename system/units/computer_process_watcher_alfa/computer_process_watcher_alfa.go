package computer_process_watcher_alfa

import (
	_ "embed"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitSystemProcess struct {
	units_common.Unit

	processIdActive   bool
	processId         uint32
	processNameActive bool
	processName       string
	periodMs          int

	actualProcessName string
}

//go:embed "image.png"
var Image []byte

func New() common_interfaces.IUnit {
	var c UnitSystemProcess
	return &c
}

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Computer.Process.Watcher.Alfa"
	info.Category = "computer"
	info.DisplayName = "Process"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitSystemProcess) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("process_name", "Process Name", "notepad.exe", "string", "", "", "processes")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "0")
	return meta.Marshal()
}

type ProcessInfo struct {
	Name string
	Id   int
	Info string
}
