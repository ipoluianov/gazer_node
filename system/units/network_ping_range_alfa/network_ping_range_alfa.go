package network_ping_range_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math"
	"net"
	"runtime"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities"
	"github.com/ipoluianov/gazer_node/utilities/gazerping"
	"github.com/ipoluianov/gazer_node/utilities/uom"
)

type UnitPingRange struct {
	units_common.Unit

	addr1     string
	ip1       net.IP
	addr2     string
	ip2       net.IP
	timeoutMs int
	frameSize int

	currentIp    net.IP
	currentCount int
}

func New() iunit.IUnit {
	var c UnitPingRange
	return &c
}

const (
	ItemNameAddress1 = "Address1"
	ItemNameAddress2 = "Address2"
	ItemNameTime     = "Time"
	ItemNameIP       = "IP"
	ItemNameDataSize = "DataSize"
	ItemLiveCount    = "LiveCount"
	ItemWorkCounter  = "WorkCounter"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Network.Ping.Range.Alfa"
	info.Category = "network"
	info.DisplayName = "Ping Range"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitPingRange) ipToUInt32(ip net.IP) uint32 {
	var result uint32
	result = (uint32(ip[0]) << 24) | (uint32(ip[1]) << 16) | (uint32(ip[2]) << 8) | (uint32(ip[3]) << 0)
	return result
}

func (c *UnitPingRange) incrementIP(ip net.IP) net.IP {
	result := make([]byte, 4)
	copy(result, ip)
	ip = result

	// Increment IP
	if ip[3] != 254 {
		ip[3]++
	} else {
		ip[3] = 1
		if ip[2] != 254 {
			ip[2]++
		} else {
			ip[2] = 1
			if ip[1] != 254 {
				ip[1]++
			} else {
				ip[1] = 1
				if ip[0] != 254 {
					ip[0]++
				}
			}
		}
	}

	return result
}

func (c *UnitPingRange) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("addr1", "Address", "localhost", "string", "", "", "")
	meta.Add("addr2", "Address", "localhost", "string", "", "", "")
	meta.Add("timeout", "Timeout, ms", "1000", "num", "100", "10000", "0")
	meta.Add("frame_size", "Frame Size, bytes", "64", "num", "4", "1400", "0")
	return meta.Marshal()
}

func (c *UnitPingRange) InternalUnitStart() error {
	var err error
	c.SetMainItem(ItemNameTime)

	c.currentCount = 0

	type Config struct {
		Addr1     string  `json:"addr1"`
		Addr2     string  `json:"addr2"`
		Timeout   float64 `json:"timeout"`
		FrameSize float64 `json:"frame_size"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameTime, err.Error(), "error")
		return err
	}

	c.addr1 = config.Addr1
	if c.addr1 == "" {
		err = errors.New("wrong address")
		c.SetString(ItemNameTime, err.Error(), "error")
		return err
	}

	c.addr2 = config.Addr2
	if c.addr2 == "" {
		err = errors.New("wrong address")
		c.SetString(ItemNameTime, err.Error(), "error")
		return err
	}

	c.ip1 = net.ParseIP(c.addr1).To4()
	c.ip2 = net.ParseIP(c.addr2).To4()
	c.currentIp = net.ParseIP(c.addr1).To4()

	c.timeoutMs = int(math.Round(config.Timeout))
	if c.timeoutMs < 100 {
		err = errors.New("wrong timeout (<100)")
		c.SetString(ItemNameTime, err.Error(), "error")
		return err
	}
	if c.timeoutMs > 10000 {
		err = errors.New("wrong timeout (>10000)")
		c.SetString(ItemNameTime, err.Error(), "error")
		return err
	}

	c.frameSize = int(math.Round(config.FrameSize))
	if c.frameSize < 1 {
		err = errors.New("wrong Frame Size (<1)")
		c.SetString(ItemNameTime, err.Error(), "error")
		return err
	}
	if c.frameSize > 1400 {
		err = errors.New("wrong FrameSize (>1400)")
		c.SetString(ItemNameTime, err.Error(), "error")
		return err
	}

	c.SetString(ItemNameAddress1, c.addr1, "-")
	c.SetString(ItemNameAddress2, c.addr2, "-")
	//c.SetString(ItemNameTime, "", "")
	//c.SetString(ItemNameIP, "", "-")
	c.SetInt(ItemNameDataSize, c.frameSize, uom.BYTES)

	c.SetPropertyIfDoesntExist(ItemNameAddress1, "color", "#AA0000")

	go c.Tick()
	return nil
}

func (c *UnitPingRange) InternalUnitStop() {
}

func (c *UnitPingRange) Tick() {
	var lastError string
	var lastIP string

	c.SetString(ItemNameTime, "", "started")

	c.Started = true
	var dtLastPingTime time.Time
	for !c.Stopping {
		for {
			if c.Stopping || time.Now().Sub(dtLastPingTime) > time.Duration(10)*time.Millisecond {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			break
		}

		if !utilities.IsRoot() && runtime.GOOS == "linux" {
			c.SetString(ItemNameTime, "no root", "error")
			//c.SetError("ping.NewPinger: " + "no root")
			dtLastPingTime = time.Now().UTC()
			continue
		}

		addr := c.currentIp.String()

		if addr == "" {
			//c.SetError("ipaddress == ''")
			c.SetString(ItemNameTime, "wrong address", "error")
			continue
		}

		var err error

		var pingTime int
		var peer net.Addr

		useUdpSocket := (!utilities.IsRoot() && runtime.GOOS != "windows")

		c.SetString(ItemNameIP, addr, uom.NONE)

		pingTime, peer, err = gazerping.Ping(addr, c.frameSize, c.timeoutMs, useUdpSocket)

		if err == nil {
			c.currentCount++
			c.SetInt(ItemWorkCounter, c.currentCount, uom.NONE)
			ip := peer.String()
			if ip != lastIP {
				lastIP = ip
				//c.SetString(ItemNameIP, ip, "-")
			}
			if !c.Stopping {
				t := pingTime
				c.SetInt(ItemNameTime, t, uom.MS)
				if lastError != "" {
					//c.SetError("")
				}
				lastError = ""
			}
		} else {
			if lastError != err.Error() {
				lastError = err.Error()
				lastIP = ""
				//c.SetError(lastError)
				//c.SetString(ItemNameIP, lastIP, "error")
			}
			c.SetString(ItemNameTime, lastError, "error")
		}

		c.currentIp = c.incrementIP(c.currentIp)

		maxIp := c.ipToUInt32(c.ip2)
		curIp := c.ipToUInt32(c.currentIp)
		if curIp > maxIp {
			c.currentIp = c.ip1
			c.SetInt(ItemLiveCount, c.currentCount, uom.NONE)
			c.currentCount = 0
		}

		time.Sleep(10 * time.Millisecond)

		dtLastPingTime = time.Now().UTC()
	}
	c.SetString(ItemNameTime, "", "stopped")
	c.Started = false
}
