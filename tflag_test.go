package tflag

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type slot struct {
	value interface{}
	usage string
}

type mock struct {
	data map[string]slot
}

func (m *mock) Var(p flag.Value, name string, usage string) {}

func (m *mock) IntVar(p *int, name string, value int, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func (m *mock) BoolVar(p *bool, name string, value bool, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func (m *mock) UintVar(p *uint, name string, value uint, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func (m *mock) Int64Var(p *int64, name string, value int64, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func (m *mock) StringVar(p *string, name string, value string, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func (m *mock) Uint64Var(p *uint64, name string, value uint64, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func (m *mock) Float64Var(p *float64, name string, value float64, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func (m *mock) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	m.data[name] = slot{usage: usage, value: value}
}

func newMock() *mock {
	return &mock{data: make(map[string]slot)}
}

type Config1 struct {
	IntVar      int           `flag:"intvar" usage:"intvar usage" default:"-10"`
	BoolVar     bool          `flag:"boolvar" usage:"boolvar usage" default:"true"`
	UintVar     uint          `flag:"uintvar" usage:"uintvar usage" default:"10"`
	Int64Var    int64         `flag:"int64var" usage:"int64var usage" default:"-10"`
	StringVar   string        `flag:"stringvar" usage:"stringvar usage" default:"hello"`
	Uint64Var   uint64        `flag:"uint64var" usage:"uint64var usage" default:"10"`
	Float64Var  float64       `flag:"float64var" usage:"float64var usage" default:"22.33"`
	DurationVar time.Duration `flag:"durationvar" usage:"durationvar usage" default:"1s"`
}

func TestVarFlagBasic(t *testing.T) {
	cfg := &Config1{}
	m := newMock()
	pf, err := varflag(m, "foo", cfg)
	if err != nil {
		t.Fatal(err)
	}
	pf()
	assert.Contains(t, m.data, "foo.intvar")
	assert.Equal(t, m.data["foo.intvar"], slot{usage: "intvar usage", value: -10})

	assert.Contains(t, m.data, "foo.boolvar")
	assert.Equal(t, m.data["foo.boolvar"], slot{usage: "boolvar usage", value: true})

	assert.Contains(t, m.data, "foo.uintvar")
	assert.Equal(t, m.data["foo.uintvar"], slot{usage: "uintvar usage", value: uint(10)})

	assert.Contains(t, m.data, "foo.int64var")
	assert.Equal(t, m.data["foo.int64var"], slot{usage: "int64var usage", value: int64(-10)})

	assert.Contains(t, m.data, "foo.stringvar")
	assert.Equal(t, m.data["foo.stringvar"], slot{usage: "stringvar usage", value: "hello"})

	assert.Contains(t, m.data, "foo.uint64var")
	assert.Equal(t, m.data["foo.uint64var"], slot{usage: "uint64var usage", value: uint64(10)})

	assert.Contains(t, m.data, "foo.float64var")
	assert.Equal(t, m.data["foo.float64var"], slot{usage: "float64var usage", value: float64(22.33)})

	assert.Contains(t, m.data, "foo.durationvar")
	assert.Equal(t, m.data["foo.durationvar"], slot{usage: "durationvar usage", value: time.Second})
}
