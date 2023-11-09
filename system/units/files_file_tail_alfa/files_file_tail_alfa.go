package files_file_tail_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitFileTail struct {
	units_common.Unit
	fileName string
	periodMs int
	size     int
}

func New() iunit.IUnit {
	var c UnitFileTail
	return &c
}

const (
	ItemNameContent = "Content"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Files.File.Tail.Alfa"
	info.Category = "file"
	info.DisplayName = "File Tail"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitFileTail) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("file_name", "File Name", "file.txt", "string", "", "", "")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "0")
	meta.Add("size", "Size, ms", "1000", "num", "1", "10000", "0")
	return meta.Marshal()
}

func (c *UnitFileTail) InternalUnitStart() error {
	var err error
	c.SetPropertyIfDoesntExist(ItemNameContent, "history_disabled", "true")

	c.SetString(ItemNameContent, "", "")
	c.SetMainItem(ItemNameContent)

	type Config struct {
		FileName string  `json:"file_name"`
		Period   float64 `json:"period"`
		Size     float64 `json:"size"`
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

	c.size = int(config.Size)
	if c.periodMs < 100 {
		err = errors.New("wrong size")
		c.SetString(ItemNameContent, err.Error(), "error")
		return err
	}

	go c.Tick()
	return nil
}

func (c *UnitFileTail) InternalUnitStop() {
}

func (c *UnitFileTail) Tick() {
	c.Started = true
	dtOperationTime := time.Now().UTC()
	for !c.Stopping {
		for {
			if c.Stopping || time.Since(dtOperationTime) > time.Duration(c.periodMs)*time.Millisecond {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			break
		}
		dtOperationTime = time.Now().UTC()

		file, err := os.OpenFile(c.fileName, os.O_RDONLY, 0666)
		if err == nil {
			fi, err := os.Lstat(c.fileName)
			if err == nil {
				fileOffset := fi.Size() - int64(c.size)
				chunkSize := c.size
				if fileOffset < 0 {
					fileOffset = 0
					chunkSize = int(fi.Size()) - int(fileOffset)
				}
				_, err = file.Seek(fileOffset, 0)
				if err == nil {
					buffer := make([]byte, chunkSize)
					_, err = file.Read(buffer)
					if err == nil {
						c.SetString(ItemNameContent, string(buffer), "")
						//c.SetError("")
					}
				}
			}
		}

		if err != nil {
			c.SetString(ItemNameContent, err.Error(), "error")
			//c.SetError(err.Error())
		}
	}
	c.SetString(ItemNameContent, "", "stopped")
	c.Started = false
}
