package serialport_keyvalue_watcher_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/tarm/serial"
)

type ConfigItem struct {
	Name      string `json:"name"`
	UOM       string `json:"uom"`
	IsControl bool   `json:"is_control"`
}

type UnitSerialPortKeyValue struct {
	units_common.Unit
	serialConfig *serial.Config
	serialPort   *serial.Port
	inputBuffer  []byte

	port       string
	receiveAll bool
	items      map[string]*ConfigItem

	receivedVariables map[string]string
}

func New() common_interfaces.IUnit {
	var c UnitSerialPortKeyValue
	c.inputBuffer = make([]byte, 0)
	c.receivedVariables = make(map[string]string)
	c.items = make(map[string]*ConfigItem)
	return &c
}

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "SerialPort.KeyValue.Watcher.Alfa"
	info.Category = "serial_port"
	info.DisplayName = "Serial Port Key=Value"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitSerialPortKeyValue) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("port", "Serial Port", "COM1", "string", "", "", "serial-ports")
	meta.Add("baud", "Baud", "9600", "num", "0", "999999999", "0")
	meta.Add("data_size", "Data Size", "8", "num", "4", "8", "0")
	meta.Add("parity", "Parity", "none", "string", "", "", "serial-port-parity")
	meta.Add("stop_bits", "Stop Bits", "1", "string", "", "", "serial-port-stop-bits")
	meta.Add("receive_all", "Receive All", "true", "bool", "", "", "")
	t1 := meta.Add("items", "Elements", "", "table", "", "", "")
	t1.Add("name", "ID", "item_name", "string", "", "", "")
	t1.Add("uom", "UOM", "V", "string", "", "", "")
	t1.Add("is_control", "IsControl", "false", "bool", "", "", "")
	return meta.Marshal()
}

func (c *UnitSerialPortKeyValue) InternalInitItems() {
	// c.SetStringForAll("", uom.STARTED)
}

func (c *UnitSerialPortKeyValue) InternalDeInitItems() {
	// c.SetStringForAll("", uom.STOPPED)
}

func (c *UnitSerialPortKeyValue) InternalUnitStart() error {
	var err error
	c.SetString("status", "starting", "")
	c.SetMainItem("status")

	c.port = ""

	type Config struct {
		Port     string  `json:"port"`
		Baud     float64 `json:"baud"`
		DataSize float64 `json:"data_size"`
		Parity   string  `json:"parity"`
		StopBits string  `json:"stop_bits"`

		ReceiveAll bool          `json:"receive_all"`
		Items      []*ConfigItem `json:"items"`
	}

	var config Config
	conf := c.GetConfig()
	logger.Println("SerialPort Config: ", conf)
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		logger.Println("ERROR[UnitSerialPortKeyValue]:", err)
		err = errors.New("config error")
		c.SetString("status", err.Error(), "error")
		return err
	}

	c.port = config.Port

	if runtime.GOOS == "windows" {
		c.port = "\\\\.\\" + config.Port
	}

	c.receiveAll = config.ReceiveAll

	c.items = make(map[string]*ConfigItem)

	parity := serial.ParityNone
	if config.Parity == "none" {
		parity = serial.ParityNone
	}
	if config.Parity == "odd" {
		parity = serial.ParityOdd
	}
	if config.Parity == "even" {
		parity = serial.ParityEven
	}
	if config.Parity == "mark" {
		parity = serial.ParityMark
	}
	if config.Parity == "space" {
		parity = serial.ParitySpace
	}

	stopBits := serial.Stop1
	if config.StopBits == "1" {
		stopBits = serial.Stop1
	}
	if config.StopBits == "1.5" {
		stopBits = serial.Stop1Half
	}
	if config.StopBits == "2" {
		stopBits = serial.Stop2
	}

	c.serialConfig = &serial.Config{
		Name:        c.port,
		Baud:        int(config.Baud),
		ReadTimeout: 100 * time.Millisecond,
		Size:        byte(config.DataSize),
		Parity:      parity,
		StopBits:    stopBits,
	}

	c.receivedVariables = make(map[string]string)

	go c.Tick()
	return nil
}

func (c *UnitSerialPortKeyValue) InternalUnitStop() {
	if c.serialPort != nil {
		c.serialPort.Close()
	}
}

func (c *UnitSerialPortKeyValue) Tick() {
	var err error
	c.Started = true

	for !c.Stopping {
		if c.serialPort == nil {
			c.LogInfo("try to open serial port " + c.serialConfig.Name)
			c.serialPort, err = serial.OpenPort(c.serialConfig)
			if err != nil {
				c.serialPort = nil
				c.SetString("status", err.Error(), "error")
				c.SetError(err.Error())
				c.LogError(err.Error())
				for vName := range c.receivedVariables {
					c.SetString(vName, "", "error")
				}
				for i := 0; i < 10; i++ {
					if c.Stopping {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				if c.Stopping {
					break
				}
			} else {
				c.SetString("status", "connected", "")
				//c.SetString("status", "waiting for data", "")
			}
		}

		if c.serialPort != nil {
			buffer := make([]byte, 32)
			n, err := c.serialPort.Read(buffer)
			if err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), "eof") {
					c.serialPort.Close()
					c.serialPort = nil
					c.SetString("status", err.Error(), "error")
					for vName := range c.receivedVariables {
						c.SetString(vName, "", "error")
					}
				}
			} else {
				if n > 0 {
					c.inputBuffer = append(c.inputBuffer, buffer[:n]...)

					found := true
					for found {
						found = false
						currentLine := make([]byte, 0)
						for index, b := range c.inputBuffer {
							if b == 10 || b == 13 {
								// parse currentLine
								if len(currentLine) > 0 {
									parts := strings.Split(string(currentLine), "=")
									if len(parts) > 1 {
										if len(parts[0]) > 0 {
											key := parts[0]
											value := parts[1]

											if item, ok := c.items[key]; ok {
												finalValue := value
												valueAsFloat, err := strconv.ParseFloat(value, 64)
												if err == nil {
													finalValue = strconv.FormatFloat(valueAsFloat, 'f', -1, 64)
												}
												c.receivedVariables[key] = finalValue
												c.SetString(key, finalValue, item.UOM)
											} else {
												if c.receiveAll {
													finalValue := value
													c.receivedVariables[key] = finalValue
													c.SetString(key, finalValue, "")
												}
											}

											time.Sleep(100 * time.Microsecond)
										}
									}

								}
								c.inputBuffer = c.inputBuffer[index+1:]
								found = true
								break
							} else {
								if b >= 32 && b < 128 {
									currentLine = append(currentLine, b)
								}
							}
						}
					}
				}
			}
		}
	}

	if c.serialPort != nil {
		c.serialPort.Close()
		c.serialPort = nil
	}

	for vName := range c.receivedVariables {
		c.SetString(vName, "", "stopped")
	}

	c.receivedVariables = make(map[string]string)

	c.SetString("status", "", "stopped")
	c.Started = false
}

func (c *UnitSerialPortKeyValue) ItemChanged(itemId uint64, itemName string, value common_interfaces.ItemValue) {
	if c.serialPort == nil {
		return
	}

	if strings.HasPrefix(itemName, c.Id()+"/") {
		countChartsToRemove := len(c.Id() + "/")
		localName := itemName[countChartsToRemove:]
		logger.Println("Send to Serial", "["+localName+"] =", value.Value)
		strForSend := localName + "=" + value.Value + "\r\n"
		_, err := c.serialPort.Write([]byte(strForSend))
		if err != nil {
			logger.Println("Send to Serial ERROR", "["+localName+"] =", value.Value)
		}
	}
}
