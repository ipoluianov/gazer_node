package common_interfaces

import (
	"time"
)

type IDataStorage interface {
	SetItemByNameOld(name string, value string, UOM string, dt time.Time, external bool) error
	SetAllItemsByUnitName(name string, value string, UOM string, dt time.Time, external bool) error
	TouchItem(name string) (*Item, error)
	GetItem(name string) (Item, error)
	GetUnitValues(unitId string) []ItemGetUnitItems
	RenameItems(oldPrefix string, newPrefix string)
	RemoveItemsOfUnit(unitId string) error
	SetProperty(itemName string, propName string, propValue string)
	SetPropertyIfDoesntExist(itemName string, propName string, propValue string)
}
