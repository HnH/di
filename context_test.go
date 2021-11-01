package di_test

import (
	"context"
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
	suite.context = di.Ctx(context.Background()).Put(di.NewContainer())
}

func (suite *ContextSuite) TearDownTest() {
	suite.context.Container().Reset()
}

func (suite *ContextSuite) TestPut() {
	var ctx = context.Background()
	suite.Require().NotNil(di.Ctx(ctx).Put(di.NewContainer()))
}

func (suite *ContextSuite) TestDefaultContainer() {
	var (
		container = di.Ctx(context.Background()).Container()
		shape     Shape
	)

	suite.Require().EqualError(di.NewResolver(container).Resolve(&shape), "di: no binding found for: di_test.Shape")
}

func (suite *ContextSuite) TestResolve() {
	suite.Require().NoError(suite.context.Container().Singleton(newCircle))

	var shape Shape
	suite.Require().NoError(suite.context.Resolver().Resolve(&shape))
	suite.Require().IsType(&Circle{}, shape)

	suite.context.Container().Reset()
	suite.Require().EqualError(suite.context.Resolver().Resolve(&shape), "di: no binding found for: di_test.Shape")
}

func (suite *ContextSuite) TestRaw() {
	suite.Require().NotNil(suite.context.Raw())
}
