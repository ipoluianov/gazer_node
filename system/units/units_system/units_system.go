package units_system

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/resources"
	"github.com/ipoluianov/gazer_node/system/protocols/nodeinterface"
	"github.com/ipoluianov/gazer_node/system/units/blockchain_ethereum_balance_alfa"
	"github.com/ipoluianov/gazer_node/system/units/blockchain_ethereum_lastblock_alfa"
	"github.com/ipoluianov/gazer_node/system/units/computer_network_adapters_alfa"
	"github.com/ipoluianov/gazer_node/system/units/computer_process_watcher_alfa"
	"github.com/ipoluianov/gazer_node/system/units/computer_storage_watcher_alfa"
	"github.com/ipoluianov/gazer_node/system/units/computer_system_memory_alfa"
	"github.com/ipoluianov/gazer_node/system/units/database_postgresql_query_alfa"
	"github.com/ipoluianov/gazer_node/system/units/files_file_content_alfa"
	"github.com/ipoluianov/gazer_node/system/units/files_file_size_alfa"
	"github.com/ipoluianov/gazer_node/system/units/files_file_tail_alfa"
	"github.com/ipoluianov/gazer_node/system/units/files_tabtable_directory_alfa"
	"github.com/ipoluianov/gazer_node/system/units/files_tabtable_singlefile_alfa"
	"github.com/ipoluianov/gazer_node/system/units/general_console_keyvalue_alfa"
	"github.com/ipoluianov/gazer_node/system/units/general_console_singlevalue_alfa"
	"github.com/ipoluianov/gazer_node/system/units/general_hhgttg_42_alfa"
	"github.com/ipoluianov/gazer_node/system/units/general_manual_items_alfa"
	"github.com/ipoluianov/gazer_node/system/units/general_signal_generator_alfa"
	"github.com/ipoluianov/gazer_node/system/units/network_http_rest_alfa"
	"github.com/ipoluianov/gazer_node/system/units/network_ping_regular_alfa"
	"github.com/ipoluianov/gazer_node/system/units/network_ssl_expires_alfa"
	"github.com/ipoluianov/gazer_node/system/units/network_tcp_connect_alfa"
	"github.com/ipoluianov/gazer_node/system/units/network_udp_fields_alfa"
	"github.com/ipoluianov/gazer_node/system/units/raspberrypi_cpu_temperature_alfa"
	"github.com/ipoluianov/gazer_node/system/units/raspberrypi_gpio_control_alfa"
	"github.com/ipoluianov/gazer_node/system/units/serialport_keyvalue_watcher_alfa"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

type UnitsSystem struct {
	units        []common_interfaces.IUnit
	unitTypes    []*UnitType
	unitTypesMap map[string]*UnitType
	//iDataStorage common_interfaces.IDataStorage
	mtx sync.Mutex

	output chan common_interfaces.UnitMessage
}

var unitCategoriesIcons map[string][]byte
var unitCategoriesNames map[string]string

func init() {
	unitCategoriesIcons = make(map[string][]byte)
	unitCategoriesNames = make(map[string]string)
}

func New(iDataStorage common_interfaces.IDataStorage) *UnitsSystem {
	var c UnitsSystem
	c.unitTypes = make([]*UnitType, 0)
	c.unitTypesMap = make(map[string]*UnitType)
	c.output = make(chan common_interfaces.UnitMessage)

	c.RegUnitType(network_ping_regular_alfa.Info())
	c.RegUnitType(network_tcp_connect_alfa.Info())
	c.RegUnitType(network_http_rest_alfa.Info())
	c.RegUnitType(network_ssl_expires_alfa.Info())
	c.RegUnitType(network_udp_fields_alfa.Info())

	c.RegUnitType(computer_system_memory_alfa.Info())
	c.RegUnitType(computer_process_watcher_alfa.Info())
	c.RegUnitType(computer_storage_watcher_alfa.Info())
	c.RegUnitType(computer_network_adapters_alfa.Info())

	c.RegUnitType(files_file_size_alfa.Info())
	c.RegUnitType(files_file_content_alfa.Info())
	c.RegUnitType(files_file_tail_alfa.Info())
	c.RegUnitType(files_tabtable_singlefile_alfa.Info())
	c.RegUnitType(files_tabtable_directory_alfa.Info())

	c.RegUnitType(general_console_singlevalue_alfa.Info())
	c.RegUnitType(general_console_keyvalue_alfa.Info())
	c.RegUnitType(general_manual_items_alfa.Info())
	c.RegUnitType(general_hhgttg_42_alfa.Info())
	c.RegUnitType(general_signal_generator_alfa.Info())

	c.RegUnitType(serialport_keyvalue_watcher_alfa.Info())

	if runtime.GOOS == "linux" {
		c.RegUnitType(raspberrypi_gpio_control_alfa.Info())
		c.RegUnitType(raspberrypi_cpu_temperature_alfa.Info())
	}

	c.RegUnitType(database_postgresql_query_alfa.Info())

	c.RegUnitType(blockchain_ethereum_lastblock_alfa.Info())
	c.RegUnitType(blockchain_ethereum_balance_alfa.Info())

	c.initCategories()

	return &c
}

