package di_test

import (
	"errors"
	"reflect"

	"github.com/HnH/di"
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
	suite.Require().EqualError(suite.container.Call("STRING!"), "di: invalid function")
}

func (suite *ContainerSuite) TestCallUnboundArg() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))
	suite.Require().EqualError(suite.container.Call(func(s Shape, d Database) {}), "di: no binding found for: di_test.Database")
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
	suite.Require().EqualError(suite.container.Resolve("STRING!"), "di: invalid abstraction")
}

func (suite *ContainerSuite) TestResolveReceiverNotAPointer() {
	var s Shape
	suite.Require().EqualError(suite.container.Resolve(s), "di: invalid abstraction")
}

func (suite *ContainerSuite) TestResolveUnbound() {
	var s Shape
	suite.Require().EqualError(suite.container.Resolve(&s), "di: no binding found for: di_test.Shape")
}

func (suite *ContainerSuite) TestResolveInvokeError() {
	suite.Require().NoError(suite.container.Factory(func() (Shape, error) {
		return nil, errors.New("dummy error")
	}))

	var s Shape
	suite.Require().EqualError(suite.container.Resolve(&s), "dummy error")
}

func (suite *ContainerSuite) TestFillStruct() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))
	suite.Require().NoError(suite.container.Singleton(suite.newRectangle, di.WithName("R")))
	suite.Require().NoError(suite.container.Singleton(suite.newMySQL))

	var target = struct {
		S Shape    `di:"type"`
		D Database `di:"type"`
		R Shape    `di:"name"`
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
	suite.Require().NoError(suite.container.Singleton(suite.newCircle, di.WithName("circle")))
	suite.Require().NoError(suite.container.Singleton(suite.newRectangle, di.WithName("square")))

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
	suite.Require().NoError(suite.container.Singleton(suite.newCircle, di.WithName("circle")))
	suite.Require().NoError(suite.container.Singleton(suite.newRectangle, di.WithName("square")))

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
	suite.Require().EqualError(suite.container.Fill(&target), "di: invalid receiver")
}

func (suite *ContainerSuite) TestFillReceiverNil() {
	var target Shape
	suite.Require().EqualError(suite.container.Fill(target), "di: invalid receiver")
}

func (suite *ContainerSuite) TestFillReceiverNotAPointer() {
	var db MySQL
	suite.Require().EqualError(suite.container.Fill(db), "di: receiver is not a pointer")
}

func (suite *ContainerSuite) TestFillInvalidName() {
	suite.Require().NoError(suite.container.Singleton(suite.newRectangle, di.WithName("R")))
	var target = struct {
		S Shape `di:"name"`
	}{}

	suite.Require().EqualError(suite.container.Fill(&target), "di: no binding found for: Shape")
}

func (suite *ContainerSuite) TestFillInvalidTag() {
	suite.Require().NoError(suite.container.Singleton(suite.newRectangle, di.WithName("R")))
	var target = struct {
		S Shape `di:"invalid"`
	}{}

	suite.Require().EqualError(suite.container.Fill(&target), "di: S has an invalid struct tag")
}

func (suite *ContainerSuite) TestFillSliceUnbound() {
	var list []Shape
	suite.Require().EqualError(suite.container.Fill(&list), "di: no binding found for: di_test.Shape")
}

func (suite *ContainerSuite) TestFillInvalidMap() {
	suite.Require().NoError(suite.container.Singleton(suite.newCircle))

	var list = map[int]Shape{}
	suite.Require().EqualError(suite.container.Fill(&list), "di: invalid receiver")
}

func (suite *ContainerSuite) TestFillMapUnbound() {
	var list map[string]Shape
	suite.Require().EqualError(suite.container.Fill(&list), "di: no binding found for: di_test.Shape")
}
