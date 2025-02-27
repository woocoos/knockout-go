package sdkmock

import (
	"encoding/json"
	"github.com/woocoos/knockout-go/api/msg"
	"net/http"
	"net/http/httptest"
)

type Server struct {
	*httptest.Server
	serverMux *http.ServeMux
	handler   map[string]http.HandlerFunc
}

// NewServer create a mock server.
// handler is a map of path to handler function.支持根据测试场景重定义请求处理.
// 支持的路径有:
//
//	/token
//	/api/v2/alerts
func NewServer(handler map[string]http.HandlerFunc) *Server {
	s := Server{
		serverMux: http.NewServeMux(),
		handler:   make(map[string]http.HandlerFunc),
	}
	if len(handler) > 0 {
		s.handler = handler
	}
	s.token()
	s.msg()
	s.Server = httptest.NewServer(s.serverMux)
	return &s
}

func (s *Server) token() {
	s.serverMux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if handler := s.handler["/token"]; handler != nil {
			handler(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		d, _ := json.Marshal(map[string]any{
			"access_token": "mock",
			"expires_in":   3600, // defaultExpiryDelta = 10 * time.Second, so set 11 seconds and sleep 1 second
			"scope":        "user",
			"token_type":   "bearer",
		})
		w.Write(d)
	})
}

func (s *Server) msg() {
	// http://127.0.0.1:10072/api/v2/alerts
	s.serverMux.HandleFunc("/api/v2/alerts", func(w http.ResponseWriter, r *http.Request) {
		if handler := s.handler["/api/v2/alerts"]; handler != nil {
			handler(w, r)
			return
		}
		vs := msg.GettableAlerts{}
		w.Header().Set("Content-Type", "application/json")
		d, _ := json.Marshal(vs)
		w.Write(d)
	})
}
