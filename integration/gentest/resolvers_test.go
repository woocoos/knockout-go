package gentest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gin-gonic/gin"
	kjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"github.com/tsingsun/woocoo/pkg/gds"
	"github.com/woocoos/knockout-go/integration/gentest/ent"
	"github.com/woocoos/knockout-go/integration/gentest/ent/user"
	"github.com/woocoos/knockout-go/integration/helloapp/ent/migrate"
	"github.com/woocoos/knockout-go/pkg/pagination"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/woocoos/knockout-go/integration/gentest/ent/runtime"
)

type TestSuite struct {
	suite.Suite
	client           *ent.Client
	queryResolver    queryResolver
	mutationResolver mutationResolver
}

func (s *TestSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	dr, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	s.Require().NoError(err)
	s.client = ent.NewClient(ent.Driver(dr), ent.Debug())
	s.queryResolver = queryResolver{&Resolver{s.client}}
	s.mutationResolver = mutationResolver{&Resolver{s.client}}
	s.NoError(s.client.Schema.Create(context.Background(),
		migrate.WithForeignKeys(false)),
	)

}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) refreshData() {
	_, _ = s.client.User.Delete().Exec(context.Background())
	_, _ = s.client.RefSchema.Delete().Exec(context.Background())
	builder := make([]*ent.UserCreate, 0)
	for i := 0; i < 20; i++ {
		row := s.client.User.Create()
		row.SetName("user" + strconv.Itoa(i)).Mutation().SetID(i + 1)
		builder = append(builder, row)
	}
	refBuilder := make([]*ent.RefSchemaCreate, 0)
	for i := 0; i < 20; i++ {
		row := s.client.RefSchema.Create()
		row.SetName("ref" + strconv.Itoa(i)).SetUserID(1) //  user1
		refBuilder = append(refBuilder, row)
	}

	s.NoError(s.client.User.CreateBulk(builder...).Exec(context.Background()))
	s.NoError(s.client.RefSchema.CreateBulk(refBuilder...).Exec(context.Background()))
}

