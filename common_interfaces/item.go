package common_interfaces

import (
	"strconv"
	"strings"
)

type ItemValue struct {
	Value string `json:"v"`
	DT    int64  `json:"t"`
	UOM   string `json:"u"`
}

type Item struct {
	Id         uint64    `json:"id"`
	Name       string    `json:"name"`
	Value      ItemValue `json:"value"`
	Properties map[string]*ItemProperty

	postprocessingTrim      bool
	postprocessingAdjust    bool
	postprocessingScale     float64
	postprocessingOffset    float64
	postprocessingPrecision int
}

type ItemGetUnitItems struct {
	Item
}

type ItemStateInfo struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Value       string `json:"v"`
	DT          int64  `json:"t"`
	UOM         string `json:"u"`
}

func NewItem() *Item {
	var c Item
	c.Properties = make(map[string]*ItemProperty)
	return &c
}

func (c *Item) SetPostprocessingTrim(postprocessingTrim bool) {
	c.postprocessingTrim = postprocessingTrim
}

func (c *Item) SetPostprocessingAdjust(postprocessingAdjust bool) {
	c.postprocessingAdjust = postprocessingAdjust
}

func (c *Item) SetPostprocessingScale(postprocessingScale float64) {
	c.postprocessingScale = postprocessingScale
}

func (c *Item) SetPostprocessingOffset(postprocessingOffset float64) {
	c.postprocessingOffset = postprocessingOffset
}

func (c *Item) SetPostprocessingPrecision(postprocessingPrecision int) {
	c.postprocessingPrecision = postprocessingPrecision
}

func (c *Item) PostprocessingValue(value string) string {
	if c.postprocessingTrim {
		value = strings.Trim(value, " \r\n\t")
	}

	if c.postprocessingAdjust {
		var err error
		var valueFloat float64
		valueFloat, err = strconv.ParseFloat(value, 64)
		if err == nil {
			valueFloat = valueFloat*c.postprocessingScale + c.postprocessingOffset
			value = strconv.FormatFloat(valueFloat, 'f', c.postprocessingPrecision, 64)
			if strings.Contains(value, ".") {
				value = strings.TrimRight(value, "0")
			}
		}
	}
	return value
}

func (c *Item) SetPropertyIfDoesntExist(propName string, propValue string) {
	if _, ok := c.Properties[propName]; !ok {
		c.Properties[propName] = &ItemProperty{
			Name:  propName,
			Value: propValue,
		}
	}
}

func (c *Item) SetProperty(propName string, propValue string) {
	c.Properties[propName] = &ItemProperty{
		Name:  propName,
		Value: propValue,
	}
}
