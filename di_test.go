package di_test

func newCircle() Shape {
	return &Circle{a: 100500}
}

func newRectangle() Shape {
	return &Rectangle{a: 255}
}

func newMySQL() Database {
	return &MySQL{}
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

type Database interface {
	Connect() bool
}

type MySQL struct{}

func (m MySQL) Connect() bool {
	return true
}
