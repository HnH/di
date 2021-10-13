package di_test

import (
	"errors"
	"reflect"
)

func (suite *ContainerSuite) TestCallMulti() {
	suite.Require().NoError(suite.container.Singleton(func() (Shape, Database, error) {
		return &Rectangle{a: 777}, &MySQL{}, nil
	}))

	suite.Require().NoError(suite.container.Call(func(s Shape, db Database) {
		suite.Require().IsType(&Rectangle{}, s)
		suite.Require().IsType(&MySQL{}, db)
	}))
}

func (suite *ContainerSuite) TestCallNotAFunc() {
	suite.Require().EqualError(suite.container.Call("STRING!"), "container: invalid function")
}

func (suite *ContainerSuite) TestCallUnboundArg() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))
	suite.Require().EqualError(suite.container.Call(func(s Shape, d Database) {}), "container: no concrete found for: di_test.Database")
}

func (suite *ContainerSuite) TestCallReturnedError() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))
	suite.Require().EqualError(suite.container.Call(func(s Shape) (err error) { return errors.New("dummy error") }), "dummy error")
}

func (suite *ContainerSuite) TestResolve() {
	suite.Require().NoError(suite.container.Singleton(suite.newRectangle))
	suite.Require().NoError(suite.container.Singleton(suite.newMySQL))

	var s Shape
	suite.Require().NoError(suite.container.Resolve(&s))
	suite.Require().IsType(&Rectangle{}, s)

	var db Database
	suite.Require().NoError(suite.container.Resolve(&db))
	suite.Require().IsType(&MySQL{}, db)
}

func (suite *ContainerSuite) TestResolveUnsupportedReceiver() {
	suite.Require().EqualError(suite.container.Resolve("STRING!"), "container: invalid abstraction")
}

func (suite *ContainerSuite) TestResolveReceiverNotAPointer() {
	var s Shape
	suite.Require().EqualError(suite.container.Resolve(s), "container: invalid abstraction")
}

func (suite *ContainerSuite) TestResolveUnbound() {
	var s Shape
	suite.Require().EqualError(suite.container.Resolve(&s), "container: no concrete found for: di_test.Shape")
}

func (suite *ContainerSuite) TestResolveInvokeError() {
	suite.Require().NoError(suite.container.Transient(func() (Shape, error) {
		return nil, errors.New("dummy error")
	}))

	var s Shape
	suite.Require().EqualError(suite.container.Resolve(&s), "dummy error")
}

func (suite *ContainerSuite) TestFillStruct() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))
	suite.Require().NoError(suite.container.NamedSingleton("R", suite.newRectangle))
	suite.Require().NoError(suite.container.Singleton(suite.newMySQL))

	var target = struct {
		S Shape    `container:"type"`
		D Database `container:"type"`
		R Shape    `container:"name"`
		X string
		y int
	}{
		X: "dummy string",
		y: 100,
	}

	suite.Require().NoError(suite.container.Fill(&target))

	suite.Require().IsType(&Circle{}, target.S)
	suite.Require().IsType(&MySQL{}, target.D)
	suite.Require().IsType(&Rectangle{}, target.R)
	suite.Require().Equal(target.X, "dummy string")
	suite.Require().Equal(target.y, 100)
}

func (suite *ContainerSuite) TestFillSlice() {
	suite.Require().NoError(suite.container.NamedSingleton("circle", suite.newCircle))
	suite.Require().NoError(suite.container.NamedSingleton("square", suite.newRectangle))

	var shapes []Shape
	suite.Require().NoError(suite.container.Fill(&shapes))
	suite.Require().Equal(2, len(shapes))

	var list = map[string]struct{}{
		reflect.TypeOf(shapes[0]).Elem().Name(): {},
		reflect.TypeOf(shapes[1]).Elem().Name(): {},
	}

	_, ok := list["Circle"]
	suite.Require().True(ok)

	_, ok = list["Rectangle"]
	suite.Require().True(ok)
}

func (suite *ContainerSuite) TestFillMap() {
	suite.Require().NoError(suite.container.NamedSingleton("circle", suite.newCircle))
	suite.Require().NoError(suite.container.NamedSingleton("square", suite.newRectangle))

	var shapes map[string]Shape
	suite.Require().NoError(suite.container.Fill(&shapes))
	suite.Require().Equal(2, len(shapes))

	_, ok := shapes["circle"]
	suite.Require().True(ok)
	suite.Require().IsType(&Circle{}, shapes["circle"])

	_, ok = shapes["square"]
	suite.Require().True(ok)
	suite.Require().IsType(&Rectangle{}, shapes["square"])

}

func (suite *ContainerSuite) TestFillReceiverInvalid() {
	var target = 0
	suite.Require().EqualError(suite.container.Fill(&target), "container: invalid receiver")
}

func (suite *ContainerSuite) TestFillReceiverNil() {
	var target Shape
	suite.Require().EqualError(suite.container.Fill(target), "container: invalid receiver")
}

func (suite *ContainerSuite) TestFillReceiverNotAPointer() {
	var db MySQL
	suite.Require().EqualError(suite.container.Fill(db), "container: receiver is not a pointer")
}

func (suite *ContainerSuite) TestFillInvalidName() {
	suite.Require().NoError(suite.container.NamedSingleton("R", suite.newRectangle))
	var target = struct {
		S Shape `container:"name"`
	}{}

	suite.Require().EqualError(suite.container.Fill(&target), "container: cannot resolve S field")
}

func (suite *ContainerSuite) TestFillInvalidTag() {
	suite.Require().NoError(suite.container.NamedSingleton("R", suite.newRectangle))
	var target = struct {
		S Shape `container:"invalid"`
	}{}

	suite.Require().EqualError(suite.container.Fill(&target), "container: S has an invalid struct tag")
}

func (suite *ContainerSuite) TestFillInvalidMap() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))

	var list = map[int]Shape{}
	suite.Require().EqualError(suite.container.Fill(&list), "container: invalid receiver")
}
