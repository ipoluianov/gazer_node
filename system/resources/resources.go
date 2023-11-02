package resources

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/protocols/nodeinterface"
	"github.com/ipoluianov/gazer_node/system/settings"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

type Resources struct {
	mtx sync.Mutex
	ss  *settings.Settings
}

func NewResources(ss *settings.Settings) *Resources {
	var c Resources
	c.ss = ss
	return &c
}

func (c *Resources) dir() string {
	return c.ss.ServerDataPath() + "/res"
}

func (c *Resources) fileName(name string) string {
	return base64.StdEncoding.EncodeToString([]byte(name))
}

func (c *Resources) Rename(id string, props []nodeinterface.PropItem) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	var err error

	dir := c.dir()

	var bs []byte

	bs, err = os.ReadFile(dir + "/" + id + ".info")
	if err != nil {
		return errors.New("ERR_NO_RESOURCE")
	}

	var info common_interfaces.ResourcesItemInfo
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return errors.New("ERR_NO_RESOURCE")
	}

	if info.Properties == nil {
		info.Properties = make([]*common_interfaces.ItemProperty, 0)
	}

	for _, prop := range props {
		found := false
		for existingPropIndex, existingProp := range info.Properties {
			if prop.PropName == existingProp.Name {
				info.Properties[existingPropIndex].Value = prop.PropValue
				found = true
				break
			}
		}
		if !found {
			info.Properties = append(info.Properties, &common_interfaces.ItemProperty{
				Name:  prop.PropName,
				Value: prop.PropValue,
			})
		}
	}

	bs, _ = json.MarshalIndent(info, "", " ")
	err = os.WriteFile(dir+"/"+id+".info", bs, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (c *Resources) Remove(id string) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	logger.Println("Resource Removing", id)

	var err error

	dir := c.dir()

	_, err = os.ReadFile(dir + "/" + id + ".info")
	if err != nil {
		return errors.New("ERR_NO_RESOURCE")
	}

	err = os.Remove(dir + "/" + id + ".info")
	if err != nil {
		return err
	}
	err = os.Remove(dir + "/" + id + ".content")
	if err != nil {
		return err
	}

	return nil
}

