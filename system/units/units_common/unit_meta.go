package units_common

import "github.com/ipoluianov/gazer_node/common_interfaces"

type UnitMeta struct {
	TypeName    string
	Category    string
	DisplayName string
	Constructor func() common_interfaces.IUnit
	ImgBytes    []byte
	Description string
}
