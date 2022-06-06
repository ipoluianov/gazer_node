package xchg

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gazercloud/gazernode/utilities/logger"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	mtx               sync.Mutex
	httpClientSend    *http.Client
	httpClientReceive *http.Client
	httpClientPing    *http.Client
	xchgIP            string
	stopping          bool
	IPsByAddress      map[string]string
	OnReceived        func([]byte) ([]byte, error)

	// Local keys
	privateKey    *rsa.PrivateKey
	privateKeyBS  []byte
	privateKey64  string
	privateKeyHex string

	publicKeyBS  []byte
	publicKey64  string
	publicKeyHex string

	// AES Key
	aesKey  []byte
	counter uint64
	lid     uint64
}

func NewClient(localAddr string, onRcv func([]byte) ([]byte, error)) *Client {
	var c Client
	c.OnReceived = onRcv

	c.httpClientSend = &http.Client{}
	c.httpClientSend.Timeout = 3000 * time.Millisecond

	c.httpClientReceive = &http.Client{}
	c.httpClientReceive.Timeout = 60000 * time.Millisecond

	c.httpClientPing = &http.Client{}
	c.httpClientPing.Timeout = 3000 * time.Millisecond

	c.IPsByAddress = make(map[string]string)
	c.xchgIP = ""

	c.generateKeys()

	go c.thRcv()
	return &c
}

func (c *Client) generateKeys() {
	var err error
	privateKeyBS, _ := base64.StdEncoding.DecodeString("MIIEowIBAAKCAQEAw3HnYPGjGltAf1vIw7U8/VrYrAtICk6gPy+K+q+YuQTjYJ8bdc7T5HcshkHpJ5gT9JR9fhC/JhFsRe1ZOV/CxLHYyD0ruo8ouyolC29CSHmeNqRp2TiV8sC642HoTphGRf0MQ0uaq7h7AYdVMxgUUKPgJs5eLI4KQnJa+Dwl0+HUUq54g2qQja4wAgrXhbtm+qm3hcJBycQbuBG2LfGl+lboA7cn0Vo+03QxQlXAp0MBuVOBIQ29PjR2hrq/T6+f48r4XzrUFfrV8iFrQtIq4R33j6UO/88jWcXXnlRAXt4/Eg65W+avBf83UIUVMMtn1QUcpBnyKis2qPF9o+bvCQIDAQABAoIBABfRouQyrrEAm/ypf+8yAEvULYHSIiZ3bJomviZNDizGRru4yEz0NuiqCXgXQkX8B7qP+jdJ7THDf9GJ2ozeecsk7YmBwvmKhulAeqFJHufcQobgRLIfbk7WZDBf90LU1gOjkkIFTcVNx1fpWV3PunIVdrTkA6Akc2WjsCh+lBGdRpB7wrW4KpzKQJNyEo1rSefeS9N55YP13l0WEArIPWIxwe1tJqFdA9pQ345Nt2OO/NUFEFoWpRb17LXeFtXsd/yAQa+1NdnPB7kz8j/G9yD6OkM0mx2tHCcRZDmEtrmstlcZ8mxWj0R6HFJupFEsgJ+tpwwOkXI4zRB6bj1VG30CgYEA4I/HXlgm7tMxVQZld3rZ9sxwZ6/nqbXx0p3A4IqPR+K0xn8oLWSjrzMYBbH1xqBu1Z0BQQUQH6SWHSSn2e92dQnpS2CMXgWXVx2bqCJpF2u7ty0A09qkFuZqop+eZS0qYPAWvufHu/i1IvSP1p32dXygvISEKm/vOaXP6OIdAIcCgYEA3s6XYesRXQrQiBsh5ce0XNDp7bsyCEveQz0cVD+5rA+l2FZF1WNVTt7e0Y5Yzo4kAGQvdkXBMRWYdkz9bjzkZHtR8r3Hgg0G6XXBTO1ooVrdhCEp0Faub4SVNj91IqA0RJO+jmWaEvmASTpN8Mn/zomNFyZsrFyaLiOdulRiR+8CgYEAgwdh7UrCbNgOEO6KhgzI4ZiofdfF9OCVGa+yu1IeCHPfx3KqntH6MGA/xBLytdMm2L2j3ax2nAANFzQsPJ3dIK2H0tOjE7lvdQVxrclmSKQ0A83ejb8lv7bywbEhWyffcnCk1P+pK6UTDDJnO3MwO51crKMl+x0VGS4HAnvtMEECgYA4OzOBhu4O6VfPwelAMLKYajFfykrKRTuHBLlNmfemMRzOCJf/Tt6M1Tqu8JoBJ2Z2otJHqzsixCyCTtP3Km8J3QXFmZfsfpUr/ogWfiRV9LTLUANZjUbg5jkyQ7mwT3ZhiFgjYAkOmOGDma9qAdEJszVkjlIG/if7VQnNqNZVCQKBgHfruxoZplK2BX+ldRGQRiPlos4eDHEc4wfVzD0KghxmtANmGbCa8o2jh3fRLoW9voJ8DKnWh4Eb1ClC6Pu/hK/exPcRYvAV9AhHO5TcEeNjW47pN76KXt+PY2arSkNYI/7OA+l3amLTJchI6Lkwpa2Nu9uPz7Bgn/bzY/Rtc4Uh")
	c.privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBS)
	if err != nil {
		panic(err)
	}

	//c.privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)

	c.privateKeyBS = x509.MarshalPKCS1PrivateKey(c.privateKey)

	c.publicKeyBS = x509.MarshalPKCS1PublicKey(&c.privateKey.PublicKey)
	c.publicKey64 = base64.StdEncoding.EncodeToString(c.publicKeyBS)
	c.publicKeyHex = hex.EncodeToString(c.publicKeyBS)
	fmt.Println("XCHG --- Address: ", c.publicKey64)
}

