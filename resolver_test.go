package di_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/HnH/di"
	"github.com/stretchr/testify/suite"
)

func TestResolverSuite(t *testing.T) {
	suite.Run(t, new(ResolverSuite))
}

type ResolverSuite struct {
	container di.Container
	resolver  di.Resolver

	suite.Suite
}

func (suite *ResolverSuite) SetupSuite() {
	suite.container = di.NewContainer()
	suite.resolver = di.NewResolver(suite.container)
}

func (suite *ResolverSuite) TearDownTest() {
	suite.container.Reset()
}

func (suite *ResolverSuite) TestNewResolver() {
	var rsl = di.NewResolver()
	suite.Require().NotNil(rsl)
	suite.Require().EqualError(di.Call(context.Background(), func(s Shape) { return }), "di: no binding found for di_test.Shape")
}

func (suite *ResolverSuite) TestCallMulti() {
	suite.Require().NoError(suite.container.Singleton(func() (Shape, Database, error) {
		return &Rectangle{a: 777}, &MySQL{}, nil
	}))

	suite.Require().NoError(suite.resolver.Call(func(s Shape, db Database) {
		suite.Require().IsType(&Rectangle{}, s)
		suite.Require().IsType(&MySQL{}, db)
	}))
}

func (suite *ResolverSuite) TestCallReturn() {
	suite.Require().NoError(suite.container.Singleton(newCircle))

	var (
		str string
		db  Database
	)

	suite.Require().NoError(suite.resolver.Call(func(s Shape) (string, Database, error) {
		return "mysql", &MySQL{}, nil
	}, di.WithReturn(&str, &db)))

	suite.Require().Equal("mysql", str)
	suite.Require().IsType(&MySQL{}, db)
}

func (suite *ResolverSuite) TestCallWith() {
	suite.Require().NoError(suite.container.Singleton(newCircle))

	var db = newMySQL()

	suite.Require().EqualError(suite.resolver.Call(func(s Shape, db Database) { return }), "di: no binding found for di_test.Database")
	suite.Require().NoError(suite.resolver.With(db).Call(func(s Shape, db Database) { return }))
}

func (suite *ResolverSuite) TestCallImplementationWithDiff() {
	var circle = newCircle()
	suite.Require().NoError(suite.container.Implementation(circle))

	suite.Require().EqualError(suite.resolver.Call(func(s Shape) { return }), "di: no binding found for di_test.Shape")
	suite.Require().NoError(suite.resolver.With(circle).Call(func(s Shape) { return }))
}

func (suite *ResolverSuite) TestCallNotAFunc() {
	suite.Require().EqualError(suite.resolver.Call("STRING!"), "di: invalid function")
}

func (suite *ResolverSuite) TestCallUnboundArg() {
	suite.Require().NoError(suite.container.Singleton(newCircle))
	suite.Require().EqualError(suite.resolver.Call(func(s Shape, d Database) {}), "di: no binding found for di_test.Database")
}

func (suite *ResolverSuite) TestCallReceiverNumMismatch() {
	suite.Require().NoError(suite.container.Singleton(newCircle))

	var str string
	suite.Require().EqualError(suite.resolver.Call(func(s Shape) (string, Database) {
		return "mysql", &MySQL{}
	}, di.WithReturn(&str)), "di: cannot assign 2 returned values to 1 receivers")

	var err error
	suite.Require().EqualError(suite.resolver.Call(func(s Shape) (string, error) {
		return "mysql", nil
	}, di.WithReturn(&str, &err)), "di: cannot assign 1 returned values to 2 receivers")
}

func (suite *ResolverSuite) TestCallReceiverTypeMismatch() {
	suite.Require().NoError(suite.container.Singleton(newCircle))

	var str, db string
	suite.Require().EqualError(suite.resolver.Call(func(s Shape) (string, Database) {
		return "mysql", &MySQL{}
	}, di.WithReturn(&str, &db)), "di: cannot assign returned value of type Database to string")
}

func (suite *ResolverSuite) TestCallReturnedError() {
	suite.Require().NoError(suite.container.Singleton(newCircle))
	suite.Require().EqualError(suite.resolver.Call(func(s Shape) (err error) { return errors.New("dummy error") }), "dummy error")
}

func (suite *ResolverSuite) TestResolve() {
	suite.Require().NoError(suite.container.Singleton(newRectangle))
	suite.Require().NoError(suite.container.Singleton(newMySQL))

	var s Shape
	suite.Require().NoError(suite.resolver.Resolve(&s))
	suite.Require().IsType(&Rectangle{}, s)

	var db Database
	suite.Require().NoError(suite.resolver.Resolve(&db))
	suite.Require().IsType(&MySQL{}, db)
}

