package di_test

import (
	"errors"
	"testing"

	"github.com/HnH/di"
	"github.com/stretchr/testify/suite"
)

func TestContainerSuite(t *testing.T) {
	suite.Run(t, new(ContainerSuite))
}

type ContainerSuite struct {
	container di.Container
	resolver  di.Resolver

	suite.Suite
}

func (suite *ContainerSuite) SetupSuite() {
	suite.container = di.NewContainer()
	suite.resolver = di.NewResolver(suite.container)
}

func (suite *ContainerSuite) TearDownTest() {
	suite.container.Reset()
}

func (suite *ContainerSuite) TestSingleton() {
	suite.Require().NoError(suite.container.Singleton(newCircle))

	suite.Require().NoError(suite.resolver.Call(func(s1 Shape) {
		s1.SetArea(666)
	}))

	suite.Require().NoError(suite.resolver.Call(func(s2 Shape) {
		suite.Require().Equal(s2.GetArea(), 666)
	}))
}

func (suite *ContainerSuite) TestSingletonAlias() {
	suite.Require().NoError(suite.container.Singleton(func() Shape {
		return &Rectangle{a: 4444}
	}, di.WithName("kek", "bek")))

	var s Shape
	suite.Require().NoError(suite.resolver.Resolve(&s, di.WithName("kek")))
	suite.Require().IsType(&Rectangle{}, s)
	suite.Require().Equal(4444, s.GetArea())

	var s2 Shape
	suite.Require().NoError(suite.resolver.Resolve(&s2, di.WithName("bek")))
	suite.Require().IsType(&Rectangle{}, s2)
	suite.Require().Equal(4444, s2.GetArea())

	var s3 Shape
	suite.Require().EqualError(suite.resolver.Resolve(&s3), "di: no binding found for: di_test.Shape")
}

func (suite *ContainerSuite) TestSingletonMulti() {
	suite.Require().NoError(suite.container.Singleton(func() (Shape, Database, error) {
		return &Rectangle{a: 777}, &MySQL{}, nil
	}))

	var s Shape
	suite.Require().NoError(suite.resolver.Resolve(&s))
	suite.Require().IsType(&Rectangle{}, s)
	suite.Require().Equal(777, s.GetArea())

	var db Database
	suite.Require().NoError(suite.resolver.Resolve(&db))
	suite.Require().IsType(&MySQL{}, db)

	var err error
	suite.Require().EqualError(suite.resolver.Resolve(&err), "di: no binding found for: error")
}

func (suite *ContainerSuite) TestSingletonMultiNaming() {
	suite.Require().NoError(suite.container.Singleton(func() (Shape, Database, error) {
		return &Rectangle{a: 777}, &MySQL{}, nil
	}, di.WithName("kek", "bek")))

	var s Shape
	suite.Require().NoError(suite.resolver.Resolve(&s, di.WithName("kek")))
	suite.Require().IsType(&Rectangle{}, s)
	suite.Require().Equal(777, s.GetArea())

	var db Database
	suite.Require().NoError(suite.resolver.Resolve(&db, di.WithName("bek")))
	suite.Require().IsType(&MySQL{}, db)

	var err error
	suite.Require().EqualError(suite.resolver.Resolve(&err), "di: no binding found for: error")
}

func (suite *ContainerSuite) TestSingletonMultiNamingCountMismatch() {
	suite.Require().EqualError(
		suite.container.Singleton(func() (Shape, Database, error) {
			return &Rectangle{a: 777}, &MySQL{}, nil
		}, di.WithName("kek", "bek", "dek")),
		"di: the constructor that returns multiple values must be called with either one name or number of names equal to number of values",
	)
}

func (suite *ContainerSuite) TestSingletonBindError() {
	suite.Require().EqualError(suite.container.Singleton(func() (Shape, error) {
		return nil, errors.New("binding error")
	}), "binding error")
}

func (suite *ContainerSuite) TestSingletonResolverNotAFunc() {
	suite.Require().EqualError(suite.container.Singleton("STRING!"), "di: the constructor must be a function")
}

