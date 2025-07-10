package api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"
	"github.com/tsingsun/woocoo/pkg/cache/lfu"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/gds"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/woocoos/knockout-go/api/auth"
	"github.com/woocoos/knockout-go/api/fs"
	"github.com/woocoos/knockout-go/api/msg"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

// AssumeRoleResponse contains the result of successful AssumeRole request.
type AssumeRoleResponse struct {
	XMLName xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleResponse" json:"-"`

	Result           AssumeRoleResult `xml:"AssumeRoleResult"`
	ResponseMetadata struct {
		RequestID string `xml:"RequestId,omitempty"`
	} `xml:"ResponseMetadata,omitempty"`
}

// AssumeRoleResult - Contains the response to a successful AssumeRole
// request, including temporary credentials that can be used to make
// MinIO API requests.
type AssumeRoleResult struct {
	// The identifiers for the temporary security credentials that the operation
	// returns.
	AssumedRoleUser AssumedRoleUser `xml:",omitempty"`

	// The temporary security credentials, which include an access key ID, a secret
	// access key, and a security (or session) token.
	//
	// Note: The size of the security token that STS APIs return is not fixed. We
	// strongly recommend that you make no assumptions about the maximum size. As
	// of this writing, the typical size is less than 4096 bytes, but that can vary.
	// Also, future updates to AWS might require larger sizes.
	Credentials struct {
		AccessKey    string    `xml:"AccessKeyId" json:"accessKey,omitempty"`
		SecretKey    string    `xml:"SecretAccessKey" json:"secretKey,omitempty"`
		Expiration   time.Time `xml:"Expiration" json:"expiration,omitempty"`
		SessionToken string    `xml:"SessionToken" json:"sessionToken,omitempty"`
	} `xml:",omitempty"`

	// A percentage value that indicates the size of the policy in packed form.
	// The service rejects any policy with a packed size greater than 100 percent,
	// which means the policy exceeded the allowed space.
	PackedPolicySize int `xml:",omitempty"`
}

// AssumedRoleUser - The identifiers for the temporary security credentials that
// the operation returns. Please also see https://docs.aws.amazon.com/goto/WebAPI/sts-2011-06-15/AssumedRoleUser
type AssumedRoleUser struct {
	Arn           string
	AssumedRoleID string `xml:"AssumeRoleId"`
}

type apiSuite struct {
	suite.Suite
	sdk           *SDK
	mockServerUrl string
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
	mux.HandleFunc(`/sts`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		resp := AssumeRoleResponse{
			Result: AssumeRoleResult{
				AssumedRoleUser: AssumedRoleUser{
					Arn:           "arn:aws:sts::123456789012:assumed-role/role-name/role-session-name",
					AssumedRoleID: "role-session-name",
				},
				Credentials: struct {
					AccessKey    string    `xml:"AccessKeyId" json:"accessKey,omitempty"`
					SecretKey    string    `xml:"SecretAccessKey" json:"secretKey,omitempty"`
					Expiration   time.Time `xml:"Expiration" json:"expiration,omitempty"`
					SessionToken string    `xml:"SessionToken" json:"sessionToken,omitempty"`
				}{
					"test", "test1234", time.Now().Add(1 * time.Hour), "test1234",
				},
			},
		}
		d, err := xml.Marshal(resp)
		t.Require().NoError(err)
		w.Write(d)
	})
	// GetObjectTest,'fstest' is the bucket name
	mux.HandleFunc(`/fstest/`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		// mock body
		w.Write([]byte("test"))
	})
	mux.HandleFunc("/graphql/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		as, err := io.ReadAll(r.Body)
		t.Require().NoError(err)
		var req fs.GraphqlRequest
		t.Require().NoError(json.Unmarshal(as, &req))
		// 解析query中的operation
		doc, err := parser.ParseQuery(&ast.Source{Input: req.Query})
		op := doc.Operations[0].SelectionSet[0]
		opf := op.(*ast.Field)
		ret := make([]byte, 0)
		if opf.Name == "fileIdentitiesForApp" {
			resp := fs.Result{
				Data: fs.Data{
					FileIdentitiesForApp: []*fs.FileIdentity{
						{
							ID:              fs.ID(1),
							TenantID:        fs.ID(1),
							AccessKeyID:     "test",
							AccessKeySecret: "test1234",
							RoleArn:         "arn:aws:s3:::*",
							Policy:          "",
							DurationSeconds: 3600,
							IsDefault:       true,
							Source: &fs.FileSource{
								ID:                fs.ID(1),
								Bucket:            "fstest",
								BucketURL:         t.mockServerUrl + "/fstest",
								Endpoint:          t.mockServerUrl,
								EndpointImmutable: false,
								Kind:              fs.KindMinio,
								Region:            "minio",
								StsEndpoint:       t.mockServerUrl + "/sts",
							},
						},
					},
				},
			}
			ret, err = json.Marshal(resp)
			t.Require().NoError(err)
		}
		w.Write(ret)
	})
	mux.HandleFunc("/org/domain", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ah := r.Header.Values("Authorization")
			t.Require().Len(ah, 2)
			w.Header().Set("Content-Type", "application/json")
			as, err := io.ReadAll(r.Body)
			t.Require().NoError(err)
			var req auth.GetDomainRequest
			t.Require().NoError(json.Unmarshal(as, &req))
			resp := auth.Domain{
				ID:             req.OrgID,
				Name:           "test",
				LocalCurrency:  "HKD",
				ParentCurrency: "HKD",
				ParentID:       1000,
				ParentName:     "test",
			}
			ret, err := json.Marshal(resp)
			t.Require().NoError(err)
			w.Write(ret)
		}
	})
	return mux
}