func (c *Client) getIPsByAddress(_ string) []string {
	return []string{"127.0.0.1"}
}

func (c *Client) findServerForHosting(publicKeyBS []byte) (resultIp string) {
	//fmt.Println("XCHG --- findServerForHosting", hex.EncodeToString(publicKeyBS))
	ips := c.getIPsByAddress(hex.EncodeToString(publicKeyBS))
	for _, ip := range ips {
		resultIp = ip
		break

		code, _, err := c.Request(c.httpClientPing, "http://"+ip+":8987", map[string][]byte{"f": []byte("i")})
		if err != nil {
			continue
		}
		if code == 200 {
			fmt.Println("XCHG --- server found: ", ip)
			resultIp = ip
			break
		}
	}
	fmt.Println("XCHG --- findServerForHosting result", resultIp)
	return
}

func (c *Client) findServerByAddress(addr string) (resultIp string) {
	//fmt.Println("findServerByAddress", addr)
	ips := c.getIPsByAddress(addr)
	for _, ip := range ips {
		code, _, err := c.Request(c.httpClientPing, "http://"+ip+":8987", map[string][]byte{"f": []byte("p"), "a": []byte(addr)})
		if err != nil {
			continue
		}
		if code == 200 {
			fmt.Println("server found: ", ip)
			resultIp = ip
			break
		}
	}
	fmt.Println("findServerByAddress result", resultIp)
	return
}

func (c *Client) Send(addr string, data []byte) (err error) {
	//fmt.Println("Send to", addr, "data_len:", len(data))
	var ok bool
	var code int
	currentIP := ""
	c.mtx.Lock()
	currentIP, ok = c.IPsByAddress[addr]
	c.mtx.Unlock()

	needToResend := false

	if ok && currentIP != "" {
		var resp []byte
		//fmt.Println("Send(1): found ip:", currentIP)
		code, resp, err = c.Request(c.httpClientSend, "http://"+currentIP+":8987", map[string][]byte{"f": []byte("w"), "a": []byte(addr), "d": data})
		if err != nil || code != 200 {
			fmt.Println("Send(1) error", err, code, string(resp))
			needToResend = true
			c.mtx.Lock()
			c.IPsByAddress[addr] = ""
			currentIP = ""
			c.mtx.Unlock()
		} else {
			//fmt.Println("Send(1) OK")
		}
	} else {
		needToResend = true
	}

	if needToResend {
		fmt.Println("resend")
		currentIP = c.findServerByAddress(addr)
		if currentIP != "" {
			code, _, err = c.Request(c.httpClientSend, "http://"+currentIP+":8987", map[string][]byte{"f": []byte("w"), "a": []byte(addr), "d": data})
			if code == 200 && err == nil {
				c.mtx.Lock()
				c.IPsByAddress[addr] = currentIP
				c.mtx.Unlock()
			}
		} else {
			err = errors.New("no route to host")
		}
	}

	return
}

