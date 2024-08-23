package fs

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type FileIdentityAPI api

type GetFileIdentitiesRequest struct {
	// TenantIDs the tenant id
	TenantIDs []int `binding:"required" form:"tenantID"`
	// IsDefault The default identity of the tenant
	IsDefault bool `form:"isDefault"`
}

type ID int

func (i *ID) UnmarshalJSON(data []byte) error {
	// 去除引号
	str := strings.Trim(string(data), `"`)
	// 将字符串转换为int
	val, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	// int赋值
	*i = ID(val)
	return nil
}

func (i *ID) String() string {
	return strconv.Itoa(int(*i))
}

type Result struct {
	Data Data `json:"data"`
}

type Data struct {
	FileIdentitiesForApp []*FileIdentity `json:"fileIdentitiesForApp"`
}

type FileIdentity struct {
	ID              ID          `json:"id"`
	TenantID        ID          `json:"tenantID"`
	AccessKeyID     string      `json:"accessKeyID"`
	AccessKeySecret string      `json:"accessKeySecret"`
	RoleArn         string      `json:"roleArn"`
	Policy          string      `json:"policy,omitempty"`
	DurationSeconds int         `json:"durationSeconds,omitempty"`
	IsDefault       bool        `json:"isDefault"`
	Source          *FileSource `json:"source"`
}

type FileSource struct {
	ID                ID     `json:"id,omitempty"`
	Kind              Kind   `json:"kind,omitempty"`
	Endpoint          string `json:"endpoint,omitempty"`
	EndpointImmutable bool   `json:"endpointImmutable,omitempty"`
	StsEndpoint       string `json:"stsEndpoint,omitempty"`
	Region            string `json:"region,omitempty"`
	Bucket            string `json:"bucket,omitempty"`
	BucketURL         string `json:"bucketUrl,omitempty"`
}

// GetFileIdentities (POST fileIdentitiesFull)
func (a *FileIdentityAPI) GetFileIdentities(ctx context.Context, req *GetFileIdentitiesRequest) (ret []*FileIdentity, resp *http.Response, err error) {
	query := buildGraphQLQuery(req)
	body := GraphqlRequest{Query: query}
	contentType := "application/json"
	path := "/graphql/query"

	request, err := a.client.prepareRequest("POST", a.client.cfg.BasePath+path, contentType, body)
	if err != nil {
		return
	}
	resp, err = a.client.Do(ctx, request)
	if err != nil {
		return
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusOK {
		result := Result{}
		err = a.client.decode(respBody, &result, resp.Header.Get("Content-Type"))
		if err == nil {
			ret = result.Data.FileIdentitiesForApp
			return
		}
	} else if resp.StatusCode >= 300 {
		err = errors.New(string(respBody))
	}
	return
}

func buildGraphQLQuery(req *GetFileIdentitiesRequest) string {
	var where []string
	ids := make([]string, len(req.TenantIDs))
	if len(req.TenantIDs) != 0 {
		for i, num := range req.TenantIDs {
			ids[i] = strconv.Itoa(num)
		}
		where = append(where, "tenantIDIn:["+strings.Join(ids, ",")+"]")
	}
	if !req.IsDefault {
		where = append(where, "isDefault: false")
	}

	return `query {
		fileIdentitiesForApp(where:{` + strings.Join(where, ",") + `}){
			id,tenantID,accessKeyID,accessKeySecret,roleArn,policy,durationSeconds,isDefault,
			source{
				id,kind,endpoint,endpointImmutable,stsEndpoint,region,bucket,bucketURL
			}
		}
	}`
}

func ToProviderConfig(fi *FileIdentity) *ProviderConfig {
	return &ProviderConfig{
		Kind:              fi.Source.Kind,
		Bucket:            fi.Source.Bucket,
		BucketUrl:         fi.Source.BucketURL,
		Endpoint:          fi.Source.Endpoint,
		EndpointImmutable: fi.Source.EndpointImmutable,
		StsEndpoint:       fi.Source.StsEndpoint,
		AccessKeyID:       fi.AccessKeyID,
		AccessKeySecret:   fi.AccessKeySecret,
		Policy:            fi.Policy,
		Region:            fi.Source.Region,
		RoleArn:           fi.RoleArn,
		DurationSeconds:   fi.DurationSeconds,
	}
}
