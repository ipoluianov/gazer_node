package unit_udp_int_fields

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"time"

	"github.com/gazercloud/gazernode/common_interfaces"
	"github.com/gazercloud/gazernode/resources"
	"github.com/gazercloud/gazernode/system/units/units_common"
	"github.com/gazercloud/gazernode/utilities/uom"
)

type UnitUdpIntFields struct {
	units_common.Unit
	port int
}

func New() common_interfaces.IUnit {
	var c UnitUdpIntFields
	return &c
}

const (
	ItemNameSource = "source"
	ItemNameStatus = "status"
)

var Image []byte

func init() {
	Image = resources.R_files_sensors_unit_network_tcp_connect_png
}

func (c *UnitUdpIntFields) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("port", "UDP Port", "7401", "num", "1", "65535", "0")
	return meta.Marshal()
}

func (c *UnitUdpIntFields) InternalUnitStart() error {
	var err error

	type Config struct {
		Port float64 `json:"port"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}

	c.port = int(math.Round(config.Port))
	if c.port < 1 {
		err = errors.New("wrong port (<1)")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}
	if c.port > 65535 {
		err = errors.New("wrong timeout (>65535)")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}

	c.SetMainItem(ItemNameStatus)
	c.SetString(ItemNameStatus, "started", "-")

	go c.Tick()
	return nil
}

func (c *UnitUdpIntFields) InternalUnitStop() {
}

func (c *UnitUdpIntFields) Tick() {
	c.Started = true

	var conn net.PacketConn
	var err error

	for !c.Stopping {
		if conn == nil {
			conn, err = net.ListenPacket("udp", ":"+fmt.Sprint(c.port))
			if err != nil {
				c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
				return
			}
			c.SetString(ItemNameStatus, "listening port:"+fmt.Sprint(c.port), "-")
		}

		err = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		var n int
		var addr net.Addr
		buffer := make([]byte, 10*1024)
		n, addr, err = conn.ReadFrom(buffer)
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		}
		if err != nil {
			c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
			return
		}
		c.SetString(ItemNameSource, addr.String(), "-")
		data := buffer[:n]
		_ = data
		// Parse
		for i := 0; i < len(data); i += 4 {
			val := binary.LittleEndian.Uint32(data[i:])
			c.SetInt32("item_"+fmt.Sprint(i/4), int32(val), "")
		}
	}
	c.SetString(ItemNameStatus, "", uom.STOPPED)
	c.Started = false
}