func (c *Client) requestInit() error {
	var code int
	var data []byte
	var err error

	{
		requestInit1 := make([]byte, 0)
		requestInit1 = append(requestInit1, 0x00) // Init1
		requestInit1 = append(requestInit1, c.publicKeyBS...)

		code, data, err = c.Request(c.httpClientReceive, "http://"+c.xchgIP+":8987", map[string][]byte{"f": []byte("b"), "d": []byte(base64.StdEncoding.EncodeToString(requestInit1))})
		if err != nil {
			fmt.Println("rcv err:", err)
			c.xchgIP = ""
			return err
		}
		if code != 200 {
			fmt.Println("code:", code)
			c.xchgIP = ""
			return errors.New("Code != 200")
		}

		fmt.Println("Received Data init 1:", string(data))

		var encryptedBytes []byte
		var decryptedBytes []byte
		encryptedBytes, err = base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			fmt.Println("ERROR: ", err)
			return err
		}

		decryptedBytes, err = rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, encryptedBytes)
		if err != nil {
			return err
		}
		c.aesKey = decryptedBytes
		logger.Println("AES: ", hex.EncodeToString(c.aesKey))
		if err != nil {
			return err
		}
	}

	{
		requestInit2 := make([]byte, 1+4)
		requestInit2[0] = 0x01
		binary.LittleEndian.PutUint32(requestInit2[1:], uint32(len(c.publicKeyBS)))
		requestInit2 = append(requestInit2, c.publicKeyBS...)
		var encryptedPublicKey []byte
		encryptedPublicKey, err = c.encryptAES(c.publicKeyBS, c.aesKey)
		if err != nil {
			return err
		}

		requestInit2 = append(requestInit2, encryptedPublicKey...)

		code, data, err = c.Request(c.httpClientReceive, "http://"+c.xchgIP+":8987", map[string][]byte{"f": []byte("b"), "d": []byte(base64.StdEncoding.EncodeToString(requestInit2))})
		if err != nil {
			fmt.Println("rcv err:", err)
			c.xchgIP = ""
			return err
		}
		if code != 200 {
			fmt.Println("code:", code, string(data))
			c.xchgIP = ""
			return errors.New("Code != 200")
		}

		var encryptedBytes []byte
		//var decryptedBytes []byte
		encryptedBytes, err = base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return err
		}

		var init2Response []byte
		init2Response, err = c.decryptAES(encryptedBytes, c.aesKey)
		if err != nil {
			fmt.Println("1111")
			return err
		}

		if len(init2Response) != 16 {
			fmt.Println("11112222", len(init2Response))
			err = errors.New("len(init2Response) != 8")
			return err
		}

		c.lid = binary.LittleEndian.Uint64(init2Response[0:])
		c.counter = binary.LittleEndian.Uint64(init2Response[8:])

		fmt.Println("lid:", c.lid, "counter:", c.counter)
	}

	return nil
}

func (c *Client) encryptAES(decryptedMessage []byte, key []byte) (encryptedMessage []byte, err error) {
	var ch cipher.Block
	ch, err = aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(ch)
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	encryptedMessage = gcm.Seal(nonce, nonce, decryptedMessage, nil)
	return
}

