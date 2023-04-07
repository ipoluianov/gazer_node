package server

import (
	"encoding/json"

	"github.com/ipoluianov/gazer_node/system/protocols/nodeinterface"
)

func (c *Server) UnitTypeList(request []byte) (response []byte, err error) {
	var req nodeinterface.UnitTypeListRequest
	var resp nodeinterface.UnitTypeListResponse
	req.Offset = 0
	req.Category = ""
	req.Filter = ""
	req.MaxCount = 1000
	err = json.Unmarshal(request, &req)
	if err != nil {
		return
	}

	resp = c.system.UnitTypes(req.Category, req.Filter, req.Offset, req.MaxCount)

	response, err = json.MarshalIndent(resp, "", " ")
	return
}

func (c *Server) UnitTypeCategories(request []byte) (response []byte, err error) {
	var req nodeinterface.UnitTypeCategoriesRequest
	var resp nodeinterface.UnitTypeCategoriesResponse
	err = json.Unmarshal(request, &req)
	if err != nil {
		return
	}

	resp = c.system.UnitCategories()

	response, err = json.MarshalIndent(resp, "", " ")
	return
}

func (c *Server) UnitTypeConfigMeta(request []byte) (response []byte, err error) {
	var req nodeinterface.UnitTypeConfigMetaRequest
	var resp nodeinterface.UnitTypeConfigMetaResponse
	err = json.Unmarshal(request, &req)
	if err != nil {
		return
	}

	resp.UnitType, resp.UnitTypeConfigMeta, err = c.system.GetConfigByType(req.UnitType)
	if err != nil {
		return
	}

	response, err = json.MarshalIndent(resp, "", " ")
	return
}
