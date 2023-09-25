package system

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"time"

	"encoding/pem"

	"github.com/ipoluianov/gazer_node/common_interfaces"
	"github.com/ipoluianov/gazer_node/system/history"
	"github.com/ipoluianov/gazer_node/system/resources"
	"github.com/ipoluianov/gazer_node/system/settings"
	"github.com/ipoluianov/gazer_node/system/units/units_system"
	"github.com/ipoluianov/gazer_node/utilities"
	"github.com/ipoluianov/xchg/xchg"
)

type System struct {
	nodeName string
	ss       *settings.Settings

	currentMasterKey string
	currentGuestKey  string

	items       []*common_interfaces.Item
	itemsByName map[string]*common_interfaces.Item
	itemsById   map[uint64]*common_interfaces.Item
	nextItemId  uint64

	requester common_interfaces.Requester

	unitsSystem *units_system.UnitsSystem

	//unitsChannels []chan *common_interfaces.UnitMessage

	//cloudConnection *cloud.Connection
	xchgPoint *XchgServer

	history   *history.History
	resources *resources.Resources

	users      []*common_interfaces.User
	userByName map[string]*common_interfaces.User
	sessions   map[string]*UserSession

	itemWatchers map[string]*ItemWatcher

	apiCallsCount int

	stopping bool
	stopped  bool

	maintenanceLastValuesDT time.Time

	mtxSystem sync.Mutex
}

func RSAPrivateKeyFromHex(privateKey64 string) (privateKey *rsa.PrivateKey, err error) {
	var privateKeyBS []byte
	privateKeyBS, err = hex.DecodeString(privateKey64)
	if err != nil {
		return
	}
	privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBS)
	return
}

func RSAPrivateKeyToHex(privateKey *rsa.PrivateKey) (privateKey64 string) {
	if privateKey == nil {
		return
	}
	privateKeyBS := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKey64 = hex.EncodeToString(privateKeyBS)
	return
}

func NewSystem(ss *settings.Settings) *System {
	var c System
	c.ss = ss
	c.itemWatchers = make(map[string]*ItemWatcher)
	c.items = make([]*common_interfaces.Item, 0)
	c.itemsByName = make(map[string]*common_interfaces.Item)
	c.itemsById = make(map[uint64]*common_interfaces.Item)

	// Master Key
	masterKey, err := os.ReadFile(ss.ServerDataPath() + "/masterkey.txt")
	if err != nil {
		masterKey = utilities.GenerateRandomBytesWithSHA(30)
		masterKey = []byte(base32.StdEncoding.EncodeToString(masterKey))
		os.WriteFile(ss.ServerDataPath()+"/masterkey.txt", masterKey, 0666)
	}
	c.currentMasterKey = string(masterKey)

	// Guest Key
	guestKey, err := os.ReadFile(ss.ServerDataPath() + "/guestkey.txt")
	if err != nil {
		guestKey = utilities.GenerateRandomBytesWithSHA(20)
		guestKey = []byte(base32.StdEncoding.EncodeToString(guestKey))
		os.WriteFile(ss.ServerDataPath()+"/guestkey.txt", guestKey, 0666)
	}
	c.currentGuestKey = string(guestKey)

	// Private Key
	privateKeyPEMBS, err := os.ReadFile(ss.ServerDataPath() + "/private_key.pem")
	privateKeyPEM := string(privateKeyPEMBS)
	if err != nil {
		privateKey, _ := xchg.GenerateRSAKey()
		privateKeyPEM = RSAPrivateKeyToPem(privateKey)
		os.WriteFile(ss.ServerDataPath()+"/private_key.pem", []byte(privateKeyPEM), 0666)
	}

	// Address
	privateKey, _ := RSAPrivateKeyFromPem(privateKeyPEM)
	publicKeyBS := xchg.RSAPublicKeyToDer(&privateKey.PublicKey)
	address := xchg.AddressForPublicKeyBS(publicKeyBS)
	os.WriteFile(ss.ServerDataPath()+"/address.txt", []byte(address), 0666)

	c.xchgPoint = NewXchgServer(privateKey, c.currentMasterKey, c.currentGuestKey)

	c.unitsSystem = units_system.New(&c)
	go c.processUnitMessages(c.unitsSystem.OutputChannel())

	c.history = history.NewHistory(c.ss)
	c.resources = resources.NewResources(c.ss)

	c.users = make([]*common_interfaces.User, 0)
	c.userByName = make(map[string]*common_interfaces.User)
	c.sessions = make(map[string]*UserSession)

	return &c
}