func (c *UnitsSystem) OutputChannel() chan common_interfaces.UnitMessage {
	return c.output
}

func CategoryOfUnit(unitType string) string {
	parts := strings.FieldsFunc(unitType, func(r rune) bool {
		return r == '.'
	})
	if len(parts) > 0 {
		return parts[0]
	}
	return "Unknown"
}

func NameOfUnit(unitType string) string {
	parts := strings.FieldsFunc(unitType, func(r rune) bool {
		return r == '.'
	})
	if len(parts) > 2 {
		result := ""
		for i := 1; i < len(parts)-1; i++ {
			if len(result) > 0 {
				result += " "
			}
			result += parts[i]
		}
		return result
	}
	return "Unknown"
}

func (c *UnitsSystem) initCategories() {
	unitCategoriesIcons = make(map[string][]byte)
	unitCategoriesNames = make(map[string]string)
	for _, value := range c.unitTypesMap {
		unitCategoriesNames[CategoryOfUnit(value.TypeCode)] = CategoryOfUnit(value.TypeCode)
		unitCategoriesIcons[CategoryOfUnit(value.TypeCode)] = resources.R_files_sensors_category_general_png
	}
}

func (c *UnitsSystem) RegUnitType(info units_common.UnitMeta) *UnitType {
	var sType UnitType
	sType.TypeCode = info.TypeName
	sType.Category = info.Category
	sType.DisplayName = info.DisplayName
	if len(sType.DisplayName) == 0 {
		sType.DisplayName = NameOfUnit(info.TypeName)
	}
	sType.Constructor = info.Constructor
	sType.Picture = info.ImgBytes

	if sType.Picture == nil {
		sType.Picture = computer_system_memory_alfa.Image
	}

	sType.Description = info.Description
	c.unitTypes = append(c.unitTypes, &sType)
	c.unitTypesMap[info.TypeName] = &sType
	return &sType
}

func (c *UnitsSystem) RegisterUnit(typeName string, category string, displayName string, constructor func() common_interfaces.IUnit, imgBytes []byte, description string) *UnitType {
	var sType UnitType
	sType.TypeCode = typeName
	sType.Category = category
	sType.DisplayName = displayName
	sType.Constructor = constructor
	sType.Picture = imgBytes
	sType.Description = description
	c.unitTypes = append(c.unitTypes, &sType)
	c.unitTypesMap[typeName] = &sType
	return &sType
}

func (c *UnitsSystem) UnitTypes() []nodeinterface.UnitTypeListResponseItem {
	result := make([]nodeinterface.UnitTypeListResponseItem, 0)
	for _, st := range c.unitTypes {
		var unitTypeInfo nodeinterface.UnitTypeListResponseItem
		unitTypeInfo.Type = st.TypeCode
		unitTypeInfo.Category = st.Category
		unitTypeInfo.DisplayName = st.DisplayName
		unitTypeInfo.Help = st.Help
		unitTypeInfo.Description = st.Description
		unitTypeInfo.Image = st.Picture
		result = append(result, unitTypeInfo)
	}
	return result
}

