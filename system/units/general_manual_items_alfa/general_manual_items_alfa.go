package general_manual_items_alfa

import (
	_ "embed"
	"encoding/json"
	"errors"
	"time"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
)

type Item struct {
	Name      string `json:"item_name"`
	InitValue string `json:"init_value"`
}

type Config struct {
	Items []Item `json:"items"`
}

type UnitManual struct {
	units_common.Unit
	fileName string
	periodMs int
	config   Config
}

func New() common_interfaces.IUnit {
	var c UnitManual
	return &c
}

const (
	ItemNameStatus = "Status"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "General.Manual.Items.Alfa"
	info.Category = "general"
	info.DisplayName = "Manual Items"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitManual) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	t1 := meta.Add("items", "Items", "", "table", "", "", "")
	t1.Add("item_name", "Item Name", "item1", "string", "", "", "")
	t1.Add("init_value", "Init Value", "42", "string", "", "", "")
	return meta.Marshal()
}

func (c *UnitManual) InternalUnitStart() error {
	var err error
	c.SetString(ItemNameStatus, "", "starting")
	c.SetMainItem(ItemNameStatus)

	err = json.Unmarshal([]byte(c.GetConfig()), &c.config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	go c.Tick()

	c.SetString(ItemNameStatus, "", "started")
	return nil
}

func (c *UnitManual) InternalUnitStop() {
}

func (c *UnitManual) Tick() {
	c.Started = true

	for _, item := range c.config.Items {
		c.TouchItem(item.Name)
	}

	for !c.Stopping {
		for {
			if c.Stopping {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			break
		}
	}

	c.SetString(ItemNameStatus, "", "stopped")
	c.Started = false
}
