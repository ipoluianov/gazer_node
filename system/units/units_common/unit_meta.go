package units_common

import "github.com/ipoluianov/gazer_node/iunit"

type UnitMeta struct {
	TypeName    string
	Category    string
	DisplayName string
	Constructor func() iunit.IUnit
	ImgBytes    []byte
	Description string
}
