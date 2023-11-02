package files_file_content_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

type UnitFileContent struct {
	units_common.Unit
	fileName    string
	periodMs    int
	allowChange bool
	trim        bool
	parse       bool
	scale       float64
	offset      float64
	uom         string
	precision   int
}

func New() iunit.IUnit {
	var c UnitFileContent
	return &c
}

const (
	ItemNameContent     = "Content"
	ItemNameReadResult  = "ReadResult"
	ItemNameWriteResult = "WriteResult"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Files.File.Content.Alfa"
	info.Category = "file"
	info.DisplayName = "File Content"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitFileContent) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("file_name", "File Name", "file.txt", "string", "", "", "")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "0")
	meta.Add("allow_change", "Allow Change", "false", "bool", "", "", "")
	meta.Add("trim", "Trim", "false", "bool", "", "", "")
	meta.Add("parse", "Parse", "false", "bool", "", "", "")
	meta.Add("scale", "Scale", "1", "num", "-999999999", "99999999", "6")
	meta.Add("offset", "Offset", "0", "num", "-999999999", "99999999", "6")
	meta.Add("uom", "UOM", "", "string", "", "", "")
	meta.Add("precision", "Precision", "3", "num", "0", "99", "")
	return meta.Marshal()
}

func (c *UnitFileContent) InternalUnitStart() error {
	var err error
	c.SetString(ItemNameContent, "", "")
	c.SetMainItem(ItemNameContent)

	type Config struct {
		FileName    string  `json:"file_name"`
		Period      float64 `json:"period"`
		AllowChange bool    `json:"allow_change"`
		Trim        bool    `json:"trim"`
		ParseFloat  bool    `json:"parse"`
		Scale       float64 `json:"scale"`
		Offset      float64 `json:"offset"`
		UOM         string  `json:"uom"`
		Precision   float64 `json:"precision"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameContent, err.Error(), "error")
		return err
	}

	c.fileName = config.FileName
	if c.fileName == "" {
		err = errors.New("wrong file")
		c.SetString(ItemNameContent, err.Error(), "error")
		return err
	}

	c.periodMs = int(config.Period)
	if c.periodMs < 100 {
		err = errors.New("wrong period")
		c.SetString(ItemNameContent, err.Error(), "error")
		return err
	}

	c.precision = int(config.Precision)
	if c.precision < 0 || c.precision > 100 {
		err = errors.New("wrong precision")
		c.SetString(ItemNameContent, err.Error(), "error")
		return err
	}

	c.trim = config.Trim
	c.parse = config.ParseFloat
	c.scale = config.Scale
	c.offset = config.Offset
	c.uom = config.UOM
	c.allowChange = config.AllowChange

	go c.Tick()
	return nil
}

func (c *UnitFileContent) InternalUnitStop() {
}

func (c *UnitFileContent) Tick() {
	c.Started = true
	dtOperationTime := time.Now().UTC()
	for !c.Stopping {
		for {
			if c.Stopping || time.Now().Sub(dtOperationTime) > time.Duration(c.periodMs)*time.Millisecond {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			break
		}
		dtOperationTime = time.Now().UTC()

		content, err := os.ReadFile(c.fileName)

		if len(content) > 1024 {
			err = errors.New("too much data")
			content = content[:1024]
		}
		contentStr := string(content)
		if c.trim {
			contentStr = strings.Trim(contentStr, " \n\r\t")
		}

		var contentFloat float64

		if c.parse {
			contentFloat, err = strconv.ParseFloat(contentStr, 64)
			if err == nil {
				contentFloat = contentFloat*c.scale + c.offset
				contentStr = strconv.FormatFloat(contentFloat, 'f', c.precision, 64)
				if strings.Index(contentStr, ".") >= 0 {
					contentStr = strings.TrimRight(contentStr, "0")
				}
			}
		}

		if err == nil {
			c.SetString(ItemNameContent, contentStr, c.uom)
			c.SetError("")
			c.SetString(ItemNameReadResult, "Success", "")
		} else {
			c.SetString(ItemNameContent, string(content), "error")
			c.SetError(err.Error())
			c.SetString(ItemNameReadResult, err.Error(), "error")
		}
	}
	c.SetString(ItemNameContent, "", "stopped")
	c.Started = false
}

func (c *UnitFileContent) ItemChanged(itemId uint64, itemName string, value common_interfaces.ItemValue) {
	if c.allowChange {
		if c.ItemFullName(ItemNameContent) == itemName {
			err := os.WriteFile(c.fileName, []byte(value.Value), 0666)
			if err != nil {
				c.SetString(ItemNameWriteResult, err.Error(), "error")
				logger.Println("UnitFileContent::ItemChanged error:", err, "itemName:", itemName, "value:", value.Value)
			} else {
				c.SetString(ItemNameWriteResult, "Success", "")
				logger.Println("UnitFileContent::ItemChanged success", "itemName:", itemName, "value:", value.Value)
			}
		}
	}
}
