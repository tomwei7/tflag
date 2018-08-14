package tflag

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type flagSet []string

func (fs *flagSet) Set(val string) error {
	*fs = append(*fs, val)
	return nil
}

func (fs *flagSet) String() string {
	return strings.Join(*fs, ",")
}

// ParseFunc .
type ParseFunc func() error

// Var .
func Var(namespace string, v interface{}) (ParseFunc, error) {
	return varflag(defaultFlager{}, namespace, v)
}

// VarFlagSet .
func VarFlagSet(fs *flag.FlagSet, namespace string, v interface{}) (ParseFunc, error) {
	return varflag(fs, namespace, v)
}

type flager interface {
	Var(p flag.Value, name string, usage string)
	IntVar(p *int, name string, value int, usage string)
	BoolVar(p *bool, name string, value bool, usage string)
	UintVar(p *uint, name string, value uint, usage string)
	Int64Var(p *int64, name string, value int64, usage string)
	StringVar(p *string, name string, value string, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
	Float64Var(p *float64, name string, value float64, usage string)
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
}

type defaultFlager struct{}

func (defaultFlager) Var(p flag.Value, name string, usage string) {
	flag.Var(p, name, usage)
}

func (defaultFlager) IntVar(p *int, name string, value int, usage string) {
	flag.IntVar(p, name, value, usage)
}

func (defaultFlager) BoolVar(p *bool, name string, value bool, usage string) {
	flag.BoolVar(p, name, value, usage)
}

func (defaultFlager) UintVar(p *uint, name string, value uint, usage string) {
	flag.UintVar(p, name, value, usage)
}

func (defaultFlager) Int64Var(p *int64, name string, value int64, usage string) {
	flag.Int64Var(p, name, value, usage)
}

func (defaultFlager) StringVar(p *string, name string, value string, usage string) {
	flag.StringVar(p, name, value, usage)
}

func (defaultFlager) Uint64Var(p *uint64, name string, value uint64, usage string) {
	flag.Uint64Var(p, name, value, usage)
}

func (defaultFlager) Float64Var(p *float64, name string, value float64, usage string) {
	flag.Float64Var(p, name, value, usage)
}

func (defaultFlager) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	flag.DurationVar(p, name, value, usage)
}

func varflag(f flager, namespace string, v interface{}) (ParseFunc, error) {
	prefix := ""
	if namespace != "" {
		prefix = namespace + "."
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("v must non-pointer and not nil")
	}
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	vs := varState{f: f}
	return vs.parseFunc, vs.initFlag(prefix, rv)
}

type tagOpt struct {
	flag       string
	env        string
	usage      string
	defaultVar string
}

func parseTag(name string, tag reflect.StructTag) tagOpt {
	opt := tagOpt{flag: strings.ToLower(name)}
	if name, ok := tag.Lookup("flag"); ok {
		opt.flag = name
	}
	if env, ok := tag.Lookup("env"); ok {
		opt.env = env
	}
	if usage, ok := tag.Lookup("usage"); ok {
		opt.usage = usage
	}
	if defaultVar, ok := tag.Lookup("default"); ok {
		opt.defaultVar = defaultVar
	}
	return opt
}

type varState struct {
	f   flager
	pfs []ParseFunc
}

func (v *varState) initFlag(prefix string, rv reflect.Value) (err error) {
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("v must pointer of struct")
	}
	tv := rv.Type()
	for i := 0; i < tv.NumField(); i++ {
		opt := parseTag(tv.Field(i).Name, tv.Field(i).Tag)
		if opt.flag == "-" {
			continue
		}
		if err = v.setFlag(prefix, rv.Field(i), opt); err != nil {
			return
		}
	}
	return
}

func (v *varState) parseFunc() (err error) {
	for _, fn := range v.pfs {
		if err = fn(); err != nil {
			return
		}
	}
	return
}

func (v *varState) setFlag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	switch rv.Kind() {
	case reflect.Int:
		err = v.setIntFlag(prefix, rv, opt)
	case reflect.Bool:
		err = v.setBoolFlag(prefix, rv, opt)
	case reflect.Uint:
		err = v.setUintFlag(prefix, rv, opt)
	case reflect.Int64:
		if rv.Type().Name() == "Duration" {
			err = v.setDurationFlag(prefix, rv, opt)
		} else {
			err = v.setInt64Flag(prefix, rv, opt)
		}
	case reflect.String:
		err = v.setStringFlag(prefix, rv, opt)
	case reflect.Uint64:
		err = v.setUint64Flag(prefix, rv, opt)
	case reflect.Float64:
		err = v.setFloat64Flag(prefix, rv, opt)
	case reflect.Slice:
		err = v.setSliceFlag(prefix, rv, opt)
	case reflect.Struct:
		prefix = prefix + opt.flag + "."
		err = v.initFlag(prefix, rv)
	}
	return nil
}