func (c *Client) thRcv() {
	var code int
	var data []byte
	var err error

	for !c.stopping {
		if c.xchgIP == "" {
			c.xchgIP = c.findServerForHosting(c.publicKeyBS)
		}

		if c.xchgIP == "" {
			time.Sleep(1 * time.Second)
			fmt.Println("no server for hosting found")
			continue
		}

		if len(c.aesKey) != 32 {
			err = c.requestInit()
			if err != nil {
				fmt.Println("XCHG -- no secret bytes", err)
				time.Sleep(1 * time.Second)
				c.xchgIP = ""
				c.aesKey = nil
				c.lid = 0
				c.counter = 0
				continue
			}
		}

		if len(c.aesKey) != 32 {
			time.Sleep(1 * time.Second)
			fmt.Println("XCHG -- no secret bytes")
			c.xchgIP = ""
			c.aesKey = nil
			c.lid = 0
			c.counter = 0
			continue
		}

		var ch cipher.Block
		ch, err = aes.NewCipher(c.aesKey)
		if err != nil {
			time.Sleep(1 * time.Second)
			fmt.Println("XCHG -- cannot create Cipher")
			c.xchgIP = ""
			c.aesKey = nil
			c.counter = 0
			c.lid = 0
			continue
		}
		var gcm cipher.AEAD
		gcm, err = cipher.NewGCM(ch)
		nonce := make([]byte, gcm.NonceSize())
		_, err = io.ReadFull(rand.Reader, nonce)
		if err != nil {
			time.Sleep(1 * time.Second)
			fmt.Println("XCHG -- cannot fill nonce")
			c.xchgIP = ""
			c.aesKey = nil
			c.counter = 0
			c.lid = 0
			continue
		}

		c.counter++

		readRequestBS := make([]byte, 9)
		readRequestBS[0] = 0x02
		binary.LittleEndian.PutUint64(readRequestBS[1:], c.lid)
		counterBS := make([]byte, 8)
		binary.LittleEndian.PutUint64(counterBS, c.counter)
		encryptedCounter := gcm.Seal(nonce, nonce, counterBS, nil)
		readRequestBS = append(readRequestBS, encryptedCounter...)

		fmt.Println("XCHG - READ")
		code, data, err = c.Request(c.httpClientReceive, "http://"+c.xchgIP+":8987", map[string][]byte{"f": []byte("b"), "d": []byte(base64.StdEncoding.EncodeToString(readRequestBS))})
		if err != nil {
			fmt.Println("XCHG - READ ERR:", err)
			c.xchgIP = ""
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		/*if code == 502 {
			fmt.Println("Wrong Counter", code)
			time.Sleep(1 * time.Second)
			continue
		}*/

		//fmt.Println("Code", code, data)

		if code != 200 && code != 204 {

			c.xchgIP = ""
			c.aesKey = nil
			c.lid = 0
			time.Sleep(1 * time.Second)
			continue
		}

		if code == 200 {
			if len(data) > 0 {
				data, _ = base64.StdEncoding.DecodeString(string(data))
				data, err = c.decryptAES(data, c.aesKey)
				if err != nil {
					fmt.Println("Decrypt error", err)
					c.xchgIP = ""
					c.aesKey = nil
					c.lid = 0
					time.Sleep(1 * time.Second)
					continue
				}
				transactionId := binary.LittleEndian.Uint64(data[0:])
				data = data[8:]
				fmt.Println("Received request", transactionId, string(data))
				response, err := c.OnReceived(data)
				if err != nil {
					continue
				}
				fmt.Println("RESPONSE: ", string(response))

				{
					putRequestBS := make([]byte, 9)
					putRequestBS[0] = 0x03
					binary.LittleEndian.PutUint64(putRequestBS[1:], c.lid)

					responseBS := make([]byte, 8)
					binary.LittleEndian.PutUint64(responseBS, transactionId)
					responseBS = append(responseBS, response...)
					encryptedResponse := gcm.Seal(nonce, nonce, responseBS, nil)

					putRequestBS = append(putRequestBS, encryptedResponse...)
					code, data, err = c.Request(c.httpClientReceive, "http://"+c.xchgIP+":8987", map[string][]byte{"f": []byte("b"), "d": []byte(base64.StdEncoding.EncodeToString(putRequestBS))})
				}

				_ = response
			}

		}
	}
}

func (c *Client) decryptAES(message []byte, key []byte) (decryptedMessage []byte, err error) {
	ch, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(ch)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(message) < nonceSize {
		return nil, errors.New("wrong nonce")
	}
	nonce, ciphertext := message[:nonceSize], message[nonceSize:]
	decryptedMessage, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return
	}
	return
}

func (c *Client) Request(httpClient *http.Client, url string, parameters map[string][]byte) (code int, data []byte, err error) {

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range parameters {
		var fw io.Writer
		fw, err = writer.CreateFormField(key)
		if err != nil {
			return
		}
		_, err = fw.Write(value)
		if err != nil {
			return
		}
	}
	err = writer.Close()
	if err != nil {
		return
	}
	var response *http.Response
	response, err = c.Post(httpClient, url, writer.FormDataContentType(), &body)
	if err != nil {
		return
	}
	code = response.StatusCode
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		_ = response.Body.Close()
		return
	}
	_ = response.Body.Close()
	return
}

func (c *Client) Post(httpClient *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return httpClient.Do(req)
}
