package server_rest_unit_alfa

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/ipoluianov/gazer_node/iunit"
	"github.com/ipoluianov/gazer_node/system/units/units_common"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gazer_node/utilities/uom"
)

type UnitServerRestUnitAlfa struct {
	units_common.Unit
	port              int
	unitId            string
	receivedVariables map[string]string

	srv *http.Server
}

func New() iunit.IUnit {
	var c UnitServerRestUnitAlfa
	c.receivedVariables = make(map[string]string)
	return &c
}

const (
	ItemNameStatus = "Status"
)

//go:embed "image.png"
var Image []byte

func Info() units_common.UnitMeta {
	var info units_common.UnitMeta
	info.TypeName = "Server.Rest.Unit.Alfa"
	info.Category = "server"
	info.DisplayName = "Server REST Alfa"
	info.Constructor = New
	info.ImgBytes = Image
	info.Description = ""
	return info
}

func (c *UnitServerRestUnitAlfa) GetConfigMeta() string {
	meta := units_common.NewUnitConfigItem("", "", "", "", "", "", "")
	meta.Add("port", "Port", "8880", "num", "", "", "")
	meta.Add("unit_id", "Unit ID", "", "string", "", "", "")
	return meta.Marshal()
}

func (c *UnitServerRestUnitAlfa) InternalUnitStart() error {
	var err error

	type Config struct {
		Port   float64 `json:"port"`
		UnitId string  `json:"unit_id"`
	}

	var config Config
	err = json.Unmarshal([]byte(c.GetConfig()), &config)
	if err != nil {
		err = errors.New("config error")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.port = int(math.Round(config.Port))
	if c.port < 80 {
		err = errors.New("wrong port")
		c.SetString(ItemNameStatus, err.Error(), "error")
		return err
	}

	c.unitId = config.UnitId

	c.receivedVariables = make(map[string]string)

	c.SetMainItem(ItemNameStatus)

	c.SetString(ItemNameStatus, "", "")
	go c.Tick()
	return nil
}

func (c *UnitServerRestUnitAlfa) InternalUnitStop() {
}

func (c *UnitServerRestUnitAlfa) ThServer() {
	err := c.srv.ListenAndServe()
	if err != nil {
		c.SetString(ItemNameStatus, err.Error(), uom.ERROR)
	} else {
		c.SetString(ItemNameStatus, "stopped", uom.NONE)
	}
}

func (c *UnitServerRestUnitAlfa) Tick() {
	var err error
	c.Started = true
	dtLastTime := time.Now().UTC()

	c.srv = &http.Server{
		Addr: ":" + fmt.Sprint(c.port),
	}
	c.srv.Handler = c
	go c.ThServer()

	for !c.Stopping {
		for {
			if c.Stopping || time.Now().Sub(dtLastTime) > time.Duration(1000)*time.Millisecond {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if c.Stopping {
			c.SetString(ItemNameStatus, "stopped", "")
			break
		}
		dtLastTime = time.Now().UTC()
	}

	for vName, _ := range c.receivedVariables {
		c.SetString(vName, "", "stopped")
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = c.srv.Shutdown(ctx); err != nil {
			logger.Println(err)
		}
	}

	c.SetString(ItemNameStatus, "", "stopped")
	c.Started = false
}

func (c *UnitServerRestUnitAlfa) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	state, err := c.Node().GetUnitState(c.unitId)
	if err == nil {
		bs, err := json.MarshalIndent(state, "", " ")
		if err == nil {
			w.Write(bs)
		}
	}
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}
