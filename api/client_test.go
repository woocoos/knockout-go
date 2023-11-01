package api

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"github.com/tsingsun/woocoo/pkg/cache/lfu"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/woocoos/knockout-go/api/file"
	"github.com/woocoos/knockout-go/api/msg"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	cnfStr = `
kosdk:
  client:
    timeout: 2s
    oauth2:
      clientID: 206734260394752
      clientSecret: T2UlqISVFq4DR9InXamj3l74iWdu3Tyr
      endpoint:
        tokenURL: http://127.0.0.1:10070/token
      scopes:
      storeKey: local
  signer:
    authScheme: "KO-HMAC-SHA1"
    authHeaders:  ["timestamp", "nonce"]
    signedLookups:
      accessToken: "header:authorization>bearer"
      timestamp:
      nonce:
      url: CanonicalUri
    nonceLen: 12
  plugin:
    file:
      basePath: http://127.0.0.1:10070
    msg:
      basePath: http://127.0.0.1:10070
cache:
  memory:
    driverName: local
    size: 10000
    samples: 10000
`
)

type apiSuite struct {
	suite.Suite
	sdk *SDK
}

func TestApiSuite(t *testing.T) {
	suite.Run(t, &apiSuite{})
}

func (t *apiSuite) mockHttpServer() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		d, err := json.Marshal(map[string]string{
			"access_token": "90d64460d14870c08c81352a05dedd3465940a7c",
			"expires_in":   "11", // defaultExpiryDelta = 10 * time.Second, so set 11 seconds and sleep 1 second
			"scope":        "user",
			"token_type":   "bearer",
		})
		t.Require().NoError(err)
		w.Write(d)
	})
	mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ah := r.Header.Values("Authorization")
			t.Require().Len(ah, 2)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
		case http.MethodPost:
			as, err := io.ReadAll(r.Body)
			t.Require().NoError(err)
			var data msg.PostAlertsRequest
			t.Require().NoError(json.Unmarshal(as, &data))
			t.Require().Len(data.PostableAlerts, 1)
			w.WriteHeader(http.StatusOK)
		}
	})
	mux.HandleFunc(`/files/`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		d, err := json.Marshal(file.FileInfo{
			ID: "1",
		})
		t.Require().NoError(err)
		w.Write(d)
	})
	return mux
}

func (t *apiSuite) SetupSuite() {
	srv := httptest.NewServer(t.mockHttpServer())
	cnf := conf.NewFromBytes([]byte(cnfStr))
	cnf.Parser().Set("kosdk.client.oauth2.endpoint.tokenURL", srv.URL+"/token")
	_, err := lfu.NewTinyLFU(cnf.Sub("cache.memory"))
	t.Require().NoError(err)
	sdk, err := NewSDK(cnf.Sub("kosdk"))
	t.Require().NoError(err)
	err = sdk.RegisterPlugin("file", conf.NewFromStringMap(map[string]any{
		"basePath": srv.URL,
	}))
	t.Require().NoError(err)
	err = sdk.RegisterPlugin("msg", conf.NewFromStringMap(map[string]any{
		"basePath": srv.URL,
	}))
	t.sdk = sdk
}

func (t *apiSuite) TestGetPlugin() {
	_, ok := t.sdk.GetPlugin("file")
	t.Require().True(ok)
	_, ok = t.sdk.GetPlugin("msg")
	t.Require().True(ok)
	_, ok = t.sdk.GetPlugin("not-exist")
	t.Require().False(ok)
}

func (t *apiSuite) TestMsg() {
	t.Run("getAlerts", func() {
		ret, resp, err := t.sdk.Msg().AlertAPI.GetAlerts(context.Background(), &msg.GetAlertsRequest{
			Active:      nil,
			Silenced:    nil,
			Inhibited:   nil,
			Unprocessed: nil,
			Filter:      nil,
			Receiver:    nil,
		})
		t.Require().NoError(err)
		t.NotNil(ret)
		t.Equal(200, resp.StatusCode)
	})
	t.Run("postAlerts", func() {
		resp, err := t.sdk.Msg().AlertAPI.PostAlerts(context.Background(), &msg.PostAlertsRequest{
			PostableAlerts: msg.PostableAlerts{
				{
					EndsAt: time.Now(),
					Alert: &msg.Alert{
						Labels: map[string]string{
							"summary": "test",
						},
					},
					Annotations: map[string]string{
						"annotation": "test",
					},
				},
			},
		})
		t.Require().NoError(err)
		t.Equal(200, resp.StatusCode)
	})
}

func (t *apiSuite) TestFile() {
	ret, resp, err := t.sdk.File().FileAPI.GetFile(context.Background(), &file.GetFileRequest{
		FileId: "1",
	})
	t.Require().NoError(err)
	t.NotNil(ret)
	t.Equal(200, resp.StatusCode)
}
