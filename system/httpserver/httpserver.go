package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	nodeinterface2 "github.com/ipoluianov/gazer_node/system/protocols/nodeinterface"
	"github.com/ipoluianov/gazer_node/system/system"
	"github.com/ipoluianov/gazer_node/utilities/logger"
	"github.com/ipoluianov/gazer_node/utilities/packer"
)

type HttpServer struct {
	srv      *http.Server
	r        *mux.Router
	system   *system.System
	rootPath string

	stopping bool
}

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func NewHttpServer(sys *system.System) *HttpServer {
	var c HttpServer
	c.rootPath = CurrentExePath() + "/www"
	c.system = sys
	return &c
}

func (c *HttpServer) Start() {
	//logger.Println("HttpServer start")

	//generateTLS(c.system.Settings())

	c.r = mux.NewRouter()

	// API
	c.r.HandleFunc("/api/request", c.processApiRequest)

	// Static files
	c.r.NotFoundHandler = http.HandlerFunc(c.processFile)

	/*cert, err := tls.X509KeyPair(certPublic(c.system.Settings()), certPrivate(c.system.Settings()))
	if err != nil {
		logger.Println("[HttpServer]", "Start error(X509KeyPair):", err)
		return
	}*/
	c.srv = &http.Server{
		Addr: ":8084",
		/*TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},*/
	}

	//c.srv = &http.Server{Addr: ":8084"} // 127.0.0.1
	c.srv.Handler = c.r
	go c.thListen()
}

func (c *HttpServer) thListen() {
	//logger.Println("HttpServer thListen begin")
	err := c.srv.ListenAndServe()
	if err != nil {
		logger.Println("HttpServer thListen error: ", err)
	}
	logger.Println("HttpServer thListen end")
}

func (c *HttpServer) Stop() error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = c.srv.Shutdown(ctx); err != nil {
		logger.Println(err)
	}
	return err
}

/*func (c *HttpServer) Request(requestText string) (string, error) {
	var err error
	var respBytes []byte

	type Request struct {
		Function string `json:"func"`
		Path     string `json:"path"`
		Layer    string `json:"layer"`
	}
	var req Request
	err = json.Unmarshal([]byte(requestText), &req)
	if err != nil {
		return "", err
	}

	type Response struct {
		Value    string `json:"v"`
		DateTime string `json:"t"`
		Error    string `json:"e"`
	}

	var resp Response
	resp.Value = "123"
	resp.DateTime = time.Now().Format("2006-01-02 15-04-05.999")
	resp.Error = "ok"

	respBytes, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		return "", err
	}

	return string(respBytes), nil
}*/

func (c *HttpServer) processApiRequest(w http.ResponseWriter, r *http.Request) {
	var err error
	var responseText []byte
	var sessionToken string
	usingZ := false

	requestJson := r.FormValue("rj")
	requestType := r.FormValue("rt")
	requestJsonZ := r.FormValue("rjz")
	function := r.FormValue("fn")
	sessionToken = r.FormValue("s")

	if r.Method == "POST" {
		if err := r.ParseMultipartForm(1000000); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		requestJson = r.FormValue("rj")
		requestType = r.FormValue("rt")
		requestJsonZ = r.FormValue("rjz")
		function = r.FormValue("fn")
	}

	if requestType == "z" {
		requestJson = packer.UnpackString(requestJsonZ)
		usingZ = true
	}

	if requestJson == "" {
		requestJson = "{}"
	}

	//if strings.Contains(function, "session") {
	//logger.Println("function", function, "request", requestJson)
	//}

	if len(function) > 0 {
		if sessionToken == "" {
			sessionTokenCookie, errSessionToken := r.Cookie("session_token")
			if errSessionToken == nil {
				sessionToken = sessionTokenCookie.Value
			}
		}

		if err == nil {
			responseText, err = c.RequestJson(function, []byte(requestJson), r.RemoteAddr, false)
		}

		if function == nodeinterface2.FuncSessionOpen && err == nil {
			// Set cookie
			var sessionOpenResponse nodeinterface2.SessionOpenResponse
			errSessionOpenResp := json.Unmarshal(responseText, &sessionOpenResponse)
			if errSessionOpenResp == nil {
				expiration := time.Now().Add(365 * 24 * time.Hour)
				cookie := http.Cookie{Name: "session_token", Path: "/", Value: sessionOpenResponse.SessionToken, Expires: expiration}
				http.SetCookie(w, &cookie)
			}
		}

		if function == nodeinterface2.FuncSessionRemove && err == nil {
			// Set cookie
			var sessionRemoveRequest nodeinterface2.SessionRemoveRequest
			errSessionOpenResp := json.Unmarshal([]byte(requestJson), &sessionRemoveRequest)
			if errSessionOpenResp == nil {
				if sessionRemoveRequest.SessionToken == sessionToken {
					expiration := time.Now().Add(-365 * 24 * time.Hour)
					cookie := http.Cookie{Name: "session_token", Path: "/", Value: "", Expires: expiration}
					http.SetCookie(w, &cookie)
				}
			}
		}

		if function == nodeinterface2.FuncSessionActivate && err == nil {
			// Set cookie
			var sessionActivateResponse nodeinterface2.SessionActivateResponse
			errSessionActivateResp := json.Unmarshal(responseText, &sessionActivateResponse)
			if errSessionActivateResp == nil {
				expiration := time.Now().Add(365 * 24 * time.Hour)
				cookie := http.Cookie{Name: "session_token", Value: sessionActivateResponse.SessionToken, Expires: expiration}
				http.SetCookie(w, &cookie)
			}
		}
	}

	if err != nil {
		w.WriteHeader(500)
		b := []byte(err.Error())
		_, _ = w.Write(b)
		return
	}

	if usingZ {
		//println("local call", function, requestJson)
		responseText = packer.PackBytes(responseText)
	}

	_, _ = w.Write([]byte(responseText))
}

func SplitRequest(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/'
	})
}

func (c *HttpServer) processFile(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	//c.processFileLocal(w, r)
	//c.file(w, r, r.URL.Path)
}