func (c *UnitsSystem) UnitCategories() nodeinterface.UnitTypeCategoriesResponse {
	var result nodeinterface.UnitTypeCategoriesResponse
	result.Items = make([]nodeinterface.UnitTypeCategoriesResponseItem, 0)
	addedCats := make(map[string]bool)

	catAllName := ""
	var unitCategoryInfoAll nodeinterface.UnitTypeCategoriesResponseItem
	unitCategoryInfoAll.Name = catAllName
	unitCategoryInfoAll.DisplayName = "All"
	if imgBytes, ok := unitCategoriesIcons[catAllName]; ok {
		unitCategoryInfoAll.Image = imgBytes
	} else {
		unitCategoryInfoAll.Image = resources.R_files_sensors_category_general_png
	}
	result.Items = append(result.Items, unitCategoryInfoAll)
	addedCats[catAllName] = true

	for _, st := range c.unitTypes {
		if _, ok := addedCats[st.Category]; !ok {
			var unitCategoryInfo nodeinterface.UnitTypeCategoriesResponseItem
			unitCategoryInfo.Name = st.Category
			if catName, ok := unitCategoriesNames[st.Category]; ok {
				unitCategoryInfo.DisplayName = catName
			} else {
				unitCategoryInfo.DisplayName = st.Category
			}
			if imgBytes, ok := unitCategoriesIcons[st.Category]; ok {
				unitCategoryInfo.Image = imgBytes
			} else {
				unitCategoryInfo.Image = st.Picture
			}
			result.Items = append(result.Items, unitCategoryInfo)
			addedCats[st.Category] = true
		}
	}
	return result
}

func (c *UnitsSystem) UnitTypeForDisplayByType(t string) string {
	if res, ok := c.unitTypesMap[t]; ok {
		return res.DisplayName
	}
	return t
}

func (c *UnitsSystem) Start() {
	for _, unit := range c.units {
		c.StartUnit(unit.Id())
		//unit.Start(c.iDataStorage)
	}
}