func (suite *ContainerSuite) TestSingletonResolverNumUsefulValues() {
	suite.Require().EqualError(suite.container.Factory(func() {}), "di: the constructor must return useful values")

	suite.Require().EqualError(suite.container.Factory(func() error {
		return errors.New("dummy error")
	}), "di: the constructor must return useful values")
}

func (suite *ContainerSuite) TestSingletonResolvableArgs() {
	suite.Require().NoError(suite.container.Singleton(newCircle))
	suite.Require().NoError(suite.container.Singleton(func(s Shape) Database {
		suite.Require().Equal(s.GetArea(), 100500)
		return &MySQL{}
	}))
}

func (suite *ContainerSuite) TestSingletonNonResolvableArgs() {
	suite.Require().EqualError(suite.container.Singleton(func(s Shape) Shape {
		return &Circle{a: s.GetArea()}
	}), "di: no binding found for: di_test.Shape")
}

func (suite *ContainerSuite) TestSingletonNamed() {
	suite.Require().NoError(suite.container.Singleton(func() Shape {
		return &Circle{a: 13}
	}, di.WithName("theCircle")))

	var sh Shape
	suite.Require().NoError(suite.resolver.Resolve(&sh, di.WithName("theCircle")))
	suite.Require().Equal(sh.GetArea(), 13)
}

func (suite *ContainerSuite) TestFactory() {
	suite.Require().NoError(suite.container.Factory(newCircle))

	suite.Require().NoError(suite.resolver.Call(func(s1 Shape) {
		s1.SetArea(13)
	}))

	suite.Require().NoError(suite.resolver.Call(func(s2 Shape) {
		suite.Require().Equal(s2.GetArea(), 100500)
	}))
}

func (suite *ContainerSuite) TestFactoryNamed() {
	suite.Require().NoError(suite.container.Factory(newCircle, di.WithName("theCircle")))

	var sh Shape
	suite.Require().NoError(suite.resolver.Resolve(&sh, di.WithName("theCircle")))
	suite.Require().Equal(sh.GetArea(), 100500)
}

func (suite *ContainerSuite) TestFactoryMultiError() {
	suite.Require().EqualError(suite.container.Factory(func() (Circle, Rectangle, Database) {
		return Circle{a: 666}, Rectangle{a: 666}, &MySQL{}
	}), "di: factory resolvers must return exactly one value and optionally one error")

	suite.Require().EqualError(suite.container.Factory(func() (Shape, Database) {
		return &Circle{a: 666}, &MySQL{}
	}), "di: factory resolvers must return exactly one value and optionally one error")

	suite.Require().NoError(suite.container.Factory(func() (Shape, error) {
		return nil, errors.New("dummy error")
	}))
}

func (suite *ContainerSuite) TestImplementation() {
	suite.Require().NoError(suite.container.Implementation(newCircle()))

	suite.Require().EqualError(suite.resolver.Call(func(s1 Shape) { return }), "di: no binding found for: di_test.Shape")
	suite.Require().NoError(suite.resolver.Call(func(s1 *Circle) { return }))
}

func (suite *ContainerSuite) TestImplementationWithoutName() {
	suite.Require().NoError(suite.container.Implementation(newCircle(), di.WithName("theCircle")))

	var c *Circle
	suite.Require().EqualError(suite.resolver.Resolve(&c), "di: no binding found for: *di_test.Circle")
	suite.Require().NoError(suite.resolver.Resolve(&c, di.WithName("theCircle")))
}

func (suite *ContainerSuite) TestCoverageBump() {
	suite.Require().NoError(di.Singleton(newCircle))
	suite.Require().NoError(di.Factory(newCircle))
	suite.Require().NoError(di.Implementation(newCircle()))
	suite.Require().NoError(di.Call(func(s Shape) { return }))
	suite.Require().NoError(di.With(newCircle()).Call(func(s Shape) { return }))

	var target Shape
	suite.Require().NoError(di.Resolve(&target))

	var list []Shape
	suite.Require().NoError(di.Fill(&list))
	di.Reset()
}