func (suite *ResolverSuite) TestResolveMultiContainer() {
	var (
		localContainer = di.NewContainer()
		localResolver  = di.NewResolver(localContainer, suite.container)
	)

	suite.Require().NoError(suite.container.Singleton(newRectangle))
	suite.Require().NoError(localContainer.Singleton(newMySQL))

	var s Shape
	suite.Require().NoError(localResolver.Resolve(&s))
	suite.Require().IsType(&Rectangle{}, s)

	var db Database
	suite.Require().EqualError(suite.resolver.Resolve(&db), "di: no binding found for di_test.Database")
	suite.Require().Nil(db)

	suite.Require().NoError(localResolver.Resolve(&db))
	suite.Require().IsType(&MySQL{}, db)
}

func (suite *ResolverSuite) TestResolveUnsupportedReceiver() {
	suite.Require().EqualError(suite.resolver.Resolve("STRING!"), "di: invalid receiver")
}

func (suite *ResolverSuite) TestResolveReceiverNotAPointer() {
	var s Shape
	suite.Require().EqualError(suite.resolver.Resolve(s), "di: invalid receiver")
}

func (suite *ResolverSuite) TestResolveUnbound() {
	var s Shape
	suite.Require().EqualError(suite.resolver.Resolve(&s), "di: no binding found for di_test.Shape")
}

func (suite *ResolverSuite) TestResolveInvokeError() {
	suite.Require().NoError(suite.container.Factory(func() (Shape, error) {
		return nil, errors.New("dummy error")
	}))

	var s Shape
	suite.Require().EqualError(suite.resolver.Resolve(&s), "dummy error")
}

func (suite *ResolverSuite) TestFillStruct() {
	suite.Require().NoError(suite.container.Singleton(newCircle))
	suite.Require().NoError(suite.container.Singleton(newRectangle, di.WithName("R")))
	suite.Require().NoError(suite.container.Singleton(newMySQL))

	var target = struct {
		S Shape    `di:"type"`
		D Database `di:"type"`
		R Shape    `di:"name"`
		U *struct {
			U Shape `di:"type"`
		} `di:"recursive"`
		X string
		y int
	}{
		U: &struct {
			U Shape `di:"type"`
		}{},
		X: "dummy string",
		y: 100,
	}

	suite.Require().NoError(suite.resolver.Fill(&target))

	suite.Require().IsType(&Circle{}, target.S)
	suite.Require().IsType(&MySQL{}, target.D)
	suite.Require().IsType(&Rectangle{}, target.R)
	suite.Require().IsType(&Circle{}, target.U.U)
	suite.Require().Equal(target.X, "dummy string")
	suite.Require().Equal(target.y, 100)
}

func (suite *ResolverSuite) TestFillStructOmitempty() {
	suite.Require().NoError(suite.container.Singleton(newMySQL))

	var target = struct {
		S Shape    `di:"type"`
		D Database `di:"type"`
	}{}

	suite.Require().Error(suite.resolver.Fill(&target))

	var optionalTarget = struct {
		S Shape    `di:"type,omitempty"`
		D Database `di:"type"`
	}{}

	suite.Require().NoError(suite.resolver.Fill(&optionalTarget))
	suite.Require().Nil(optionalTarget.S)
	suite.Require().IsType(&MySQL{}, optionalTarget.D)
}