func (s *TestSuite) TestUsers_SimplePagination() {
	s.refreshData()
	all := s.client.User.Query().AllX(context.Background())
	s.T().Log(all[0].ID, all[19].ID)
	s.Run("pagination after", func() {
		ctx, _ := gin.CreateTestContext(nil)
		users, err := s.queryResolver.Users(ctx, &ent.Cursor{ID: 2}, gds.Ptr(2), nil, nil, nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(3, users.Edges[0].Node.ID)
		s.Equal(4, users.Edges[1].Node.ID)

	})

	s.Run("simple after", func() {
		ctx := pagination.WithSimplePagination(context.Background(), &pagination.SimplePagination{PageIndex: 3, CurrentIndex: 1})
		users, err := s.queryResolver.Users(ctx, &ent.Cursor{ID: 2}, gds.Ptr(2), nil, nil, nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(5, users.Edges[0].Node.ID)
		s.Equal(6, users.Edges[1].Node.ID)

	})

	s.Run("pagination before", func() {
		ctx, _ := gin.CreateTestContext(nil)
		users, err := s.queryResolver.Users(ctx, nil, nil, &ent.Cursor{ID: 5}, gds.Ptr(2), nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(3, users.Edges[0].Node.ID)
		s.Equal(4, users.Edges[1].Node.ID)

	})

	s.Run("simple before", func() {
		ctx := pagination.WithSimplePagination(context.Background(), &pagination.SimplePagination{PageIndex: 1, CurrentIndex: 3})
		users, err := s.queryResolver.Users(ctx, nil, nil, &ent.Cursor{ID: 5}, gds.Ptr(2), nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(1, users.Edges[0].Node.ID)
		s.Equal(2, users.Edges[1].Node.ID)
	})
}

// graphQLQueryToRequestBody 将 GraphQL 查询语句转换为 HTTP 请求体
func graphQLQueryToRequestBody(query string, variables map[string]any) (string, error) {
	requestBody := map[string]any{
		"query":     query,
		"variables": variables,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (s *TestSuite) TestNode() {
	s.refreshData()
	srv := handler.New(NewSchema(s.client))
	srv.AddTransport(transport.POST{})
	s.Run("node", func() {
		w := httptest.NewRecorder()
		id := base64.StdEncoding.EncodeToString([]byte("users:1"))
		gb, err := graphQLQueryToRequestBody(`
query user($id: GID!) {
	node(id: $id) {
    	... on User {
			id
			name
	    }
    }
}
`, map[string]any{"id": id})
		s.Require().NoError(err)
		bd := strings.NewReader(gb)
		r := httptest.NewRequest("POST", "/graphql/query", bd)
		r.Header.Set("Content-Type", "application/json")
		srv.ServeHTTP(w, r)
		s.Contains(w.Body.String(), `{"id":"1","name":"user0"}`)
	})
	s.Run("node-ref", func() {
		w := httptest.NewRecorder()
		id := base64.StdEncoding.EncodeToString([]byte("users:1"))
		gb, err := graphQLQueryToRequestBody(`
query user($id: GID!) {
	node(id: $id) {
    	... on User {
			id
			refs(first:2) {
              edges {
                node {
                  name
                }
              }
            }
	    }
    }
}
`, map[string]any{"id": id})

		s.Require().NoError(err)
		bd := strings.NewReader(gb)
		r := httptest.NewRequest("POST", "/graphql/query", bd)
		r.Header.Set("Content-Type", "application/json")
		srv.ServeHTTP(w, r)
		kj := koanf.New(".")
		err = kj.Load(rawbytes.Provider(w.Body.Bytes()), kjson.Parser())
		s.Require().NoError(err)
		s.Contains(w.Body.String(), `{"name":"ref0"}`)
		s.Len(kj.Slices("data.node.refs.edges"), 2)
	})
	s.Run("nodes", func() {
		w := httptest.NewRecorder()
		ids := []string{
			base64.StdEncoding.EncodeToString([]byte("users:1")),
			base64.StdEncoding.EncodeToString([]byte("users:2")),
		}
		gb, err := graphQLQueryToRequestBody(`
query user($ids: [GID!]!) {
	nodes(ids: $ids) {
    	... on User {
			id
			name
	    }
    }
}
`, map[string]any{"ids": ids})
		s.Require().NoError(err)
		bd := strings.NewReader(gb)
		r := httptest.NewRequest("POST", "/graphql/query", bd)
		r.Header.Set("Content-Type", "application/json")
		srv.ServeHTTP(w, r)
		s.Contains(w.Body.String(), `{"id":"1","name":"user0"}`)
		s.Contains(w.Body.String(), `{"id":"2","name":"user1"}`)
	})
}

func (s *TestSuite) TestDecimal() {
	srv := handler.New(NewSchema(s.client))
	srv.AddTransport(transport.POST{})
	s.Run("Op", func() {
		s.NoError(s.client.User.Update().ClearMoney().Exec(context.Background()))
	})
	s.Run("normal", func() {
		w := httptest.NewRecorder()
		bd := strings.NewReader("{\"query\":\"mutation {\\n    createUser(name:\\\"test\\\",money: 11.123){\\n      name,money\\n  }\\n}\",\"variables\":{}}")
		r := httptest.NewRequest("POST", "/graphql/query", bd)
		r.Header.Set("Content-Type", "application/json")
		srv.ServeHTTP(w, r)
		s.Require().Equal(200, w.Code)
		var ret struct {
			Data struct {
				CreateUser ent.User
			}
		}
		s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &ret))
		s.Require().Equal("test", ret.Data.CreateUser.Name)
		s.Require().Equal(11.123, ret.Data.CreateUser.Money.InexactFloat64())
	})
	s.Run("decimal default", func() {
		w := httptest.NewRecorder()
		bd := strings.NewReader("{\"query\":\"mutation {\\n    createUser(name:\\\"test\\\"){\\n      name,money\\n  }\\n}\",\"variables\":{}}")
		r := httptest.NewRequest("POST", "/graphql/query", bd)
		r.Header.Set("Content-Type", "application/json")
		srv.ServeHTTP(w, r)
		s.Require().Equal(200, w.Code)
		var ret struct {
			Data struct {
				CreateUser ent.User
			}
		}
		s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &ret))
		s.Require().Equal(float64(2), ret.Data.CreateUser.Money.InexactFloat64())
		s.Run("decimal op", func() {
			err := s.client.User.Update().AddMoney(decimal.NewFromFloat(12.12)).Where(user.Name("test")).Exec(context.Background())
			s.NoError(err)
		})
	})
	s.Run("validate", func() {
		w := httptest.NewRecorder()
		bd := strings.NewReader("{\"query\":\"mutation {\\n    createUser(name:\\\"test\\\",money: 0.1123){\\n      name,money\\n  }\\n}\",\"variables\":{}}")
		r := httptest.NewRequest("POST", "/graphql/query", bd)
		r.Header.Set("Content-Type", "application/json")
		srv.ServeHTTP(w, r)
		s.Require().Equal(200, w.Code)
		var ret struct {
			Errors []struct {
				Message string
			}
		}
		s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &ret))
		s.Require().Contains(ret.Errors[0].Message, "value out of range")
	})
}

func (s *TestSuite) TestFile() {
	srv := handler.New(NewSchema(s.client))
	srv.AddTransport(transport.POST{})
	s.Run("ent", func() {
		s.NoError(s.client.User.Create().SetID(999).SetName("filetest").SetAvatar("test").Exec(context.Background()))
		err := s.client.User.UpdateOneID(999).SetAvatar(strings.Repeat("a", 256)).Exec(context.Background())
		s.True(ent.IsValidationError(err))
	})
}

func (s *TestSuite) TestResolverPlugin() {
	s.Run("CreateUser", func() {
		s.NotPanics(func() {
			_, _ = s.mutationResolver.CreateUser(ent.NewContext(context.Background(), s.mutationResolver.client), "empty", gds.Ptr(decimal.NewFromFloat(1.1)))
		})
	})
	s.Run("CreateUserByInput", func() {
		s.NotPanics(func() {
			_, _ = s.mutationResolver.CreateUserByInput(ent.NewContext(context.Background(), s.mutationResolver.client), ent.CreateUserInput{
				Name: "test",
			})
		})
	})
	s.Run("DeleteUser1", func() {
		s.NotPanics(func() {
			defer func() {
				e := recover()
				s.Equal(DeleteUser1Panic, e)
			}()
			_, _ = s.mutationResolver.DeleteUser1(ent.NewContext(context.Background(), s.mutationResolver.client), 1)
		})
	})
}
