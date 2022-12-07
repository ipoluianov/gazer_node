package common_interfaces

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
	ItemChanged(itemName string, value ItemValue)

	InternalInitItems()
	InternalDeInitItems()

	PropSet(props []ItemProperty)
	PropGet() []ItemProperty
	Prop(name string) string
	PropSetIfNotExists(name string, value string)
}
