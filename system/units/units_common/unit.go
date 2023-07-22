package units_common

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gazer_node/utilities/uom"
)

type Unit struct {
	mtx sync.Mutex

	unitId          string
	unitType        string
	unitDisplayName string
	config          string
	iUnit           common_interfaces.IUnit
	//iDataStorage    common_interfaces.IDataStorage
	lastError   string
	lastErrorDT time.Time
	lastInfo    string
	lastInfoDT  time.Time

	Properties map[string]common_interfaces.ItemProperty

	lastLogDT time.Time

	Started  bool
	Stopping bool

	watchItems map[string]bool

	output chan common_interfaces.UnitMessage
}

func (c *Unit) Init() {
	c.Properties = make(map[string]common_interfaces.ItemProperty)
	c.output = make(chan common_interfaces.UnitMessage)
}

func (c *Unit) Dispose() {
	close(c.output)
}

func (c *Unit) OutputChannel() chan common_interfaces.UnitMessage {
	return c.output
}

func (c *Unit) PropSetIfNotExists(name string, value string) {
	c.mtx.Lock()
	if _, ok := c.Properties[name]; !ok {
		c.Properties[name] = common_interfaces.ItemProperty{
			Name:  name,
			Value: value,
		}
	}
	c.mtx.Unlock()
}

func (c *Unit) Prop(name string) string {
	result := ""
	c.mtx.Lock()
	if prop, ok := c.Properties[name]; ok {
		result = prop.Value
	}
	c.mtx.Unlock()
	return result
}

func (c *Unit) PropSet(props []common_interfaces.ItemProperty) {
	c.mtx.Lock()
	for _, prop := range props {
		c.Properties[prop.Name] = prop
	}
	c.mtx.Unlock()
}

func (c *Unit) PropGet() []common_interfaces.ItemProperty {
	result := make([]common_interfaces.ItemProperty, 0)
	c.mtx.Lock()
	for _, prop := range c.Properties {
		result = append(result, prop)
	}
	c.mtx.Unlock()
	return result
}

func (c *Unit) Id() string {
	return c.unitId
}

func (c *Unit) SetId(id string) {
	c.unitId = id
}

func (c *Unit) SetIUnit(iUnit common_interfaces.IUnit) {
	c.iUnit = iUnit
}

func (c *Unit) SetMainItem(mainItem string) {
	c.PropSetIfNotExists("main_item", c.Id()+"/"+mainItem)
}

func (c *Unit) MainItem() string {
	return c.Prop("main_item")
}

func (c *Unit) Type() string {
	return c.unitType
}

func (c *Unit) SetType(unitType string) {
	c.unitType = unitType
}

func (c *Unit) DisplayName() string {
	return c.unitDisplayName
}

func (c *Unit) SetDisplayName(unitDisplayName string) {
	c.unitDisplayName = unitDisplayName
}

func (c *Unit) SetConfig(config string) {
	c.config = config
}

func (c *Unit) GetConfig() string {
	return c.config
}

func (c *Unit) GetConfigMeta() string {
	return ""
}

func (c *Unit) InternalInitItems() {

	c.SetStringForAll("", uom.STARTED)
}

func (c *Unit) InternalDeInitItems() {
	c.SetStringForAll("", uom.STOPPED)
}

func (c *Unit) Start() error {
	var err error
	c.watchItems = make(map[string]bool)
	if c.Started {
		return errors.New("already started")
	}
	c.LogInfo("")
	c.LogInfo("starting ...")
	c.SetStringService("name", c.DisplayName(), "")
	c.SetError("")
	c.SetStringService("status", "started", "")
	c.SetStringService("Unit", c.Type(), "")

	c.Stopping = false

	c.iUnit.InternalInitItems()
	err = c.iUnit.InternalUnitStart()

	if err != nil {
		c.SetError(err.Error())
		c.LogError(err.Error())
	} else {
		c.LogInfo("started")
	}

	return err
}

func (c *Unit) Stop() {
	logger.Println("Unit Stop", c.Id())
	if !c.Started {
		logger.Println("Unit Stop - unit is not started", c.Id())
		return
	}
	c.LogInfo("stopping ...")

	/*for itemWatched, _ := range c.watchItems {
		c.iDataStorage.RemoveFromWatch(c.Id(), itemWatched)
	}*/

	c.SetStringService("status", "stopping", "")
	c.Stopping = true
	logger.Println("Unit Stop - stopping - waiting", c.Id())
	for c.Started {
		time.Sleep(100 * time.Millisecond)
	}
	logger.Println("Unit Stop - stopping - waiting is ok", c.Id())
	logger.Println("Unit Stop - stopping", c.Id())
	logger.Println("Unit Stop - stopping - InternalDeInitItems", c.Id())
	c.iUnit.InternalDeInitItems()
	c.LogInfo("stopped")
	logger.Println("Unit Stop - complete", c.Id())
}

func (c *Unit) IsStarted() bool {
	return c.Started
}

const (
	UnitServicePrefix = ".service/"
	ItemNameError     = "error"
)

/*func (c *Unit) IDataStorage() common_interfaces.IDataStorage {
	return c.iDataStorage
}*/