func (c *UnitsSystem) Stop() {
	logger.Println("UNITS_SYSTEM stopping begin")
	for _, unit := range c.units {
		go unit.Stop()
	}

	time.Sleep(100 * time.Millisecond) // Wait for units
	startedUnits := make([]string, 0)

	regularQuit := false
	for i := 0; i < 10; i++ {
		startedUnits = make([]string, 0)
		for _, unit := range c.units {
			if unit.IsStarted() {
				startedUnits = append(startedUnits, fmt.Sprint(unit.Id(), " / ", unit.Type(), " / ", unit.DisplayName()))
			}
		}
		if len(startedUnits) == 0 {
			regularQuit = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !regularQuit {
		logger.Println("Units stopping - timeout")
		for _, startedUnit := range startedUnits {
			logger.Println("Started: ", startedUnit)
		}
	}
	logger.Println("UNITS_SYSTEM stopping end")
}

func (c *UnitsSystem) MakeUnitByType(unitType string) common_interfaces.IUnit {
	var unit common_interfaces.IUnit

	for _, st := range c.unitTypes {
		if st.TypeCode == unitType {
			unit = st.Constructor()
			break
		}
	}

	if unit != nil {
		unit.Init()
	}

	return unit
}

func (c *UnitsSystem) AddUnit(unitType string, unitId string, displayName string, config string, fromCloud bool) (common_interfaces.IUnit, error) {
	var unit common_interfaces.IUnit
	nameIsExists := false
	c.mtx.Lock()
	if len(unitId) == 0 {
		maxUnitId := uint64(0)

		for _, u := range c.units { // 123
			uId := u.Id()
			if len(uId) > 1 && uId[0] == 'u' {
				uIdInt, uIdParseError := strconv.ParseUint(uId[1:], 10, 64)
				if uIdParseError == nil {
					if uIdInt > maxUnitId {
						maxUnitId = uIdInt
					}
				}
			}
		}
		maxUnitId++
		unitId = "u" + strconv.FormatUint(maxUnitId, 10)
	}

	for _, s := range c.units {
		if s.DisplayName() == displayName {
			nameIsExists = true
		}
	}
	c.mtx.Unlock()

	if fromCloud {
		if unitType == "general_cgi" || unitType == "general_cgi_key_value" {
			return nil, errors.New("cannot edit a cgi-unit via the Cloud")
		}
	}

	funcToOutput := func(ch chan common_interfaces.UnitMessage) {
		for msg := range ch {
			c.output <- msg
		}
	}

	if !nameIsExists {
		unit = c.MakeUnitByType(unitType)
		if unit != nil {
			unit.SetId(unitId)
			unit.SetDisplayName(displayName)
			unit.SetType(unitType)
			unit.SetConfig(config)
			unit.SetIUnit(unit)
			c.units = append(c.units, unit)
			go funcToOutput(unit.OutputChannel())
		} else {
			return nil, errors.New("cannot create unit")
		}
	} else {
		return nil, errors.New("unit already exists")
	}
	return unit, nil
}

func (c *UnitsSystem) GetUnitState(unitId string) (nodeinterface.UnitStateResponse, error) {
	var unit common_interfaces.IUnit
	c.mtx.Lock()
	for _, s := range c.units {
		if s.Id() == unitId {
			unit = s
		}
	}
	c.mtx.Unlock()

	if unit != nil {
		var unitState nodeinterface.UnitStateResponse
		unitState.Status = ""
		unitState.UnitDisplayName = unit.DisplayName()
		unitState.MainItem = unit.MainItem()
		unitState.Type = unit.Type()
		unitState.TypeName = c.UnitTypeForDisplayByType(unit.Type())
		if unit.IsStarted() {
			unitState.Status = "started"
		} else {
			unitState.Status = "stopped"
		}

		return unitState, nil
	}
	return nodeinterface.UnitStateResponse{}, errors.New("no unit found")
}

func (c *UnitsSystem) GetUnitStateAll() (nodeinterface.UnitStateAllResponse, error) {
	var result nodeinterface.UnitStateAllResponse
	result.Items = make([]nodeinterface.UnitStateAllResponseItem, 0)

	c.mtx.Lock()
	for _, unit := range c.units {
		var unitState nodeinterface.UnitStateAllResponseItem
		unitState.Status = ""
		unitState.UnitId = unit.Id()
		unitState.UnitDisplayName = unit.DisplayName()
		unitState.Type = unit.Type()
		unitState.TypeName = c.UnitTypeForDisplayByType(unit.Type())
		unitState.MainItem = unit.MainItem()
		if unit.IsStarted() {
			unitState.Status = "started"
		} else {
			unitState.Status = "stopped"
		}
		result.Items = append(result.Items, unitState)
	}
	c.mtx.Unlock()

	return result, nil
}

func (c *UnitsSystem) ListOfUnits() nodeinterface.UnitListResponse {
	var result nodeinterface.UnitListResponse
	result.Items = make([]nodeinterface.UnitListResponseItem, 0)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, s := range c.units {
		var sens nodeinterface.UnitListResponseItem
		sens.Id = s.Id()
		sens.Type = s.Type()
		sens.DisplayName = s.DisplayName()
		sens.Enable = s.IsStarted()
		sens.TypeForDisplay = c.UnitTypeForDisplayByType(s.Type())
		sens.Config = s.GetConfig()
		result.Items = append(result.Items, sens)
	}
	return result
}

func (c *UnitsSystem) Units() []units_common.UnitInfo {
	var result []units_common.UnitInfo
	result = make([]units_common.UnitInfo, 0)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, s := range c.units {
		var sens units_common.UnitInfo
		sens.Id = s.Id()
		sens.Type = s.Type()
		sens.DisplayName = s.DisplayName()
		sens.Enable = s.IsStarted()
		sens.TypeForDisplay = c.UnitTypeForDisplayByType(s.Type())
		sens.Config = s.GetConfig()
		sens.Properties = s.PropGet()
		result = append(result, sens)
	}
	return result
}

func (c *UnitsSystem) StartUnit(unitId string) error {
	for _, s := range c.units {
		if s.Id() == unitId {
			s.Start()
		}
	}
	return nil
}

func (c *UnitsSystem) StopUnit(unitId string) error {
	for _, s := range c.units {
		if s.Id() == unitId {
			s.Stop()
		}
	}
	return nil
}

func (c *UnitsSystem) RemoveUnits(units []string) error {
	logger.Println("UnitsSystem RemoveUnits", units)
	c.mtx.Lock()

	var deletedUnit common_interfaces.IUnit
	var unitIndex int
	idsOfDeletedUnits := make([]string, 0)

	for _, unitToRemove := range units {
		for unitIndex, deletedUnit = range c.units {
			if deletedUnit.Id() == unitToRemove {
				logger.Println("UnitsSystem RemoveUnits unit", deletedUnit.Id())
				idsOfDeletedUnits = append(idsOfDeletedUnits, deletedUnit.Id())
				logger.Println("UnitsSystem RemoveUnits stopping unit", deletedUnit.Id())
				deletedUnit.Stop()
				logger.Println("UnitsSystem RemoveUnits disposing unit", deletedUnit.Id())
				deletedUnit.Dispose()
				logger.Println("UnitsSystem RemoveUnits unit is disposed", deletedUnit.Id())
				c.units = append(c.units[:unitIndex], c.units[unitIndex+1:]...)
				break
			}
		}
	}

	logger.Println("UnitsSystem RemoveUnits removed")

	c.mtx.Unlock()

	logger.Println("UnitsSystem RemoveUnits deleting items")
	for _, idOfDeletedUnit := range idsOfDeletedUnits {
		c.output <- &common_interfaces.UnitMessageRemoteItemsOfUnit{
			UnitId: idOfDeletedUnit,
		}
	}
	logger.Println("UnitsSystem RemoveUnits items is deleted")

	return nil
}

func (c *UnitsSystem) GetUnitDisplayName(unitId string) (string, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, s := range c.units {
		if s.Id() == unitId {
			return s.DisplayName(), nil
		}
	}
	return "", errors.New("no unit found")
}

func (c *UnitsSystem) GetConfig(unitId string) (string, string, string, string, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, s := range c.units {
		if s.Id() == unitId {
			return s.DisplayName(), s.GetConfig(), s.GetConfigMeta(), s.Type(), nil
		}
	}
	return "", "", "", "", errors.New("no unit found")
}

func (c *UnitsSystem) GetConfigByType(unitType string) (string, string, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	for _, st := range c.unitTypes {
		if st.TypeCode == unitType {
			sens := c.MakeUnitByType(st.TypeCode)
			if sens != nil {
				return st.DisplayName, sens.GetConfigMeta(), nil
			} else {
				return "", "", errors.New("no unit type found")
			}
		}
	}

	return "", "", errors.New("no unit type found")
}

func (c *UnitsSystem) SetConfig(unitId string, name string, config string, fromCloud bool) error {
	var unit common_interfaces.IUnit

	c.mtx.Lock()
	for _, s := range c.units {
		if s.Id() == unitId {
			unit = s
		}
	}
	c.mtx.Unlock()

	if unit != nil {
		if fromCloud {
			if unit.Type() == "general_cgi" || unit.Type() == "general_cgi_key_value" {
				return errors.New("cannot edit a cgi-unit via the Cloud")
			}
		}
	}

	if unit != nil {

		unit.Stop()
		oldName := unit.DisplayName()

		if oldName != name {

			nameIsExists := false
			c.mtx.Lock()
			for _, s := range c.units {
				if s.DisplayName() == name {
					nameIsExists = true
				}
			}
			c.mtx.Unlock()

			if !nameIsExists {
				unit.SetDisplayName(name)
				//c.iDataStorage.RenameItems(oldName+"/", name+"/")
			}
		}

		unit.SetConfig(config)

		unit.Start()
	}

	return nil
}

func (c *UnitsSystem) SendToWatcher(unitId string, itemName string, value common_interfaces.ItemValue) {
	var targetUnit common_interfaces.IUnit

	c.mtx.Lock()
	for _, unit := range c.units {
		if unit.Id() == unitId {
			targetUnit = unit
			break
		}
	}
	c.mtx.Unlock()

	if targetUnit != nil {
		targetUnit.ItemChanged(itemName, value)
	}

}

func (c *UnitsSystem) UnitPropSet(unitId string, props []nodeinterface.PropItem) error {
	var err error
	var targetUnit common_interfaces.IUnit
	c.mtx.Lock()
	for _, unit := range c.units {
		if unit.Id() == unitId {
			targetUnit = unit
			break
		}
	}
	if targetUnit != nil {
		properties := make([]common_interfaces.ItemProperty, 0)
		for _, prop := range props {
			properties = append(properties, common_interfaces.ItemProperty{
				Name:  prop.PropName,
				Value: prop.PropValue,
			})
		}
		targetUnit.PropSet(properties)
	} else {
		err = errors.New("no unit found")
	}
	c.mtx.Unlock()
	return err
}

func (c *UnitsSystem) UnitPropGet(unitId string) ([]nodeinterface.PropItem, error) {
	var err error
	result := make([]nodeinterface.PropItem, 0)
	var targetUnit common_interfaces.IUnit
	c.mtx.Lock()
	for _, unit := range c.units {
		if unit.Id() == unitId {
			targetUnit = unit
			break
		}
	}
	if targetUnit != nil {
		props := targetUnit.PropGet()
		for _, prop := range props {
			result = append(result, nodeinterface.PropItem{
				PropName:  prop.Name,
				PropValue: prop.Value,
			})
		}
	} else {
		err = errors.New("no unit found")
	}
	c.mtx.Unlock()
	return result, err
}
