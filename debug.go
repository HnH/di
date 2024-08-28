package di

import "fmt"

type Debuggable interface {
	Visualize() []string
}

func (self *container) Visualize() []string {
	self.lock.RLock()
	defer self.lock.RUnlock()

	var out = make([]string, 0, len(self.bindings)*2)
	for t, m := range self.bindings {
		out = append(out, t.String())

		for k, v := range m {
			out = append(out, fmt.Sprintf("%s=%v", k, v))
		}
	}

	return nil
}

// Debug returns detailed information on underlying structure of Container/Resolver
func Debug[T Debuggable](in T) []string {
	// just a container
	if c, ok := any(in).(*container); ok {
		return debug(c)
	}

	var out []string
	if r, ok := any(in).(*resolver); ok {
		for i := range r.containers {
			if c, ok := r.containers[i].(*container); ok {
				out = append(out, debug(c)...)
			}
		}
	}

	return nil
}

func debug(c *container) []string {
	var out = make([]string, 0, len(c.bindings)*2)
	for t, m := range c.bindings {
		out = append(out, t.String())

		for k, v := range m {
			out = append(out, fmt.Sprintf("%s=%v", k, v))
		}
	}

	return out
}