func (c *Unit) SetStringService(name string, value string, UOM string) {
	fullName := c.Id() + "/" + UnitServicePrefix + name
	c.output <- &common_interfaces.UnitMessageItemValue{
		ItemName: fullName,
		Value:    value,
		UOM:      UOM,
	}
	//c.iDataStorage.SetItemByName(fullName, value, UOM, time.Now().UTC(), false)
}

func (c *Unit) LogInfo(value string) {
	dt := time.Now().UTC()
	if dt.Sub(c.lastLogDT) < 1*time.Microsecond {
		dt = dt.Add(1 * time.Microsecond)
	}
	c.lastLogDT = dt
	if c.lastInfo != value || time.Now().UTC().Sub(c.lastInfoDT) > 5*time.Second {
		fullName := c.Id() + "/" + UnitServicePrefix + "log"
		//c.iDataStorage.SetItemByName(fullName, value, "", dt, false)
		c.output <- &common_interfaces.UnitMessageItemValue{
			ItemName: fullName,
			Value:    value,
			UOM:      "",
		}
		c.lastInfoDT = time.Now().UTC()
	}
	c.lastInfo = value
	time.Sleep(1 * time.Microsecond)
}

func (c *Unit) LogError(value string) {
	dt := time.Now().UTC()
	if dt.Sub(c.lastLogDT) < 1*time.Microsecond {
		dt = dt.Add(1 * time.Microsecond)
	}
	c.lastLogDT = dt

	if c.lastError != value || time.Now().UTC().Sub(c.lastErrorDT) > 5*time.Second {
		fullName := c.Id() + "/" + UnitServicePrefix + "log"
		c.output <- &common_interfaces.UnitMessageItemValue{
			ItemName: fullName,
			Value:    value,
			UOM:      "error",
		}
		c.lastErrorDT = time.Now().UTC()
	}
	c.lastError = value
	time.Sleep(1 * time.Microsecond)
}

func (c *Unit) SetError(value string) {
	fullName := c.Id() + "/" + UnitServicePrefix + ItemNameError
	c.output <- &common_interfaces.UnitMessageItemValue{
		ItemName: fullName,
		Value:    value,
		UOM:      "",
	}
}

func (c *Unit) SetStringForAll(value string, UOM string) {
	//fullName := c.Id()
	//c.iDataStorage.SetAllItemsByUnitName(fullName, value, UOM, time.Now().UTC(), false)
	c.output <- &common_interfaces.UnitMessageSetAllItemsByUnitName{
		UnitId: c.Id(),
		Value:  value,
		UOM:    UOM,
	}
}

func (c *Unit) SetString(name string, value string, UOM string) {
	fullName := c.Id()
	if len(name) > 0 {
		fullName = c.Id() + "/" + name
	}
	c.output <- &common_interfaces.UnitMessageItemValue{
		ItemName: fullName,
		Value:    value,
		UOM:      UOM,
	}
}

func (c *Unit) SetPropertyIfDoesntExist(itemName string, propName string, propValue string) {
	c.output <- &common_interfaces.UnitMessageSetProperty{
		ItemName:  c.Id() + "/" + itemName,
		PropName:  propName,
		PropValue: propValue,
	}
}

func (c *Unit) TouchItem(name string) {
	fullName := c.Id()
	if len(name) > 0 {
		fullName = c.Id() + "/" + name
	}
	//c.iDataStorage.TouchItem(fullName)
	c.output <- &common_interfaces.UnitMessageItemTouch{
		ItemName: fullName,
	}
}

func (c *Unit) SetInt(name string, value int, UOM string) {
	c.SetString(name, strconv.Itoa(value), UOM)
}

func (c *Unit) SetInt64(name string, value int64, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetUInt64(name string, value uint64, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetInt32(name string, value int32, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetUInt32(name string, value uint32, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetInt8(name string, value int8, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetUInt8(name string, value uint8, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetInt16(name string, value int16, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetUInt16(name string, value uint16, UOM string) {
	c.SetString(name, fmt.Sprint(value), UOM)
}

func (c *Unit) SetFloat32(name string, value float32, UOM string, precision int) {
	c.SetString(name, strconv.FormatFloat(float64(value), 'f', precision, 64), UOM)
}

func (c *Unit) SetFloat64(name string, value float64, UOM string, precision int) {
	c.SetString(name, strconv.FormatFloat(value, 'f', precision, 64), UOM)
}

/*func (c *Unit) GetValue(name string) string {
	item, err := c.iDataStorage.GetItem(name)
	if err != nil {
		return ""
	}
	return item.Value.Value
}

func (c *Unit) GetItem(name string) (common_interfaces.ItemValue, error) {
	item, err := c.iDataStorage.GetItem(name)
	if err != nil {
		return common_interfaces.ItemValue{}, err
	}
	return item.Value, nil
}*/

/*func (c *Unit) GetItemsOfUnit(unitId string) ([]common_interfaces.ItemGetUnitItems, error) {
	return c.iDataStorage.GetUnitValues(unitId), nil
}

func (c *Unit) AddToWatch(itemName string) {
	c.iDataStorage.AddToWatch(c.Id(), itemName)
	c.watchItems[itemName] = true
}

func (c *Unit) RemoveFromWatch(itemName string) {
	c.iDataStorage.RemoveFromWatch(c.Id(), itemName)
	delete(c.watchItems, itemName)
}*/

func (c *Unit) ItemChanged(itemName string, value common_interfaces.ItemValue) {
}
