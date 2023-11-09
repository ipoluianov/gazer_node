package computer_network_adapters_alfa

import (
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

type UnitNetwork struct {
	units_common.Unit

	addressesOfInterfaces map[int]string
}

func New() iunit.IUnit {
	var c UnitNetwork
	c.addressesOfInterfaces = make(map[int]string)
	return &c
}

func (c *UnitNetwork) InternalUnitStart() error {
	c.SetString("TotalSpeed", "", "")
	c.SetMainItem("TotalSpeed")

	interfaces, err := net.Interfaces()
	if err == nil {
		for _, ni := range interfaces {
			c.SetString(ni.Name+"/InSpeed", "", "")
			c.SetString(ni.Name+"/OutSpeed", "", "")
		}
		c.SetString("TotalInSpeed", "", "")
		c.SetString("TotalOutSpeed", "", "")
	} else {
		//c.SetError("")
	}

	go c.Tick()
	return nil
}

func (c *UnitNetwork) InternalUnitStop() {
}

func (c *UnitNetwork) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	return meta.Marshal()
}

func (c *UnitNetwork) Tick() {
	var err error
	c.Started = true

	type LastCounters struct {
		DT            time.Time
		TotalIn       uint64
		TotalOut      uint64
		TotalInBytes  uint64
		TotalOutBytes uint64
	}

	lastCounters := make(map[int]LastCounters)

	for !c.Stopping {
		for i := 0; i < 10; i++ {
			if c.Stopping {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		totalSpeed := 0.0
		totalInSpeed := 0.0
		totalOutSpeed := 0.0
		var interfaces []net.Interface
		interfaces, err = net.Interfaces()
		if err == nil {
			for _, ni := range interfaces {
				rxPackets := int64(0)
				rxBytes := int64(0)
				txPackets := int64(0)
				txBytes := int64(0)

				// Addresses
				addrs, err := ni.Addrs()
				if err != nil {
					addrsString := ""
					for _, a := range addrs {
						if len(addrsString) > 0 {
							addrsString += " "
						}
						addrsString += a.String()
					}
					c.SetString(ni.Name+"/Addresses", addrsString, "-")
				}

				rxPacketsStr, errParamRxPackets := ioutil.ReadFile("/sys/class/net/" + ni.Name + "/statistics/rx_packets")
				if errParamRxPackets == nil {
					rxPackets, errParamRxPackets = strconv.ParseInt(strings.ReplaceAll(string(rxPacketsStr), "\n", ""), 10, 64)
				} else {
					logger.Println(errParamRxPackets)
				}

				//c.SetString("errParamRxPackets", errParamRxPackets.Error(), "q")

				rxBytesStr, errParamRxBytes := ioutil.ReadFile("/sys/class/net/" + ni.Name + "/statistics/rx_bytes")
				if errParamRxBytes == nil {
					rxBytes, errParamRxBytes = strconv.ParseInt(strings.ReplaceAll(string(rxBytesStr), "\n", ""), 10, 64)
				} else {
					logger.Println(errParamRxBytes)
				}

				//c.SetString("errParamRxBytes", errParamRxBytes.Error(), "q")

				txPacketsStr, errParamTxPackets := ioutil.ReadFile("/sys/class/net/" + ni.Name + "/statistics/tx_packets")
				if errParamTxPackets == nil {
					txPackets, errParamTxPackets = strconv.ParseInt(strings.ReplaceAll(string(txPacketsStr), "\n", ""), 10, 64)
				} else {
					logger.Println(errParamTxPackets)
				}

				//c.SetString("errParamTxPackets", errParamTxPackets.Error(), "q")

				txBytesStr, errParamTxBytes := ioutil.ReadFile("/sys/class/net/" + ni.Name + "/statistics/tx_bytes")
				if errParamTxBytes == nil {
					txBytes, errParamTxBytes = strconv.ParseInt(strings.ReplaceAll(string(txBytesStr), "\n", ""), 10, 64)
				} else {
					logger.Println(errParamTxBytes)
				}

				//c.SetString("errParamTxBytes", errParamTxBytes.Error(), "q")

				totalIn := uint64(rxPackets)
				totalInBytes := uint64(rxBytes)
				totalOut := uint64(txPackets)
				totalOutBytes := uint64(txBytes)

				nowTime := time.Now()

				if true {
					if cs, ok := lastCounters[ni.Index]; ok {
						seconds := nowTime.Sub(cs.DT).Seconds()
						if seconds > 0.001 {
							c.SetFloat64(ni.Name+"/InSpeed", float64(totalInBytes-cs.TotalInBytes)/seconds/1024.0, "KB/sec", 1)
							c.SetFloat64(ni.Name+"/OutSpeed", float64(totalOutBytes-cs.TotalOutBytes)/seconds/1024.0, "KB/sec", 1)
							totalInSpeed += float64(totalInBytes-cs.TotalInBytes) / seconds / 1024.0
							totalOutSpeed += float64(totalOutBytes-cs.TotalOutBytes) / seconds / 1024.0
						}
					}

					lastCounters[ni.Index] = LastCounters{
						DT:            nowTime,
						TotalIn:       totalIn,
						TotalOut:      totalOut,
						TotalInBytes:  totalInBytes,
						TotalOutBytes: totalOutBytes,
					}
				} else {
					delete(lastCounters, ni.Index)
					c.SetString(ni.Name+"/InSpeed", "", "error")
					c.SetString(ni.Name+"/OutSpeed", "", "error")
				}

			}

			totalSpeed = totalInSpeed + totalOutSpeed
			c.SetFloat64("TotalInSpeed", totalInSpeed, "KB/sec", 1)
			c.SetFloat64("TotalOutSpeed", totalOutSpeed, "KB/sec", 1)
			c.SetFloat64("TotalSpeed", totalSpeed, "KB/sec", 1)
		} else {
			//c.SetError(err.Error())
		}
	}

	c.Started = false
}