func (suite *ResolverSuite) TestFillSlice() {
	suite.Require().NoError(suite.container.Singleton(newCircle, di.WithName("circle")))
	suite.Require().NoError(suite.container.Singleton(newRectangle, di.WithName("square")))

	var shapes []Shape
	suite.Require().NoError(suite.resolver.Fill(&shapes))
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

func (suite *ResolverSuite) TestFillMap() {
	suite.Require().NoError(suite.container.Singleton(newCircle, di.WithName("circle")))
	suite.Require().NoError(suite.container.Singleton(newRectangle, di.WithName("square")))

	var shapes map[string]Shape
	suite.Require().NoError(suite.resolver.Fill(&shapes))
	suite.Require().Equal(2, len(shapes))

	_, ok := shapes["circle"]
	suite.Require().True(ok)
	suite.Require().IsType(&Circle{}, shapes["circle"])

	_, ok = shapes["square"]
	suite.Require().True(ok)
	suite.Require().IsType(&Rectangle{}, shapes["square"])
}

func (suite *ResolverSuite) TestFillStructWithSliceMap() {
	suite.Require().NoError(suite.container.Singleton(newCircle, di.WithName("circle")))
	suite.Require().NoError(suite.container.Singleton(newRectangle, di.WithName("square")))

	var shapes struct {
		list []Shape          `di:"recursive"`
		dict map[string]Shape `di:"recursive"`
	}

	suite.Require().NoError(suite.resolver.Fill(&shapes))
	suite.Require().Equal(2, len(shapes.list))
	suite.Require().Equal(2, len(shapes.dict))

	var list = map[string]struct{}{
		reflect.TypeOf(shapes.list[0]).Elem().Name(): {},
		reflect.TypeOf(shapes.list[1]).Elem().Name(): {},
	}

	_, ok := list["Circle"]
	suite.Require().True(ok)

	_, ok = list["Rectangle"]
	suite.Require().True(ok)

	var dict = map[string]struct{}{
		reflect.TypeOf(shapes.dict["circle"]).Elem().Name(): {},
		reflect.TypeOf(shapes.dict["square"]).Elem().Name(): {},
	}

	_, ok = dict["Circle"]
	suite.Require().True(ok)

	_, ok = dict["Rectangle"]
	suite.Require().True(ok)
}

func (suite *ResolverSuite) TestFillReceiverInvalid() {
	var target = 0
	suite.Require().EqualError(suite.resolver.Fill(&target), "di: invalid receiver: *int: filling *int")
}

func (suite *ResolverSuite) TestFillReceiverNil() {
	var target Shape
	suite.Require().EqualError(suite.resolver.Fill(target), "di: invalid receiver: nil")
}

func (suite *ResolverSuite) TestFillReceiverNotAPointer() {
	var db MySQL
	suite.Require().EqualError(suite.resolver.Fill(db), "di: receiver is not a pointer: struct")
}

func (suite *ResolverSuite) TestFillInvalidName() {
	suite.Require().NoError(suite.container.Singleton(newRectangle, di.WithName("R")))
	var target = struct {
		S Shape `di:"name"`
	}{}

	suite.Require().EqualError(suite.resolver.Fill(&target), `di: no binding found for di_test.Shape: filling *struct { S di_test.Shape "di:\"name\"" }`)
}

func (suite *ResolverSuite) TestFillInvalidTag() {
	suite.Require().NoError(suite.container.Singleton(newRectangle, di.WithName("R")))
	var target = struct {
		S Shape `di:"invalid"`
	}{}

	suite.Require().EqualError(suite.resolver.Fill(&target), `di: S has an invalid struct tag: filling *struct { S di_test.Shape "di:\"invalid\"" }`)
}

func (suite *ResolverSuite) TestFillRecursiveStruct() {
	suite.Require().NoError(suite.container.Singleton(newRectangle))
	var target = struct {
		inner struct {
			S Shape `di:"type"`
		} `di:"recursive"`
	}{}

	suite.Require().NoError(suite.resolver.Fill(&target))
	suite.Require().Equal(newRectangle().GetArea(), target.inner.S.GetArea())
}

func (suite *ResolverSuite) TestFillSliceUnbound() {
	var list []Shape
	suite.Require().EqualError(suite.resolver.Fill(&list), "di: no binding found for di_test.Shape: filling *[]di_test.Shape")
}

func (suite *ResolverSuite) TestFillSliceFactoryError() {
	suite.Require().NoError(suite.container.Singleton(context.Background))
	suite.Require().NoError(suite.container.Factory(func() Database { return newMongoDB(errors.New("dummy error")) }))

	var list []Database
	suite.Require().EqualError(suite.resolver.Fill(&list), "dummy error: filling *[]di_test.Database")
}

func (suite *ResolverSuite) TestFillInvalidMap() {
	suite.Require().NoError(suite.container.Singleton(newCircle))

	var list = map[int]Shape{}
	suite.Require().EqualError(suite.resolver.Fill(&list), "di: invalid receiver: *map[int]di_test.Shape: filling *map[int]di_test.Shape")
}

func (suite *ResolverSuite) TestFillMapUnbound() {
	var list map[string]Shape
	suite.Require().EqualError(suite.resolver.Fill(&list), "di: no binding found for di_test.Shape: filling *map[string]di_test.Shape")
}

func (suite *ResolverSuite) TestFillMapFactoryError() {
	suite.Require().NoError(suite.container.Singleton(context.Background))
	suite.Require().NoError(suite.container.Factory(func() Database { return newMongoDB(errors.New("dummy error")) }))

	var list map[string]Database
	suite.Require().EqualError(suite.resolver.Fill(&list), "dummy error: filling *map[string]di_test.Database")
}
