package server

import (
	"errors"

	"github.com/ipoluianov/gazer_node/protocols/nodeinterface"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

var errorCounter = 0

func (c *Server) RequestJson(function string, requestText []byte, host string, fromCloud bool, isGuest bool) ([]byte, error) {
	var err error
	var result []byte

	/*if errorCounter > 1 {
		errorCounter = 0
		return nil, errors.New("special error")
	}
	errorCounter++*/

	c.system.RegApiCall()

	if isGuest {
		if _, ok := c.guestAccess[function]; !ok {
			err = errors.New("access denied")
		}
	}

	if err == nil {
		switch function {

		// *** UnitType ***
		case nodeinterface.FuncUnitTypeList:
			result, err = c.UnitTypeList(requestText)
		case nodeinterface.FuncUnitTypeCategories:
			result, err = c.UnitTypeCategories(requestText)
		case nodeinterface.FuncUnitTypeConfigMeta:
			result, err = c.UnitTypeConfigMeta(requestText)

			// *** Unit ***
		case nodeinterface.FuncUnitAdd:
			result, err = c.UnitAdd(requestText, fromCloud)
		case nodeinterface.FuncUnitRemove:
			result, err = c.UnitRemove(requestText)
		case nodeinterface.FuncUnitState:
			result, err = c.UnitState(requestText)
		case nodeinterface.FuncUnitStateAll:
			result, err = c.UnitStateAll(requestText)
		case nodeinterface.FuncUnitItemsValues:
			result, err = c.UnitItemsValues(requestText)
		case nodeinterface.FuncUnitList:
			result, err = c.UnitList(requestText)
		case nodeinterface.FuncUnitStart:
			result, err = c.UnitStart(requestText)
		case nodeinterface.FuncUnitStop:
			result, err = c.UnitStop(requestText)
		case nodeinterface.FuncUnitSetConfig:
			result, err = c.UnitSetConfig(requestText, fromCloud)
		case nodeinterface.FuncUnitGetConfig:
			result, err = c.UnitGetConfig(requestText)
		case nodeinterface.FuncUnitPropSet:
			result, err = c.UnitPropSet(requestText)
		case nodeinterface.FuncUnitPropGet:
			result, err = c.UnitPropGet(requestText)

			// *** Service ***
		case nodeinterface.FuncServiceLookup:
			result, err = c.ServiceLookup(requestText)
		case nodeinterface.FuncServiceStatistics:
			result, err = c.ServiceStatistics(requestText)
		/*case nodeinterface.FuncServiceApi:
		result, err = c.ServiceApi(requestText)*/
		case nodeinterface.FuncServiceSetNodeName:
			result, err = c.ServiceSetNodeName(requestText)
		case nodeinterface.FuncServiceNodeName:
			result, err = c.ServiceNodeName(requestText)
		case nodeinterface.FuncServiceInfo:
			result, err = c.ServiceInfo(requestText)

			// *** Resource ***
		case nodeinterface.FuncResourceAdd:
			result, err = c.ResourceAdd(requestText)
		case nodeinterface.FuncResourceSetByPath:
			result, err = c.ResourceSetByPath(requestText)
		case nodeinterface.FuncResourceSet:
			result, err = c.ResourceSet(requestText)
		case nodeinterface.FuncResourceGet:
			result, err = c.ResourceGet(requestText)
		case nodeinterface.FuncResourceGetByPath:
			result, err = c.ResourceGetByPath(requestText)
		case nodeinterface.FuncResourceRemove:
			result, err = c.ResourceRemove(requestText)
		case nodeinterface.FuncResourceList:
			result, err = c.ResourceList(requestText)
		case nodeinterface.FuncResourcePropSet:
			result, err = c.ResourcePropSet(requestText)
		case nodeinterface.FuncResourcePropGet:
			result, err = c.ResourcePropGet(requestText)

		// *** Data Item ***
		case nodeinterface.FuncDataItemList:
			result, err = c.DataItemList(requestText)
		case nodeinterface.FuncDataItemListAll:
			result, err = c.DataItemListAll(requestText)
		case nodeinterface.FuncDataItemWrite:
			result, err = c.DataItemWrite(requestText)
		case nodeinterface.FuncDataItemHistory:
			result, err = c.DataItemHistory(requestText)
		case nodeinterface.FuncDataItemHistoryChart:
			result, err = c.DataItemHistoryChart(requestText)
		case nodeinterface.FuncDataItemRemove:
			result, err = c.DataItemRemove(requestText)
		case nodeinterface.FuncDataItemPropSet:
			result, err = c.DataItemPropSet(requestText)
		case nodeinterface.FuncDataItemPropGet:
			result, err = c.DataItemPropGet(requestText)
		default:
			err = errors.New("function not supported")
		}
	}

	if err == nil {
		return result, nil
	}

	logger.Println("Function execution error: ", err, "\r\n", function, string(requestText))
	return nil, err
}

var TempValue int

func init() {
	TempValue = 5
}

func (c *Server) initApiAccess() {
	c.guestAccess = make(map[string]bool)
	c.guestAccess[nodeinterface.FuncUnitTypeList] = true
	c.guestAccess[nodeinterface.FuncUnitTypeCategories] = true
	c.guestAccess[nodeinterface.FuncUnitTypeConfigMeta] = true

	c.guestAccess[nodeinterface.FuncUnitState] = true
	c.guestAccess[nodeinterface.FuncUnitStateAll] = true
	c.guestAccess[nodeinterface.FuncUnitItemsValues] = true
	c.guestAccess[nodeinterface.FuncUnitList] = true
	c.guestAccess[nodeinterface.FuncUnitGetConfig] = true
	c.guestAccess[nodeinterface.FuncUnitPropGet] = true

	c.guestAccess[nodeinterface.FuncDataItemList] = true
	c.guestAccess[nodeinterface.FuncDataItemListAll] = true
	c.guestAccess[nodeinterface.FuncDataItemHistory] = true
	c.guestAccess[nodeinterface.FuncDataItemHistoryChart] = true
	c.guestAccess[nodeinterface.FuncDataItemPropGet] = true

	//c.guestAccess[nodeinterface.FuncServiceLookup] = true
	c.guestAccess[nodeinterface.FuncServiceStatistics] = true
	c.guestAccess[nodeinterface.FuncServiceApi] = true
	c.guestAccess[nodeinterface.FuncServiceNodeName] = true
	c.guestAccess[nodeinterface.FuncServiceInfo] = true

	c.guestAccess[nodeinterface.FuncResourceGet] = true
	c.guestAccess[nodeinterface.FuncResourceGetByPath] = true
	c.guestAccess[nodeinterface.FuncResourceList] = true
	c.guestAccess[nodeinterface.FuncResourcePropGet] = true

	c.guestAccess[nodeinterface.FuncUnitTypeConfigMeta] = true
	c.guestAccess[nodeinterface.FuncUnitTypeConfigMeta] = true
	c.guestAccess[nodeinterface.FuncUnitTypeConfigMeta] = true
	c.guestAccess[nodeinterface.FuncUnitTypeConfigMeta] = true
	c.guestAccess[nodeinterface.FuncUnitTypeConfigMeta] = true

}
