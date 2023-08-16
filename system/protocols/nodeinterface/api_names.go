package nodeinterface

const (
	// *** UnitType ***
	FuncUnitTypeList       = "unit_type_list"
	FuncUnitTypeCategories = "unit_type_categories"
	FuncUnitTypeConfigMeta = "unit_type_config_meta"

	// *** Unit ***
	FuncUnitAdd         = "unit_add"
	FuncUnitRemove      = "unit_remove"
	FuncUnitState       = "unit_state"
	FuncUnitStateAll    = "unit_state_all"
	FuncUnitItemsValues = "unit_items_values"
	FuncUnitList        = "unit_list"
	FuncUnitStart       = "unit_start"
	FuncUnitStop        = "unit_stop"
	FuncUnitSetConfig   = "unit_set_config"
	FuncUnitGetConfig   = "unit_get_config"
	FuncUnitPropSet     = "unit_prop_set"
	FuncUnitPropGet     = "unit_prop_get"

	// *** Data Item ***
	FuncDataItemList         = "data_item_list"
	FuncDataItemListAll      = "data_item_list_all"
	FuncDataItemWrite        = "data_item_write"
	FuncDataItemHistory      = "data_item_history"
	FuncDataItemHistoryChart = "data_item_history_chart"
	FuncDataItemRemove       = "data_item_remove"
	FuncDataItemPropSet      = "data_item_prop_set"
	FuncDataItemPropGet      = "data_item_prop_get"

	// *** Service ***
	FuncServiceLookup      = "service_lookup"
	FuncServiceStatistics  = "service_statistics"
	FuncServiceApi         = "service_api"
	FuncServiceSetNodeName = "service_set_node_name"
	FuncServiceNodeName    = "service_node_name"
	FuncServiceInfo        = "service_info"

	// *** Resource ***
	FuncResourceAdd       = "resource_add"
	FuncResourceSet       = "resource_set"
	FuncResourceGet       = "resource_get"
	FuncResourceGetByPath = "resource_get_by_path"
	FuncResourceRemove    = "resource_remove"
	FuncResourceList      = "resource_list"
	FuncResourcePropSet   = "resource_prop_set"
	FuncResourcePropGet   = "resource_prop_get"
)

func ApiFunctions() []string {
	res := make([]string, 0)
	res = append(res, FuncUnitTypeList)
	res = append(res, FuncUnitTypeCategories)
	res = append(res, FuncUnitTypeConfigMeta)

	res = append(res, FuncUnitAdd)
	res = append(res, FuncUnitRemove)
	res = append(res, FuncUnitState)
	res = append(res, FuncUnitStateAll)
	res = append(res, FuncUnitItemsValues)
	res = append(res, FuncUnitList)
	res = append(res, FuncUnitStart)
	res = append(res, FuncUnitStop)
	res = append(res, FuncUnitSetConfig)
	res = append(res, FuncUnitGetConfig)
	res = append(res, FuncUnitPropSet)
	res = append(res, FuncUnitPropGet)

	res = append(res, FuncDataItemList)
	res = append(res, FuncDataItemListAll)
	res = append(res, FuncDataItemWrite)
	res = append(res, FuncDataItemHistory)
	res = append(res, FuncDataItemHistoryChart)
	res = append(res, FuncDataItemRemove)
	res = append(res, FuncDataItemPropSet)
	res = append(res, FuncDataItemPropGet)

	res = append(res, FuncServiceLookup)
	res = append(res, FuncServiceStatistics)
	res = append(res, FuncServiceApi)
	res = append(res, FuncServiceSetNodeName)
	res = append(res, FuncServiceNodeName)
	res = append(res, FuncServiceInfo)

	res = append(res, FuncResourceAdd)
	res = append(res, FuncResourceSet)
	res = append(res, FuncResourceGet)
	res = append(res, FuncResourceRemove)
	res = append(res, FuncResourceList)

	return res
}
