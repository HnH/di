package di_test

import (
	"context"
	"strings"
	"testing"

	"github.com/HnH/di"
	"github.com/stretchr/testify/suite"
)

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextSuite))
}

type ContextSuite struct {
	context di.Context

	suite.Suite
}

func (suite *ContextSuite) SetupSuite() {
	suite.context = di.Ctx(context.Background()).SetContainer(di.NewContainer())
}

func (suite *ContextSuite) TearDownTest() {
	suite.context.Container().Reset()
}

func (suite *ContextSuite) TestSetContainer() {
	var ctx = context.Background()
	suite.Require().NotNil(di.Ctx(ctx).SetContainer(di.NewContainer()).Raw())
}

func (suite *ContextSuite) TestSetResolver() {
	var (
		ctx = di.Ctx(context.Background())
		rsl = di.NewResolver(di.NewContainer())
	)

	suite.Require().NotNil(ctx.SetResolver(rsl).Raw())
	suite.Require().Equal(rsl, ctx.Resolver())
}

func (suite *ContextSuite) TestDefaultContainer() {
	var (
		container = di.Ctx(context.Background()).Container()
		shape     Shape
	)

	suite.Require().EqualError(di.NewResolver(container).Resolve(&shape), "di: no binding found for di_test.Shape")
	suite.Require().EqualError(di.Ctx(context.Background()).SetContainer(container).Resolver().Resolve(&shape), "di: no binding found for di_test.Shape")
}

func (suite *ContextSuite) TestResolve() {
	suite.Require().NoError(suite.context.Container().Singleton(newCircle))

	var shape Shape
	suite.Require().NoError(suite.context.Resolver().Resolve(&shape))
	suite.Require().IsType(&Circle{}, shape)

	suite.context.Container().Reset()
	suite.Require().EqualError(suite.context.Resolver().Resolve(&shape), "di: no binding found for di_test.Shape")
}

func (suite *ContextSuite) TestVisualize() {
	suite.Require().Equal([]string{
		"resolver has [1] containers",
		"  -> container [0] has [0] type binding(s)",
	}, suite.context.Visualize())

	suite.context.Container().Singleton(newCircle)
	suite.context.Container().Factory(newMySQL)
	var out = suite.context.Visualize()

	suite.Require().Equal("resolver has [1] containers", out[0])
	suite.Require().Equal("  -> container [0] has [2] type binding(s)", out[1])
	suite.Require().Equal("    -> [di_test.Shape] has [1] binding(s)", out[2])
	suite.Require().True(strings.Contains(out[3], "di/context_test.go:72"))
	suite.Require().Equal("    -> [di_test.Database] has [1] binding(s)", out[4])
	suite.Require().True(strings.Contains(out[5], "di/context_test.go:73"))
}

func (suite *ContextSuite) TestRaw() {
	suite.Require().NotNil(suite.context.Raw())
}
