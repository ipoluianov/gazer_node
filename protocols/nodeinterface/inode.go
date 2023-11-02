package nodeinterface

type INode interface {
	GetUnitState(unitId string) (UnitStateResponse, error)
}
