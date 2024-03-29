package system

import (
	"crypto"
	"encoding/base64"
	"errors"
	"sort"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/protocols/nodeinterface"
	"github.com/ipoluianov/gazer_node/utilities/logger"
)

var DefaultUserName string

//var DefaultUserPassword string

func init() {
	DefaultUserName = "admin"
	//DefaultUserPassword = "admin"
}

type UserSession struct {
	SessionToken              string `json:"session_token"`
	UserName                  string `json:"user_name"`
	SessionOpenTime           int64  `json:"session_open_time"`
	SessionOpenTimeForDisplay string `json:"session_open_time_for_display"`
	Host                      string `json:"host"`
}

/*func (c *System) CheckSession(sessionToken string) (string, error) {
	var userName string
	var err error
	c.mtx.Lock()
	if session, ok := c.sessions[sessionToken]; ok {
		userName = session.UserName
	} else {
		err = errors.New("wrong session token")
	}
	c.mtx.Unlock()
	return userName, err
}*/

/*func (c *System) RemoveSession(sessionToken string) error {
	var err error
	c.mtxSystem.Lock()
	if _, ok := c.sessions[sessionToken]; ok {
		delete(c.sessions, sessionToken)
	} else {
		err = errors.New("wrong session token")
	}

	bs, err := json.MarshalIndent(c.sessions, "", " ")
	if err == nil {
		err = ioutil.WriteFile(c.ss.ServerDataPath()+"/sessions.json", bs, 0666)
		if err != nil {
			logger.Println("saveSessions error", err)
		}
	} else {
		logger.Println("saveSessions (marshal) error", err)
	}

	c.mtxSystem.Unlock()

	return err
}*/

func (c *System) SessionList(userName string) (nodeinterface.SessionListResponse, error) {
	var result nodeinterface.SessionListResponse
	var err error
	c.mtxSystem.Lock()
	for _, s := range c.sessions {
		if s.UserName == userName {
			var item nodeinterface.SessionListResponseItem
			item.SessionToken = s.SessionToken
			item.UserName = s.UserName
			item.SessionOpenTime = s.SessionOpenTime
			result.Items = append(result.Items, item)
		}
	}
	c.mtxSystem.Unlock()

	sort.Slice(result.Items, func(i, j int) bool {
		return result.Items[i].SessionOpenTime < result.Items[j].SessionOpenTime
	})

	return result, err
}

/*func (c *System) OpenSession(name string, password string, host string) (nodeinterface.SessionOpenResponse, error) {
	var result nodeinterface.SessionOpenResponse
	var err error

	c.mtx.Lock()
	if user, ok := c.userByName[name]; ok {
		if c.hashPassword(password) == user.PasswordHash {
			stringForHash := time.Now().Format("2006-01-02-15-04-05") + strconv.FormatInt(rand.Int63(), 10) + "42"
			sessionToken := c.hashSession(stringForHash)
			result.SessionToken = sessionToken

			timeOpenSession := time.Now().UTC()

			var ss UserSession
			ss.UserName = name
			ss.SessionToken = result.SessionToken
			ss.SessionOpenTime = timeOpenSession.UnixNano() / 1000
			ss.SessionOpenTimeForDisplay = timeOpenSession.Format("2006-01-02 15:04:05.999")
			ss.Host = host
			c.sessions[result.SessionToken] = &ss
		} else {
			err = errors.New("wrong password")
		}
	} else {
		err = errors.New("user not found")
	}

	c.saveSessions()

	c.mtx.Unlock()
	return result, err
}*/

func (c *System) UserList() (nodeinterface.UserListResponse, error) {
	var result nodeinterface.UserListResponse
	c.mtxSystem.Lock()
	result.Items = make([]string, 0)
	for _, u := range c.users {
		result.Items = append(result.Items, u.Name)
	}
	c.mtxSystem.Unlock()
	return result, nil
}

func (c *System) UserAdd(name string, password string) (nodeinterface.UserAddResponse, error) {
	var err error
	var result nodeinterface.UserAddResponse
	c.mtxSystem.Lock()
	if _, ok := c.userByName[name]; !ok {
		var us common_interfaces.User
		us.Name = name
		us.PasswordHash = c.hashPassword(password)
		us.Properties = make(map[string]*common_interfaces.ItemProperty)
		c.users = append(c.users, &us)
		c.userByName[us.Name] = &us
	} else {
		err = errors.New("user exists already")
	}
	c.mtxSystem.Unlock()

	c.SaveConfig()

	return result, err
}

func (c *System) UserSetPassword(name string, password string) (nodeinterface.UserSetPasswordResponse, error) {
	var err error
	var result nodeinterface.UserSetPasswordResponse
	c.mtxSystem.Lock()
	if u, ok := c.userByName[name]; ok {
		u.PasswordHash = c.hashPassword(password)
	} else {
		err = errors.New("no user found")
	}
	c.mtxSystem.Unlock()

	c.SaveConfig()

	return result, err
}

func (c *System) hashPassword(password string) string {
	s := crypto.SHA256.New()
	return base64.StdEncoding.EncodeToString(s.Sum([]byte(password)))
}

/*func (c *System) hashSession(sessionData string) string {
	sum := sha256.Sum256([]byte(sessionData))
	hexStr := hex.EncodeToString(sum[0:10])
	return fmt.Sprint(hexStr)
}*/

func (c *System) UserRemove(name string) (nodeinterface.UserRemoveResponse, error) {
	var err error
	var found bool
	var result nodeinterface.UserRemoveResponse
	c.mtxSystem.Lock()
	for index, u := range c.users {
		if u.Name == name {
			c.users = append(c.users[:index], c.users[index+1:]...)
			delete(c.userByName, name)
			found = true
			break
		}
	}
	if !found {
		err = errors.New("no user found")
	}
	c.mtxSystem.Unlock()

	c.SaveConfig()

	return result, err
}

func (c *System) UserPropSet(userName string, props []nodeinterface.PropItem) error {
	logger.Println("UserPropSet", userName)
	c.mtxSystem.Lock()
	if user, ok := c.userByName[userName]; ok {
		for _, prop := range props {
			user.Properties[prop.PropName] = &common_interfaces.ItemProperty{
				Name:  prop.PropName,
				Value: prop.PropValue,
			}
		}
	} else {
		c.mtxSystem.Unlock()
		return errors.New("user not found")
	}
	c.mtxSystem.Unlock()
	c.SaveConfig()
	return nil
}

func (c *System) UserPropGet(userName string) ([]nodeinterface.PropItem, error) {
	result := make([]nodeinterface.PropItem, 0)

	c.mtxSystem.Lock()
	if user, ok := c.userByName[userName]; ok {
		for _, prop := range user.Properties {
			result = append(result, nodeinterface.PropItem{
				PropName:  prop.Name,
				PropValue: prop.Value,
			})
		}
	} else {
		c.mtxSystem.Unlock()
		return nil, errors.New("user not found")
	}
	c.mtxSystem.Unlock()
	return result, nil
}
