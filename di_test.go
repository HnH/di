package di_test

import (
	"context"
	"time"
)

func newCircle() Shape {
	return &Circle{a: 100500}
}

func newRectangle() Shape {
	return &Rectangle{a: 255}
}

type Shape interface {
	SetArea(int)
	GetArea() int
}

type Circle struct {
	a int
}

func (c *Circle) SetArea(a int) {
	c.a = a
}

func (c Circle) GetArea() int {
	return c.a
}

type Rectangle struct {
	a int
}

func (s *Rectangle) SetArea(a int) {
	s.a = a
}

func (s Rectangle) GetArea() int {
	return s.a
}

func newMySQL() Database {
	return &MySQL{}
}

func newMongoDB(err error) Database {
	return &MongoDB{
		constructErr: err,
	}
}

type Database interface {
	Connect() bool
}

type MySQL struct{}

func (m MySQL) Connect() bool {
	return true
}

type MongoDB struct {
	Shape Shape `di:"type"`

	constructCalled time.Time
	constructErr    error
}

func (m *MongoDB) Construct(context.Context) error {
	m.constructCalled = time.Now()

	return m.constructErr
}

func (m *MongoDB) Connect() bool {
	return true
}
