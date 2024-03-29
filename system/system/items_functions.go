package system

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/protocols/nodeinterface"
	"github.com/ipoluianov/gazer_node/system/history"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

func (c *System) SetItemByNameOld(name string, value string, UOM string, dt time.Time, external bool) error {
	var item *common_interfaces.Item
	if name == "" {
		return nil
	}

	c.mtxSystem.Lock()
	if i, ok := c.itemsByName[name]; ok {
		item = i
	} else {
		item = common_interfaces.NewItem()
		item.Id = c.nextItemId
		item.Name = name
		c.itemsByName[item.Name] = item
		c.itemsById[item.Id] = item
		c.items = append(c.items, item)
		c.nextItemId++
	}
	c.mtxSystem.Unlock()

	var itemValue common_interfaces.ItemValue
	itemValue.Value = value
	itemValue.DT = dt.UnixNano() / 1000
	itemValue.UOM = UOM
	err := c.SetItem(item.Id, itemValue, 0, external)
	if err != nil {
		return err
	}

	return nil
}

func (c *System) SetAllItemsByUnitName(unitName string, value string, UOM string, dt time.Time, external bool) error {
	items := make([]*common_interfaces.Item, 0)
	if unitName == "" {
		return nil
	}

	c.mtxSystem.Lock()
	for _, i := range c.items {
		if strings.HasPrefix(i.Name, unitName+"/") {
			items = append(items, i)
		}
	}
	c.mtxSystem.Unlock()

	for _, i := range items {
		var itemValue common_interfaces.ItemValue
		itemValue.Value = value
		itemValue.DT = dt.UnixNano() / 1000
		itemValue.UOM = UOM
		c.SetItem(i.Id, itemValue, 0, external)
	}

	return nil
}

func (c *System) SetItem(itemId uint64, value common_interfaces.ItemValue, counter int, external bool) error {
	var item *common_interfaces.Item

	needToWriteHistory := true

	counter++
	if counter > 10 {
		return errors.New("recursion detected")
	}
	c.mtxSystem.Lock()
	if i, ok := c.itemsById[itemId]; ok {
		item = i
		value.Value = item.PostprocessingValue(value.Value)
		item.Value = value
		if prop, ok := item.Properties["history_disabled"]; ok {
			if prop.Value == "true" {
				needToWriteHistory = false
			}
		}
	}
	c.mtxSystem.Unlock()
	if item == nil {
		logger.Println("set item error: ", itemId, "=", value.Value)
		return errors.New("item not found")
	}

	if needToWriteHistory {
		c.history.Write(item.Id, value)
	}

	if external {
		c.unitsSystem.ItemChanged(item.Id, item.Name, value)
	}

	return nil
}

func (c *System) DataItemPropSet(itemName string, props []nodeinterface.PropItem) error {
	c.mtxSystem.Lock()
	if item, ok := c.itemsByName[itemName]; ok {
		for _, prop := range props {
			item.Properties[prop.PropName] = &common_interfaces.ItemProperty{
				Name:  prop.PropName,
				Value: prop.PropValue,
			}
		}
		c.applyItemsProperties()
	} else {
		c.mtxSystem.Unlock()
		return errors.New("item not found")
	}
	c.mtxSystem.Unlock()
	c.SaveConfig()
	return nil
}

func (c *System) applyItemsProperties() {
	// Need to be synced
	for _, item := range c.items {
		for _, prop := range item.Properties {
			if prop.Name == "tune_trim" {
				item.SetPostprocessingTrim(prop.Value == "1")
			}
			if prop.Name == "tune_on" {
				item.SetPostprocessingAdjust(prop.Value == "1")
			}
			if prop.Name == "tune_scale" {
				v, _ := strconv.ParseFloat(prop.Value, 64)
				item.SetPostprocessingScale(v)
			}
			if prop.Name == "tune_offset" {
				v, _ := strconv.ParseFloat(prop.Value, 64)
				item.SetPostprocessingOffset(v)
			}
			if prop.Name == "tune_precision" {
				precision, _ := strconv.ParseInt(prop.Value, 10, 64)
				item.SetPostprocessingPrecision(int(precision))
			}
		}
	}
}

func (c *System) DataItemPropGet(itemName string) ([]nodeinterface.PropItem, error) {
	result := make([]nodeinterface.PropItem, 0)

	c.mtxSystem.Lock()
	if item, ok := c.itemsByName[itemName]; ok {
		for _, prop := range item.Properties {
			result = append(result, nodeinterface.PropItem{
				PropName:  prop.Name,
				PropValue: prop.Value,
			})

			if prop.Name == "source" {
				if prop.Value != "" {
					sourceItemId, errParseSourceItemId := strconv.ParseUint(prop.Value, 10, 64)
					if errParseSourceItemId == nil {
						if itemSource, ok := c.itemsById[sourceItemId]; ok {
							result = append(result, nodeinterface.PropItem{
								PropName:  "#source_item_name",
								PropValue: itemSource.Name,
							})
						}
					}
				}
			}
		}
	} else {
		c.mtxSystem.Unlock()
		return nil, errors.New("item not found")
	}
	c.mtxSystem.Unlock()
	return result, nil
}

