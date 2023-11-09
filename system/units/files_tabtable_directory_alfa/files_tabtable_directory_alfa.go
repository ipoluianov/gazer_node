package files_tabtable_directory_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type UnitTxtTableFolder struct {
	units_common.Unit
	directory string
	periodMs  int
}

func New() iunit.IUnit {
	var c UnitTxtTableFolder
	return &c
}

const (
	ItemNameResult = "result"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Files.TabTable.Directory.Alfa"
	info.Category = "file"
	info.DisplayName = "File Text Table Folder"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitTxtTableFolder) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("directory", "DIrectory", "/", "string", "", "", "")
	meta.Add("period", "Period, ms", "1000", "num", "0", "999999", "0")
	return meta.Marshal()
}

func (c *UnitTxtTableFolder) InternalUnitStart() error {
	var err error
	c.SetString(ItemNameResult, "", "")
	c.SetMainItem(ItemNameResult)

	type Config struct {
		Directory string  `json:"directory"`
		Period    float64 `json:"period"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameResult, err.Error(), "error")
		return err
	}

	c.directory = config.Directory
	if c.directory == "" {
		err = errors.New("wrong file")
		c.SetString(ItemNameResult, err.Error(), "error")
		return err
	}

	c.periodMs = int(config.Period)
	if c.periodMs < 100 {
		err = errors.New("wrong period")
		c.SetString(ItemNameResult, err.Error(), "error")
		return err
	}

	go c.Tick()
	return nil
}

func (c *UnitTxtTableFolder) InternalUnitStop() {
}

func ReadLastLine(fileName string) (names []string, values []string, err error) {
	// Get file information
	fileInfo, err := os.Lstat(fileName)
	if err != nil {
		return
	}

	// Open the file
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	// Determine block size
	maxBufferSize := int64(4096)
	offset := fileInfo.Size() - maxBufferSize
	if offset < 0 {
		offset = 0
		maxBufferSize = fileInfo.Size()
	}

	// Navigate to the begin of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		return
	}

	// Read the first line
	bufferFirstLine := make([]byte, maxBufferSize)
	_, err = file.Read(bufferFirstLine)
	if err != nil {
		return
	}

	for i := 0; i < len(bufferFirstLine); i++ {
		if bufferFirstLine[i] == 10 || bufferFirstLine[i] == 13 {
			names = strings.FieldsFunc(string(bufferFirstLine[:i]), func(r rune) bool {
				return r == '\t' || r == ' ' || r == ',' || r == ';' || r == '\r' || r == '\n'
			})
			break
		}
	}

	// Navigate to the last bytes of the file
	_, err = file.Seek(offset, 0)
	if err != nil {
		return
	}

	// Read the last block
	buffer := make([]byte, maxBufferSize)
	_, err = file.Read(buffer)
	if err != nil {
		return
	}

	// Find CR or LN
	foundLine := buffer
	foundChar := false
	lineIsFound := false
	for i := len(buffer) - 1; i > 0; i-- {
		if (buffer[i] == 10 || buffer[i] == 13) && foundChar {
			foundLine = buffer[i:]
			lineIsFound = true
			break
		}
		if buffer[i] > 32 {
			foundChar = true
		}
	}

	if !lineIsFound && offset == 0 {
		values = nil // it is header line
	}

	values = strings.FieldsFunc(string(foundLine), func(r rune) bool {
		return r == '\t' || r == ' ' || r == ',' || r == ';' || r == '\r' || r == '\n'
	})

	return
}

func FindLastFileInDirectory(directory string) (fileName string, err error) {
	items, err := os.ReadDir(directory)
	if err != nil {
		return
	}
	if len(items) < 1 {
		return
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name() < items[j].Name()
	})
	fileName = items[len(items)-1].Name()
	return
}

func (c *UnitTxtTableFolder) Tick() {
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

		fileName, err := FindLastFileInDirectory(c.directory)
		if err != nil {
			c.SetString(ItemNameResult, err.Error(), "error")
			//c.SetError(err.Error())
			continue
		}

		fileName = c.directory + "/" + fileName
		c.SetString("file", fileName, "")

		names, values, err := ReadLastLine(fileName)

		if err != nil {
			c.SetString(ItemNameResult, err.Error(), "error")
			//c.SetError(err.Error())
		} else {
			if len(names) == len(values) {
				for i := 0; i < len(names); i++ {
					c.SetString(names[i], values[i], "")
				}
				//c.SetError("")
				c.SetString(ItemNameResult, "ok", "")
			} else {
				err = errors.New("header fields doesn't match other rows")
				c.SetString(ItemNameResult, err.Error(), "error")
				//c.SetError(err.Error())
			}
		}

	}
	c.SetString(ItemNameResult, "", "stopped")
	c.Started = false
}