func (v *varState) setSliceFlag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var fs flagSet
	v.f.Var(&fs, prefix+opt.flag, opt.usage)
	v.pfs = append(v.pfs, func() error {
		var setFn func(flagSet, reflect.Value) error
		switch rv.Elem().Kind() {
		case reflect.Bool:
			setFn = setBoolSlice
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			setFn = setInt64Slice
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			setFn = setUint64Slice
		case reflect.Float32, reflect.Float64:
			setFn = setFloat64Slice
		default:
			return fmt.Errorf("unsupport type %s, only support base type e.g. string, bool, int", rv.Elem().Kind())
		}
		v := reflect.MakeSlice(rv.Type(), len(fs), len(fs))
		if err := setFn(fs, v); err != nil {
			return err
		}
		rv.Set(v)
		return nil
	})
	return
}

func setStringSlice(fs flagSet, v reflect.Value) error {
	for i, s := range fs {
		v.Index(i).SetString(s)
	}
	return nil
}

func setInt64Slice(fs flagSet, v reflect.Value) error {
	for i, s := range fs {
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.Index(i).SetInt(x)
	}
	return nil
}

func setUint64Slice(fs flagSet, v reflect.Value) error {
	for i, s := range fs {
		x, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.Index(i).SetUint(x)
	}
	return nil
}

func setFloat64Slice(fs flagSet, v reflect.Value) error {
	for i, s := range fs {
		x, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.Index(i).SetFloat(x)
	}
	return nil
}

func setBoolSlice(fs flagSet, v reflect.Value) error {
	for i, s := range fs {
		x, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.Index(i).SetBool(x)
	}
	return nil
}

func (v *varState) setIntFlag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var defaultVar int
	if opt.defaultVar != "" {
		if defaultVar, err = strconv.Atoi(opt.defaultVar); err != nil {
			return fmt.Errorf("invalid default value for flag %s error: %s", prefix+opt.flag, err)
		}
	}
	v.f.IntVar(rv.Addr().Interface().(*int), prefix+opt.flag, defaultVar, opt.usage)
	return
}

func (v *varState) setBoolFlag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var defaultVar bool
	if opt.defaultVar != "" {
		if defaultVar, err = strconv.ParseBool(opt.defaultVar); err != nil {
			return fmt.Errorf("invalid default value for flag %s error: %s", prefix+opt.flag, err)
		}
	}
	v.f.BoolVar(rv.Addr().Interface().(*bool), prefix+opt.flag, defaultVar, opt.usage)
	return
}

func (v *varState) setUintFlag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var defaultVar uint64
	if opt.defaultVar != "" {
		if defaultVar, err = strconv.ParseUint(opt.defaultVar, 10, 32); err != nil {
			return fmt.Errorf("invalid default value for flag %s error: %s", prefix+opt.flag, err)
		}
	}
	v.f.UintVar(rv.Addr().Interface().(*uint), prefix+opt.flag, uint(defaultVar), opt.usage)
	return
}

func (v *varState) setUint64Flag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var defaultVar uint64
	if opt.defaultVar != "" {
		if defaultVar, err = strconv.ParseUint(opt.defaultVar, 10, 64); err != nil {
			return fmt.Errorf("invalid default value for flag %s error: %s", prefix+opt.flag, err)
		}
	}
	v.f.Uint64Var(rv.Addr().Interface().(*uint64), prefix+opt.flag, defaultVar, opt.usage)
	return
}

func (v *varState) setInt64Flag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var defaultVar int64
	if opt.defaultVar != "" {
		if defaultVar, err = strconv.ParseInt(opt.defaultVar, 10, 64); err != nil {
			return fmt.Errorf("invalid default value for flag %s error: %s", prefix+opt.flag, err)
		}
	}
	v.f.Int64Var(rv.Addr().Interface().(*int64), prefix+opt.flag, defaultVar, opt.usage)
	return
}

func (v *varState) setStringFlag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	v.f.StringVar(rv.Addr().Interface().(*string), prefix+opt.flag, opt.defaultVar, opt.usage)
	return
}

func (v *varState) setFloat64Flag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var defaultVar float64
	if opt.defaultVar != "" {
		if defaultVar, err = strconv.ParseFloat(opt.defaultVar, 64); err != nil {
			return fmt.Errorf("invalid default value for flag %s error: %s", prefix+opt.flag, err)
		}
	}
	v.f.Float64Var(rv.Addr().Interface().(*float64), prefix+opt.flag, defaultVar, opt.usage)
	return
}

func (v *varState) setDurationFlag(prefix string, rv reflect.Value, opt tagOpt) (err error) {
	var defaultVar time.Duration
	if opt.defaultVar != "" {
		if defaultVar, err = time.ParseDuration(opt.defaultVar); err != nil {
			return fmt.Errorf("invalid default value for flag %s error: %s", prefix+opt.flag, err)
		}
	}
	v.f.DurationVar(rv.Addr().Interface().(*time.Duration), prefix+opt.flag, defaultVar, opt.usage)
	return
}