type ItemWatcher struct {
	UnitIDs map[string]bool
}

func (c *System) SetProperty(itemName string, propName string, propValue string) {
	item, err := c.TouchItem(itemName)
	if err == nil {
		c.mtxSystem.Lock()
		item.SetProperty(propName, propValue)
		c.mtxSystem.Unlock()
	}
}

func (c *System) SetPropertyIfDoesntExist(itemName string, propName string, propValue string) {
	item, err := c.TouchItem(itemName)
	if err == nil {
		c.mtxSystem.Lock()
		item.SetPropertyIfDoesntExist(propName, propValue)
		c.mtxSystem.Unlock()
	}
}

func (c *System) TouchItem(name string) (*common_interfaces.Item, error) {
	var item *common_interfaces.Item
	fullName := name
	c.mtxSystem.Lock()
	var ok bool
	if item, ok = c.itemsByName[fullName]; !ok {
		item = common_interfaces.NewItem()
		item.Id = c.nextItemId
		item.Name = fullName
		c.itemsByName[item.Name] = item
		c.itemsById[item.Id] = item
		c.items = append(c.items, item)
		c.nextItemId++
	}
	c.mtxSystem.Unlock()
	return item, nil
}

func (c *System) GetItem(name string) (common_interfaces.Item, error) {
	var item common_interfaces.Item
	var found bool
	c.mtxSystem.Lock()
	if i, ok := c.itemsByName[name]; ok {
		item = *i
		found = true
	}
	c.mtxSystem.Unlock()

	if !found {
		return item, errors.New("no item found")
	}

	return item, nil
}

func (c *System) RemoveItems(itemsNames []string) error {
	var err error

	c.mtxSystem.Lock()
	newItems := make([]*common_interfaces.Item, 0)
	itemsForRemove := make([]*common_interfaces.Item, 0)

	itemsNamesMap := make(map[string]bool)
	for _, itemName := range itemsNames {
		itemsNamesMap[itemName] = true
	}

	for _, item := range c.items {
		if _, ok := itemsNamesMap[item.Name]; ok {
			itemsForRemove = append(itemsForRemove, item)
		} else {
			newItems = append(newItems, item)
		}
	}
	c.items = newItems

	for _, item := range itemsForRemove {
		delete(c.itemsByName, item.Name)
		delete(c.itemsById, item.Id)
		c.history.RemoveItem(item.Id)
	}
	c.mtxSystem.Unlock()

	//c.publicChannels.RemoveItems(nil, itemsNames)

	err = c.SaveConfig()

	if len(itemsForRemove) == 0 {
		return errors.New("no items found")
	}

	return err
}

func (c *System) GetItems() []common_interfaces.Item {
	var items []common_interfaces.Item
	c.mtxSystem.Lock()
	items = make([]common_interfaces.Item, len(c.items))
	for index, item := range c.items {
		items[index] = *item
	}
	c.mtxSystem.Unlock()
	return items
}

func (c *System) RenameItems(oldPrefix string, newPrefix string) {
	c.mtxSystem.Lock()
	for _, item := range c.items {
		if strings.HasPrefix(item.Name, oldPrefix) {
			delete(c.itemsByName, item.Name)
			item.Name = strings.Replace(item.Name, oldPrefix, newPrefix, 1)
			c.itemsByName[item.Name] = item
		}
	}
	c.mtxSystem.Unlock()

	//c.publicChannels.RenameItems(oldPrefix, newPrefix)
}

func (c *System) ReadHistory(name string, dtBegin int64, dtEnd int64) (*history.ReadResult, error) {
	c.mtxSystem.Lock()
	item, ok := c.itemsByName[name]
	c.mtxSystem.Unlock()
	if ok {
		return c.history.Read(item.Id, dtBegin, dtEnd), nil
	}

	var result history.ReadResult
	return &result, errors.New("no item found")
}

func (c *System) GetStatistics() (common_interfaces.Statistics, error) {
	var res common_interfaces.Statistics
	//res.CloudSentBytes = c.publicChannels.SentBytes()
	//res.CloudReceivedBytes = c.publicChannels.ReceivedBytes()
	res.ApiCalls = c.apiCallsCount
	return res, nil
}

/*func (c *System) GetApi() (nodeinterface.ServiceApiResponse, error) {
	var res nodeinterface.ServiceApiResponse
	res.Product = productinfo.Name()
	res.Version = productinfo.Version()
	res.BuildTime = productinfo.BuildTime()
	res.SupportedFunctions = nodeinterface.ApiFunctions()

	return res, nil
}*/

func (c *System) SetNodeName(name string) error {
	c.nodeName = name
	return c.SaveConfig()
}

func (c *System) NodeName() string {
	return c.nodeName
}

func (c *System) GetInfo() (nodeinterface.ServiceInfoResponse, error) {
	var res nodeinterface.ServiceInfoResponse
	res.NodeName = c.NodeName()
	res.Version = VERSION
	res.BuildTime = BUILDTIME
	res.GuestKey = c.currentGuestKey
	res.Time = strconv.FormatInt(time.Now().UnixNano(), 10)
	return res, nil
}