func (c *Resources) Add(name string, tp string, content []byte) (string, error) {
	var err error
	c.mtx.Lock()
	defer c.mtx.Unlock()

	dir := c.dir()
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	filePathInfo := dir + "/" + "_"
	filePathContent := dir + "/" + "_"

	foundId := false
	var id string
	for i := 0; i < 100; i++ {
		id = uuid.New().String()
		filePathInfo = dir + "/" + id + ".info"
		filePathContent = dir + "/" + id + ".content"
		if _, err = os.Stat(filePathInfo); os.IsNotExist(err) {
			foundId = true
			break
		}
	}

	if !foundId {
		return "", errors.New("no id found")
	}

	var info common_interfaces.ResourcesItemInfo
	info.Id = id
	info.Type = tp
	info.Properties = make([]*common_interfaces.ItemProperty, 0)
	info.Properties = append(info.Properties, &common_interfaces.ItemProperty{
		Name:  "name",
		Value: name,
	})
	bs, _ := json.MarshalIndent(info, "", " ")
	err = os.WriteFile(filePathInfo, bs, 0666)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filePathContent, content, 0666)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (c *Resources) writeFile(name string, offset int64, data []byte) error {
	var file *os.File
	var err error
	logger.Println("Resources - writeFile", "name = "+name+" offset =", offset)

	if offset == 0 {
		file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(name, os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		_, err = file.Seek(offset, 0)
		if err != nil {
			return err
		}
	}

	_, err = file.Write(data)
	if err1 := file.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func (c *Resources) Set(id string, suffix string, offset int64, content []byte) error {
	if suffix != "" && suffix != "thumbnail" {
		return errors.New("wrong suffix")
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	var err error
	var bs []byte

	dir := c.dir()

	bs, err = os.ReadFile(dir + "/" + id + ".info")
	if err != nil {
		return errors.New("ERR_NO_RESOURCE")
	}

	var info common_interfaces.ResourcesItemInfo
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return errors.New("ERR_NO_RESOURCE")
	}

	if info.Properties == nil {
		info.Properties = make([]*common_interfaces.ItemProperty, 0)
	}

	bs, _ = json.MarshalIndent(info, "", " ")
	err = os.WriteFile(dir+"/"+id+".info", bs, 0666)
	if err != nil {
		return err
	}

	if len(suffix) > 0 {
		suffix = "." + suffix
	}

	//err = ioutil.WriteFile(dir+"/"+id+".content", content, 0666)
	err = c.writeFile(dir+"/"+id+suffix+".content", offset, content)
	if err != nil {
		return errors.New("can not save resource")
	}

	/*err = ioutil.WriteFile(dir+"/"+id+".thumbnail.png", thumbnail, 0666)
	if err != nil {
		return errors.New("can not save resource")
	}*/

	return nil
}

func (c *Resources) Get(id string, offset int64, size int64) (nodeinterface.ResourceGetResponse, error) {
	if offset < 0 || size < 1 {
		return nodeinterface.ResourceGetResponse{}, errors.New("wrong offset/size (<0/1)")
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	var err error
	var bs []byte

	dir := c.dir()

	bs, err = os.ReadFile(dir + "/" + id + ".info")
	if err != nil {
		return nodeinterface.ResourceGetResponse{}, errors.New("ERR_NO_RESOURCE")
	}

	var info common_interfaces.ResourcesItemInfo
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return nodeinterface.ResourceGetResponse{}, errors.New("ERR_NO_RESOURCE")
	}

	bs, err = os.ReadFile(dir + "/" + id + ".content")
	if err != nil {
		return nodeinterface.ResourceGetResponse{}, errors.New("ERR_NO_RESOURCE")
	}

	result := nodeinterface.ResourceGetResponse{}
	result.Id = info.Id
	result.Type = info.Type
	result.Offset = offset
	result.Size = int64(len(bs))

	if offset < int64(len(bs)) {
		if offset+size > int64(len(bs)) {
			size = int64(len(bs)) - offset
		}
		result.Content = bs[offset : offset+size]
	} else {
		result.Content = make([]byte, 0)
	}

	return result, nil
}

func (c *Resources) GetAllInfos() []common_interfaces.ResourcesItemInfo {
	var err error
	var infos []common_interfaces.ResourcesItemInfo
	var files []os.FileInfo
	files, err = ioutil.ReadDir(c.dir())
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".info") {
			var bs []byte
			bs, err = os.ReadFile(c.dir() + "/" + file.Name())
			if err == nil {
				var info common_interfaces.ResourcesItemInfo
				err = json.Unmarshal(bs, &info)
				if err == nil {
					infos = append(infos, info)
				}
			}
		}
	}
	return infos
}

func (c *Resources) GetIdByPath(path string) (id string, err error) {
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/'
	})
	infos := c.GetAllInfos()
	searchingIndex := 0
	searchingParentFolderId := ""

	if len(parts) == 0 {
		return "", errors.New("ERR_NO_RESOURCE")
	}

	for {
		found := false
		for _, info := range infos {
			if info.GetProp("name") == parts[searchingIndex] && info.GetProp("folder") == searchingParentFolderId {
				searchingIndex++
				searchingParentFolderId = info.Id
				found = true
				break
			}
		}
		if !found {
			return "", errors.New("ERR_NO_RESOURCE")
		}
		if searchingIndex == len(parts) {
			return searchingParentFolderId, nil
		}
	}
}

func (c *Resources) GetByPath(path string, offset int64, size int64) (nodeinterface.ResourceGetResponse, error) {
	id, err := c.GetIdByPath(path)
	if err != nil {
		return nodeinterface.ResourceGetResponse{}, errors.New("ERR_NO_RESOURCE")
	}
	return c.Get(id, offset, size)
}