func RSAPrivateKeyToPem(privateKey *rsa.PrivateKey) (privateKeyPem string) {
	if privateKey == nil {
		return
	}
	privateKeyBS := x509.MarshalPKCS1PrivateKey(privateKey)
	block := pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyBS,
	}
	privateKeyPem = string(pem.EncodeToMemory(&block))
	return
}

func RSAPrivateKeyFromPem(privateKeyPem string) (privateKey *rsa.PrivateKey, err error) {
	block, _ := pem.Decode([]byte(privateKeyPem))
	if block == nil {
		return nil, errors.New("no pem data")
	}
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	return
}

//var pKeyHex = "308204a60201000282010100da54fe48034d64508d14a2a917a0a22a5a127be200f009f06b8160437cb9f6a3db7ea1d83714df9bd18da8a93cf948b7554283c6aac80707d22e9cfa657e240a0775f18366230104b7a698605f194b04cb74e735b35c81e7840ecb2cde2a7bcaa6f95a2fdfdba19b621633e8c5232002d050b8bb2aedc81e390c1978afec6db35dfcc9dd557c70f6451822b9d5046d149f06e460cdb2d4b1693fe3a44bbc3997473851d37ac72142774f0bc71f2d473c82978a79e2e04acc81d0ce2e97927c0f72d7c5187f2eecda965e24788860575013d26de0b0bb8a85543fdf948502d80b63a66ee3040650308b9ce2b2e53f674123e1683012f61f2dddfb83017c291609020301000102820101008691061cca7443d4e5ef8705e33b2d581f25ef56efaf02e13cd183cc74ae85969ead61014b89c0fc5fdf08ca2e7b92d6f464c7a586133d4a13c0891e47b9c46aae0882afb31ef5fcbb58a1e81c1511c8c5c1aad3cd55c1f738cc896b810efc377e4c87caa415d1785caf44953e141521c6d549d68a71fdfaaabc8b627843a7a45766488cadeb04fdb2014adabefa9cf2a403fc4c40aa4c0eed966cd96ce8b65cdae89ea6d9b32c3a800031e620e00a1a124f7b836b1f2fa86bbf6c3fd78161738470a026c91350dafb92d269bbba34486f53510b060617a0c6766a5f6dac50c15330baaf4522c7fb3c193d8fa054be15d00d4ff3aaf2de01140ff5e90d93782502818100ec41e603fc0797dc5ff233bdf544f3d687df074b2782ff8cae4c1c0168e98cfd908a2e6f25a419d145f9bec292f87bb14c94ce50421a08b4152e2d86d2c92b7c5eface0f7ebeacfc6d1bee6dc1a2fc4c3751c34ca9643ab6acc94e1c2b0ed98a80b1504d50f9fc96d7e4450625ab290f6c41f71885847c191b6212776b2ab69302818100ec93a0cbca41de2f91f72cb4f1740ad9ae97837132abff3161f4a1f59b8de53b16ec090b019fd845c23f541d2d73221755eaeb3f0a08b4510629560880c7f70fa09c4f5e2c4d204f73a55553f50cc469f2172a85af9b6b51f70d1a6437534abac9d59a4c1b576f2cff2b3928c26a819fc8a6c98423337bb2e2da126c85ebe67302818100d986af3699f58fd01b1310aae4c9d0cc849b47c6dda152937fe399a17eac12e16014aa3e31e50ed44d5e6e520e29f51140967e030d6994fbe1c87ba878293afdaf21b35a36f36ea897f930a523b56220f68a348c4026859cae2846dfce9144a0ae6f13a5ac5a0f43ff91303041bc7ef8b14f6727cfbf34d7950bce3edf901b3b02818100814a0cce52b2bca272bb0a8bb8891a84ec8d912003f94b75c97ad02851e2b22c20d2cdfe5ddce56cfa4371cca05213877d44ed5b7e3853931432f2f9a2a7a5b5bca8b0175f4ea721c4a9ce801ba3e6939fe25932c64dc1d1019aff99554307cc1d11c7496087e0124f4167f3868c7e5abc65aa2bb4b126211528e878b697bd5102818100892e5597d206a0a9bae1fb39956dde9af0cb9581168844f76f7adf3fe04c5573d2ac2a4c9a208a9a2f3cc44b3d004c17ed51a01eb601417d7d647de06d0ac9dc2312fb8b46ea88acf17c410010215b97c1439db639a8db2d10978bf724c719e19b8f0ff7edcaf0f1154bf0f6b5cb17d3e98abb9ecc2a74f5cc8cf6334db00ba9"
//var addr = "#whucrl4odswdr7fccok3speepprdgdkym2pk45xep2dkhpqj"

