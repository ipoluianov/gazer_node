package files_file_size_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitFileSize struct {
	units_common.Unit
	fileName string
	periodMs int
}

func New() iunit.IUnit {
	var c UnitFileSize
	return &c
}

const (
	ItemNameSize   = "FileSize"
	ItemNameResult = "Result"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Files.File.Size.Alfa"
	info.Category = "file"
	info.DisplayName = "File Size"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitFileSize) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("file_name", "File Name", "file.txt", "string", "", "", "")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "0")
	return meta.Marshal()
}

func (c *UnitFileSize) InternalUnitStart() error {
	var err error
	c.SetString(ItemNameSize, "", "")
	c.SetMainItem(ItemNameSize)

	type Config struct {
		FileName string  `json:"file_name"`
		Period   float64 `json:"period"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameSize, err.Error(), "error")
		return err
	}

	c.fileName = config.FileName
	if c.fileName == "" {
		err = errors.New("wrong file")
		c.SetString(ItemNameSize, err.Error(), "error")
		return err
	}

	c.periodMs = int(config.Period)
	if c.periodMs < 100 {
		err = errors.New("wrong period")
		c.SetString(ItemNameSize, err.Error(), "error")
		return err
	}

	go c.Tick()
	return nil
}

func (c *UnitFileSize) InternalUnitStop() {
}

func (c *UnitFileSize) Tick() {
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

		stat, err := os.Stat(c.fileName)
		if err == nil {
			c.SetString(ItemNameSize, fmt.Sprint(stat.Size()), "bytes")
			//c.SetError("")
		} else {
			c.SetString(ItemNameSize, "", "")
			//c.SetError(err.Error())
		}
	}
	c.SetString(ItemNameSize, "", "stopped")
	c.Started = false
}
