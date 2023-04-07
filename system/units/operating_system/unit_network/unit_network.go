package unit_network

import "net"

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
