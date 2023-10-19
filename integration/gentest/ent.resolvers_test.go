package gentest

import (
	"context"
	"encoding/json"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"github.com/tsingsun/woocoo/pkg/gds"
	"github.com/woocoos/knockout-go/integration/gentest/ent"
	"github.com/woocoos/knockout-go/integration/gentest/ent/user"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/woocoos/knockout-go/integration/gentest/ent/runtime"
)

type TestSuite struct {
	suite.Suite
	client        *ent.Client
	queryResolver queryResolver
}

func (s *TestSuite) SetupSuite() {
	dr, err := sql.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	s.Require().NoError(err)
	s.client = ent.NewClient(ent.Driver(dr), ent.Debug())
	s.queryResolver = queryResolver{&Resolver{s.client}}
	s.NoError(s.client.Schema.Create(context.Background()))
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestUsers_SimplePagination() {
	builder := make([]*ent.UserCreate, 0)
	for i := 0; i < 20; i++ {
		builder = append(builder, s.client.User.Create().SetName("user"+strconv.Itoa(i)))
	}
	s.NoError(s.client.User.CreateBulk(builder...).Exec(context.Background()))

	s.Run("pagination after", func() {
		ctx, _ := gin.CreateTestContext(nil)
		users, err := s.queryResolver.Users(ctx, &ent.Cursor{ID: 2}, gds.Ptr(2), nil, nil, nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(3, users.Edges[0].Node.ID)
		s.Equal(4, users.Edges[1].Node.ID)

	})

	s.Run("simple after", func() {
		ctx, _ := gin.CreateTestContext(nil)
		ctx.Request = httptest.NewRequest("GET", "/?p=3&c=1", nil)
		users, err := s.queryResolver.Users(ctx, &ent.Cursor{ID: 2}, gds.Ptr(2), nil, nil, nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(5, users.Edges[0].Node.ID)
		s.Equal(6, users.Edges[1].Node.ID)

	})

	s.Run("pagination befor", func() {
		ctx, _ := gin.CreateTestContext(nil)
		users, err := s.queryResolver.Users(ctx, nil, nil, &ent.Cursor{ID: 5}, gds.Ptr(2), nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(3, users.Edges[0].Node.ID)
		s.Equal(4, users.Edges[1].Node.ID)

	})

	s.Run("simple before", func() {
		ctx, _ := gin.CreateTestContext(nil)
		ctx.Request = httptest.NewRequest("GET", "/?p=1&c=3", nil)
		users, err := s.queryResolver.Users(ctx, nil, nil, &ent.Cursor{ID: 5}, gds.Ptr(2), nil, nil)
		s.NoError(err)
		s.Len(users.Edges, 2)
		s.Equal(1, users.Edges[0].Node.ID)
		s.Equal(2, users.Edges[1].Node.ID)
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
