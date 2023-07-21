package unit_network

import (
	"net"

	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Computer.Network.Adapters.Alfa"
	info.Category = "computer"
	info.DisplayName = "Network"
	info.Constructor = New
	info.ImgBytes = nil
	info.Description = ""
	return info
}

func (c *UnitNetwork) writeAddresses(ni net.Interface) {
	// Addresses
	addrs, err := ni.Addrs()
	if err == nil {
		addrsString := ""
		for _, a := range addrs {
			if len(addrsString) > 0 {
				addrsString += " "
			}
			addrsString += a.String()
		}
		if c.addressesOfInterfaces[ni.Index] != addrsString {
			c.addressesOfInterfaces[ni.Index] = addrsString
			c.SetString(ni.Name+"/Addresses", addrsString, "-")
		}
	}
}
