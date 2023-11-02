package iunit

import (
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/protocols/nodeinterface"
)

type UnitMessage interface{}

type UnitMessageItemValue struct {
	ItemName string
	Value    string
	UOM      string
}

type UnitMessageItemTouch struct {
	ItemName string
}

type UnitMessageSetProperty struct {
	ItemName  string
	PropName  string
	PropValue string
}

type UnitMessageRemoteItemsOfUnit struct {
	UnitId string
}

type UnitMessageSetAllItemsByUnitName struct {
	UnitId string
	Value  string
	UOM    string
}

type IUnit interface {
	Init()
	Dispose()
	Id() string
	OutputChannel() chan UnitMessage
	SetId(unitId string)
	Type() string
	SetType(unitType string)
	DisplayName() string
	SetDisplayName(unitDisplayName string)
	SetIUnit(iUnit IUnit)
	MainItem() string
	Start() error
	Stop()
	IsStarted() bool
	SetConfig(config string)
	GetConfig() string
	GetConfigMeta() string

	InternalUnitStart() error
	InternalUnitStop()
	ItemChanged(itemId uint64, itemName string, value common_interfaces.ItemValue)

	InternalInitItems()
	InternalDeInitItems()

	PropSet(props []common_interfaces.ItemProperty)
	PropGet() []common_interfaces.ItemProperty
	Prop(name string) string
	PropSetIfNotExists(name string, value string)

	SetNode(inode nodeinterface.INode)
}