func (t *apiSuite) SetupSuite() {
	srv := httptest.NewServer(t.mockHttpServer())
	t.mockServerUrl = srv.URL
	cnf := conf.New(conf.WithLocalPath(filepath.Join("testdata", "kocfg.yaml"))).Load().Sub("all")
	cnf.Parser().Set("kosdk.client.oauth2.endpoint.tokenURL", t.mockServerUrl+"/token")
	cnf.Parser().Set("kosdk.plugin.fs.basePath", srv.URL)
	// fs
	fskey := "kosdk.plugin.fs.providers"
	fcfg := cnf.Parser().Get(fskey)
	fss := fcfg.([]any)
	fsit := fss[0].(map[string]any)
	fsit["endpoint"] = t.mockServerUrl
	fsit["stsEndpoint"] = t.mockServerUrl + "/sts"
	fsit["bucketUrl"] = t.mockServerUrl + "/fstest"
	cnf.Parser().Set(fskey, fss)

	_, err := lfu.NewTinyLFU(cnf.Sub("cache.memory"))
	t.Require().NoError(err)
	sdk, err := NewSDK(cnf.Sub("kosdk"))
	t.Require().NoError(err)
	err = sdk.RegisterPlugin(PluginMsg, conf.NewFromStringMap(map[string]any{
		"basePath": srv.URL,
	}))
	err = sdk.RegisterPlugin(PluginAuth, conf.NewFromStringMap(map[string]any{
		"basePath": srv.URL,
	}))
	t.sdk = sdk
}

func (t *apiSuite) TestGetToken() {
	tk, err := t.sdk.GetToken()
	t.Require().NoError(err)
	t.Equal("90d64460d14870c08c81352a05dedd3465940a7c", tk.AccessToken)
}
func (t *apiSuite) TestDefaultSDK() {
	cnf := conf.New(conf.WithLocalPath(filepath.Join("testdata", "kocfg.yaml"))).Load().Sub("default.kosdk")
	sdk, err := DefaultSDK(cnf)
	t.Require().NoError(err)
	t.NotNil(sdk.signer)
}

func (t *apiSuite) TestGetPlugin() {
	_, ok := t.sdk.GetPlugin(PluginFS)
	t.Require().True(ok)
	_, ok = t.sdk.GetPlugin(PluginMsg)
	t.Require().True(ok)
	_, ok = t.sdk.GetPlugin(PluginAuth)
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
					EndsAt: gds.Ptr(time.Now()),
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

func (t *apiSuite) TestFs() {
	t.Run("minio", func() {
		sc := &fs.ProviderConfig{
			AccessKeyID: "test",
			Kind:        "minio",
			Endpoint:    t.mockServerUrl,
			Bucket:      "fstest",
		}
		p, err := t.sdk.Fs().Client.GetProviderByBizKey(fs.GetProviderKey(sc))
		t.Require().NoError(err)
		resp, err := p.GetSTS(context.Background(), "anyname")
		t.Require().NoError(err)
		t.NotNil(resp)
		t.NotEmpty(resp.AccessKeyID)
		got, err := p.S3Client().GetObject(context.Background(), &s3.GetObjectInput{
			Bucket: &sc.Bucket,
			Key:    &sc.Bucket,
		})
		t.Require().NoError(err)
		// read body
		gt, err := io.ReadAll(got.Body)
		t.Require().NoError(err)
		t.Equal("test", string(gt))
	})

	t.Run("fileIdentitiesForApp", func() {
		isDefault := true
		ret, resp, err := t.sdk.Fs().FileIdentityAPI.GetFileIdentities(context.Background(), &fs.GetFileIdentitiesRequest{
			IsDefault: isDefault,
			TenantIDs: []int{1, 2, 3},
		})
		t.Require().NoError(err)
		t.NotNil(ret)
		t.Equal(200, resp.StatusCode)

		bizKey := fs.GetProviderKey(fs.ToProviderConfig(ret[0]))
		t.NotNil(bizKey)
		err = t.sdk.Fs().Client.RegistryProvider(fs.ToProviderConfig(ret[0]), bizKey)
		t.Require().NoError(err)
		provider, err := t.sdk.Fs().Client.GetProviderByBizKey(bizKey)
		t.Require().NoError(err)
		stsResp, err := provider.GetSTS(context.Background(), "anyname")
		t.Require().NoError(err)
		t.NotNil(stsResp)
		t.NotEmpty(stsResp.AccessKeyID)
	})
}
func (t *apiSuite) TestAuth() {
	t.Run("domain", func() {
		ret, resp, err := t.sdk.Auth().AuthAPI.GetDomain(context.Background(), &auth.GetDomainRequest{
			OrgID: 1,
		})
		t.Require().NoError(err)
		t.NotNil(ret)
		t.Equal(200, resp.StatusCode)
	})
}