func (c *System) Settings() *settings.Settings {
	return c.ss
}

func (c *System) SetRequester(requester common_interfaces.Requester) {
	c.requester = requester
	c.xchgPoint.SetRequester(c.requester)
}

func (c *System) Start() {
	c.stopping = false
	c.stopped = false

	c.LoadConfig()

	items := c.ReadLastValues()
	for _, item := range items {
		if realItem, ok := c.itemsByName[item.Name]; ok {
			realItem.Value = item.Value
		}
	}
	c.xchgPoint.Start()
	c.history.Start()
	c.unitsSystem.Start()

	go c.thMaintenance()
}

func (c *System) Stop() {
	c.stopping = true
	c.unitsSystem.Stop()
	c.history.Stop()
	c.xchgPoint.Stop()
	c.SaveConfig()
	//c.saveSessions()

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if c.stopped {
			break
		}
	}

	c.WriteLastValues(c.items)
}

func (c *System) processUnitMessages(unitChannel chan common_interfaces.UnitMessage) {
	for msg := range unitChannel {
		switch v := msg.(type) {
		case *common_interfaces.UnitMessageItemValue:
			msgItemValue := v
			//msgItemValue := msg.(*common_interfaces.UnitMessageItemValue)
			c.SetItemByNameOld(msgItemValue.ItemName, msgItemValue.Value, msgItemValue.UOM, time.Now(), false)
		case *common_interfaces.UnitMessageSetProperty:
			msgSetProperty := v
			//msgSetProperty := msg.(*common_interfaces.UnitMessageSetProperty)
			c.SetPropertyIfDoesntExist(msgSetProperty.ItemName, msgSetProperty.PropName, msgSetProperty.PropValue)
		case *common_interfaces.UnitMessageItemTouch:
			msgItemTouch := msg.(*common_interfaces.UnitMessageItemTouch)
			c.TouchItem(msgItemTouch.ItemName)
		case *common_interfaces.UnitMessageRemoteItemsOfUnit:
			msgRemoteItemsOfUnit := msg.(*common_interfaces.UnitMessageRemoteItemsOfUnit)
			c.RemoveItemsOfUnit(msgRemoteItemsOfUnit.UnitId)
		case *common_interfaces.UnitMessageSetAllItemsByUnitName:
			msgSetAllItemsByUnitName := msg.(*common_interfaces.UnitMessageSetAllItemsByUnitName)
			c.SetAllItemsByUnitName(msgSetAllItemsByUnitName.UnitId, msgSetAllItemsByUnitName.Value, msgSetAllItemsByUnitName.UOM, time.Now(), false)
		}
	}
}

func (c *System) RegApiCall() {
	c.mtxSystem.Lock()
	c.apiCallsCount++
	c.mtxSystem.Unlock()
}

func (c *System) thMaintenance() {
	for !c.stopping {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			if c.stopping {
				break
			}
		}
		if c.stopping {
			break
		}

		c.maintenanceLastValues()
	}
	c.stopped = true
}

func (c *System) maintenanceLastValues() {
	if time.Since(c.maintenanceLastValuesDT) > 10*time.Second {
		c.maintenanceLastValuesDT = time.Now()
		c.WriteLastValues(c.items)
		c.RemoveOldLastValuesFiles()
	}
}