func (c *Resources) SetByPath(path string, tp string, content []byte) (string, error) {
	id, err := c.GetIdByPath(path)
	if err != nil {
		return c.Add(path, tp, content)
	}
	return id, c.Set(id, "", 0, content)
}

func SplitWithoutEmpty(req string, sep rune) []string {
	return strings.FieldsFunc(req, func(r rune) bool {
		return r == sep
	})
}

func (c *Resources) List(tp string, filter string, offset int, maxCount int) common_interfaces.ResourcesInfo {
	var result common_interfaces.ResourcesInfo
	result.Items = make([]common_interfaces.ResourcesItemInfo, 0)
	c.mtx.Lock()
	defer c.mtx.Unlock()

	filterParts := SplitWithoutEmpty(strings.ToLower(filter), ' ')

	//allItems := make([]common_interfaces.ResourcesItemInfo, 0)
	var err error
	var files []os.FileInfo
	files, err = ioutil.ReadDir(c.dir())
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".info") {
			var bs []byte
			bs, err = os.ReadFile(c.dir() + "/" + file.Name())
			if err == nil {
				var info common_interfaces.ResourcesItemInfo
				err = json.Unmarshal(bs, &info)
				if err == nil {
					result.TotalCount++

					inFilter := 0
					/*for _, filterPart := range filterParts {
						if strings.Contains(strings.ToLower(info.Name), filterPart) {
							inFilter++
						}
					}*/
					if inFilter == len(filterParts) && (tp == "" || tp == info.Type) {
						if result.InFilterCount >= offset && len(result.Items) < maxCount {
							result.Items = append(result.Items, info)
						}
						result.InFilterCount++
					}
				}
			}
		}
	}

	return result
}

func (c *Resources) PropSet(resourceId string, props []nodeinterface.PropItem) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	var err error

	dir := c.dir()

	var bs []byte

	bs, err = ioutil.ReadFile(dir + "/" + resourceId + ".info")
	if err != nil {
		return errors.New("ERR_NO_RESOURCE")
	}

	var info common_interfaces.ResourcesItemInfo
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return errors.New("ERR_NO_RESOURCE")
	}

	if info.Properties == nil {
		info.Properties = make([]*common_interfaces.ItemProperty, 0)
	}

	for _, prop := range props {
		found := false
		for existingPropIndex, existingProp := range info.Properties {
			if prop.PropName == existingProp.Name {
				info.Properties[existingPropIndex].Value = prop.PropValue
				found = true
				break
			}
		}
		if !found {
			info.Properties = append(info.Properties, &common_interfaces.ItemProperty{
				Name:  prop.PropName,
				Value: prop.PropValue,
			})
		}
	}

	bs, _ = json.MarshalIndent(info, "", " ")
	err = os.WriteFile(dir+"/"+resourceId+".info", bs, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (c *Resources) PropGet(resourceId string) ([]nodeinterface.PropItem, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	var err error
	var bs []byte

	dir := c.dir()

	bs, err = os.ReadFile(dir + "/" + resourceId + ".info")
	if err != nil {
		return make([]nodeinterface.PropItem, 0), errors.New("ERR_NO_RESOURCE")
	}

	var info common_interfaces.ResourcesItemInfo
	err = json.Unmarshal(bs, &info)
	if err != nil {
		return make([]nodeinterface.PropItem, 0), errors.New("ERR_NO_RESOURCE")
	}

	if info.Properties == nil {
		info.Properties = make([]*common_interfaces.ItemProperty, 0)
	}

	result := make([]nodeinterface.PropItem, 0)

	for _, value := range info.Properties {
		result = append(result, nodeinterface.PropItem{
			PropName:  value.Name,
			PropValue: value.Value,
		})
	}

	return result, nil
}
