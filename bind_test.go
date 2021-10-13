package di_test

import (
	"errors"
)

func (suite *ContainerSuite) TestSingleton() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))
	suite.Require().NoError(suite.container.Singleton(func() {}))

	// suite.container equals?
	suite.Require().NoError(suite.container.Call(func(s1 Shape) {
		s1.SetArea(666)
	}))
	suite.Require().NoError(suite.container.Call(func(s2 Shape) {
		suite.Require().Equal(s2.GetArea(), 666)
	}))
}

func (suite *ContainerSuite) TestSingletonMulti() {
	suite.Require().NoError(suite.container.Singleton(func() (Shape, Database, error) {
		return &Rectangle{a: 777}, &MySQL{}, nil
	}))

	var s Shape
	suite.Require().NoError(suite.container.Resolve(&s))
	suite.Require().IsType(&Rectangle{}, s)
	suite.Require().Equal(777, s.GetArea())

	var db Database
	suite.Require().NoError(suite.container.Resolve(&db))
	suite.Require().IsType(&MySQL{}, db)

	var err error
	suite.Require().EqualError(suite.container.Resolve(&err), "container: no concrete found for: error")
}

func (suite *ContainerSuite) TestSingletonBindError() {
	suite.Require().EqualError(suite.container.Singleton(func() (Shape, error) {
		return nil, errors.New("binding error")
	}), "binding error")
}

func (suite *ContainerSuite) TestSingletonResolverNotAFunc() {
	suite.Require().EqualError(suite.container.Singleton("STRING!"), "container: the resolver must be a function")
}

func (suite *ContainerSuite) TestSingletonResolvableArgs() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))
	suite.Require().NoError(suite.container.Singleton(func(s Shape) Database {
		suite.Require().Equal(s.GetArea(), 100500)
		return &MySQL{}
	}))
}

func (suite *ContainerSuite) TestSingletonNonResolvableArgs() {
	suite.Require().EqualError(suite.container.Singleton(func(s Shape) Shape {
		return &Circle{a: s.GetArea()}
	}), "container: no concrete found for: di_test.Shape")
}

func (suite *ContainerSuite) TestSingletonNamed() {
	suite.Require().NoError(suite.container.NamedSingleton("theCircle", func() Shape {
		return &Circle{a: 13}
	}))

	var sh Shape
	suite.Require().NoError(suite.container.NamedResolve(&sh, "theCircle"))
	suite.Require().Equal(sh.GetArea(), 13)
}

func (suite *ContainerSuite) TestFactory() {
	suite.Require().NoError(suite.container.Transient(suite.newCircle))

	suite.Require().NoError(suite.container.Call(func(s1 Shape) {
		s1.SetArea(13)
	}))

	suite.Require().NoError(suite.container.Call(func(s2 Shape) {
		suite.Require().Equal(s2.GetArea(), 100500)
	}))
}

func (suite *ContainerSuite) TestFactoryNamed() {
	suite.Require().NoError(suite.container.NamedTransient("theCircle", suite.newCircle))

	var sh Shape
	suite.Require().NoError(suite.container.NamedResolve(&sh, "theCircle"))
	suite.Require().Equal(sh.GetArea(), 100500)
}

func (suite *ContainerSuite) TestFactoryMultiError() {
	suite.Require().EqualError(suite.container.Transient(func() (Circle, Rectangle, Database) {
		return Circle{a: 666}, Rectangle{a: 666}, &MySQL{}
	}), "container: transient value resolvers must return exactly one value and optionally one error")

	suite.Require().EqualError(suite.container.Transient(func() (Shape, Database) {
		return &Circle{a: 666}, &MySQL{}
	}), "container: transient value resolvers must return exactly one value and optionally one error")

	suite.Require().EqualError(suite.container.Transient(func() error {
		return errors.New("dummy error")
	}), "container: transient value resolvers must return exactly one value and optionally one error")

	suite.Require().NoError(suite.container.Transient(func() (Shape, error) {
		return nil, errors.New("dummy error")
	}))
}
