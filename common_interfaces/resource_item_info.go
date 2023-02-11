package common_interfaces

type ResourcesItemInfo struct {
	Id         string          `json:"id"`
	Type       string          `json:"type"`
	Properties []*ItemProperty `json:"p"`
}

func (c *ResourcesItemInfo) GetProp(name string) string {
	for _, p := range c.Properties {
		if p.Name == name {
			return p.Value
		}
	}
	return ""
}
