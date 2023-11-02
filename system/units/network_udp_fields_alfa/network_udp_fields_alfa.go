package network_udp_fields_alfa

import (
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/uom"
)

type UnitUdpFields struct {
	units_common.Unit
	port           int
	typeOfData     string
	sizeOfDataItem int
}

func New() iunit.IUnit {
	var c UnitUdpFields
	return &c
}

const (
	ItemNameSource = "source"
	ItemNameStatus = "status"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Network.UDP.Fields.Alfa"
	info.Category = "network"
	info.DisplayName = "UDP fields"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitUdpFields) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("port", "UDP Port", "7401", "num", "1", "65535", "0")
	meta.Add("data_type", "Data Type", "int32", "string", "int8|uint8|int16|uint16|int32|uint32|int64|uint64|float32|float64", "int8|uint8|int16|uint16|int32|uint32|int64|uint64|float32|float64", "")
	return meta.Marshal()
}

func (c *UnitUdpFields) InternalUnitStart() error {
	var err error

	type Config struct {
		Port     float64 `json:"port"`
		DataType string  `json:"data_type"`
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

	c.typeOfData = config.DataType
	c.sizeOfDataItem = 0
	switch c.typeOfData {
	case "int8":
		c.sizeOfDataItem = 1
	case "uint8":
		c.sizeOfDataItem = 1
	case "int16":
		c.sizeOfDataItem = 2
	case "uint16":
		c.sizeOfDataItem = 2
	case "int32":
		c.sizeOfDataItem = 4
	case "uint32":
		c.sizeOfDataItem = 4
	case "int64":
		c.sizeOfDataItem = 8
	case "uint64":
		c.sizeOfDataItem = 8
	case "float32":
		c.sizeOfDataItem = 4
	case "float64":
		c.sizeOfDataItem = 8
	}

	if c.sizeOfDataItem == 0 {
		err = errors.New("wrong data type")
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
		return err
	}

	c.SetMainItem(ItemNameStatus)
	c.SetString(ItemNameStatus, "started", "-")

	go c.Tick()
	return nil
}

func (c *UnitUdpFields) InternalUnitStop() {
}

func (c *UnitUdpFields) Tick() {
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
		itemNameFormat := "item_%03d"
		// Parse
		for i := 0; i < len(data); i += c.sizeOfDataItem {
			indexOfItem := i / c.sizeOfDataItem
			switch c.typeOfData {
			case "int8":
				val := data[i]
				c.SetUInt8(fmt.Sprintf(itemNameFormat, indexOfItem), uint8(val), "")
			case "uint8":
				val := data[i]
				c.SetInt8(fmt.Sprintf(itemNameFormat, indexOfItem), int8(val), "")
			case "int16":
				val := binary.LittleEndian.Uint16(data[i:])
				c.SetInt16(fmt.Sprintf(itemNameFormat, indexOfItem), int16(val), "")
			case "uint16":
				val := binary.LittleEndian.Uint16(data[i:])
				c.SetUInt16(fmt.Sprintf(itemNameFormat, indexOfItem), uint16(val), "")
			case "int32":
				val := binary.LittleEndian.Uint32(data[i:])
				c.SetInt32(fmt.Sprintf(itemNameFormat, indexOfItem), int32(val), "")
			case "uint32":
				val := binary.LittleEndian.Uint32(data[i:])
				c.SetUInt32(fmt.Sprintf(itemNameFormat, indexOfItem), uint32(val), "")
			case "int64":
				val := binary.LittleEndian.Uint64(data[i:])
				c.SetInt64(fmt.Sprintf(itemNameFormat, indexOfItem), int64(val), "")
			case "uint64":
				val := binary.LittleEndian.Uint64(data[i:])
				c.SetUInt64(fmt.Sprintf(itemNameFormat, indexOfItem), uint64(val), "")
			case "float32":
				val := math.Float32frombits(binary.LittleEndian.Uint32(data[i:]))
				c.SetFloat32(fmt.Sprintf(itemNameFormat, indexOfItem), float32(val), "", 3)
			case "float64":
				val := math.Float32frombits(binary.LittleEndian.Uint32(data[i:]))
				c.SetFloat64(fmt.Sprintf(itemNameFormat, indexOfItem), float64(val), "", 3)
			}

		}
	}

	if conn != nil {
		conn.Close()
		conn = nil
	}
	c.SetString(ItemNameStatus, "", uom.STOPPED)
	c.Started = false
}
