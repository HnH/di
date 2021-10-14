package di_test

import (
	"testing"

	"github.com/HnH/di"
	"github.com/stretchr/testify/suite"
)

type ContainerSuite struct {
	container di.Container
	resolver  di.Resolver

	suite.Suite
}

func (suite *ContainerSuite) SetupSuite() {
	suite.container = di.New()
	suite.resolver, _ = di.NewResolver(suite.container)
}

func (suite *ContainerSuite) TearDownTest() {
	suite.container.Reset()
}

func (suite *ContainerSuite) newCircle() Shape {
	return &Circle{a: 100500}
}

func (suite *ContainerSuite) newRectangle() Shape {
	return &Rectangle{a: 255}
}

func (suite *ContainerSuite) newMySQL() Database {
	return &MySQL{}
}

func (suite *ContainerSuite) TestCoverageBump() {
	suite.Require().NoError(di.Singleton(suite.newCircle))
	suite.Require().NoError(di.Factory(suite.newCircle))
	suite.Require().NoError(di.Call(func(s Shape) { return }))

	var target Shape
	suite.Require().NoError(di.Resolve(&target))

	var list []Shape
	suite.Require().NoError(di.Fill(&list))
	di.Reset()
}

func TestContainerSuite(t *testing.T) {
	suite.Run(t, new(ContainerSuite))
}
