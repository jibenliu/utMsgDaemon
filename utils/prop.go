package utils

import (
	"errors"
	"go/ast"
	"reflect"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const (
	propertiesInterface = "org.freedesktop.DBus.Properties"
)

//Property ..自动进行struct字段的属性导出(通过tag的dbus进行配置)
type Property interface {
	Introspection() []introspect.Property                //获取struct中定义的需要导出的属性
	Interface() introspect.Interface                     //获取org.freedesktop.DBus.Properties接口定义
	Export(conn *dbus.Conn, path dbus.ObjectPath) error  //导出属性
	Emit(string, []string, map[string]interface{}) error //数据信息改变后触发change single
	Lock() *sync.Mutex                                   //获取锁
}

//Changer 触发数据变更
type Changer interface {
	Change([]string, map[string]dbus.Variant) *dbus.Error //数据变更触发接口(如果实现该方法，DBus触发属性修改时，执行该方法)
}

//property 属性管理
type property struct {
	original interface{}           //导出的原始数据结构
	value    reflect.Value         //通过InDirect方法获取信息
	props    map[string]*prop.Prop //导出属性集合
	mut      sync.Mutex            //读写锁
	conn     *dbus.Conn            //dbus连接
	path     string                //服务地址
}

//Get 通过dbus获取属性
//iface:接口名称
//key:需要获取的属性名称
func (p *property) Get(iface, key string) (dbus.Variant, *dbus.Error) {
	p.mut.Lock()
	defer p.mut.Unlock()
	item, has := p.props[key]
	if !has {
		return dbus.Variant{}, prop.ErrPropNotFound
	}
	value := dbus.MakeVariant(item.Value.(reflect.Value).Interface())
	if value.Signature().Empty() {
		return value, &dbus.ErrMsgUnknownInterface
	}
	return value, nil
}

//GetAll 获取所有属性
//iface:接口名称
func (p *property) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	p.mut.Lock()
	defer p.mut.Unlock()
	result := map[string]dbus.Variant{}

	for key, value := range p.props {
		result[key] = dbus.MakeVariant(value.Value.(reflect.Value).Interface())
		if result[key].Signature().Empty() {
			return nil, prop.ErrPropNotFound
		}
	}
	return result, nil
}

//Set 设置属性
//iface:接口名称
//key:设置属性名称
//newv:新值
func (p *property) Set(iface, key string, newv dbus.Variant) *dbus.Error {
	p.mut.Lock()
	defer p.mut.Unlock()
	item, err := p.canSet(key)
	if err != nil {
		return err
	}
	if change, pass := p.original.(Changer); pass {
		return change.Change([]string{key}, map[string]dbus.Variant{key: newv})
	}

	nv := reflect.New(item.Value.(reflect.Value).Type())
	err1 := newv.Store(nv.Interface())
	if err1 != nil {
		return prop.ErrInvalidArg
	}
	item.Value.(reflect.Value).Set(nv.Elem())
	return nil
}

//Introspection 获取所有导出属性信息
func (p *property) Introspection() []introspect.Property {
	s := make([]introspect.Property, 0, len(p.props))
	for k, v := range p.props {
		p := introspect.Property{Name: k, Type: dbus.SignatureOf(v.Value.(reflect.Value).Interface()).String()}
		if v.Writable {
			p.Access = "readwrite"
		} else {
			p.Access = "read"
		}
		p.Annotations = []introspect.Annotation{
			{
				Name:  "org.freedesktop.DBus.Property.EmitsChangedSignal",
				Value: v.Emit.String(),
			},
		}
		s = append(s, p)
	}
	return s
}

//Interface 获取org.freedesktop.DBus.Properties接口信息
func (p *property) Interface() introspect.Interface {
	return interfaceInfo(p)
}

//Export 导出服务
func (p *property) Export(conn *dbus.Conn, path dbus.ObjectPath) error {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.conn = conn
	p.path = string(path)
	return conn.Export(p, path, propertiesInterface)
}

//SetBatch 批量进行属性设置
func (p *property) SetBatch(iface string, keys []string, newv map[string]dbus.Variant) *dbus.Error {
	p.mut.Lock()
	defer p.mut.Unlock()
	if len(keys) == 0 {
		keys = make([]string, 0, len(newv))
		for key := range newv {
			keys = append(keys, key)
		}
	}
	for _, item := range keys {
		if _, err := p.canSet(item); err != nil {
			return err
		}
	}

	if change, pass := p.original.(Changer); pass {
		return change.Change(keys, newv)
	}

	value := make(map[string]reflect.Value)
	for _, key := range keys {
		item := p.props[key]
		nv := reflect.New(item.Value.(reflect.Value).Type())
		err := newv[key].Store(nv.Interface())
		if err != nil {
			return prop.ErrInvalidArg
		}
		value[key] = nv
	}
	for k, v := range value {
		p.props[k].Value.(reflect.Value).Set(v.Elem())
	}
	return nil
}

//Emit 触发属性变更事件
func (p *property) Emit(iface string, keys []string, props map[string]interface{}) error {
	p.mut.Lock()
	defer p.mut.Unlock()
	lkeys := len(keys)
	lprops := len(props)
	if lkeys == 0 && lprops == 0 {
		return prop.ErrInvalidArg
	} else if lkeys != 0 && lprops != 0 {
		if lkeys != lprops {
			return prop.ErrInvalidArg
		}
		for _, item := range keys {
			if _, has := p.props[item]; !has {
				return prop.ErrPropNotFound
			}
			if _, has := props[item]; !has {
				return prop.ErrPropNotFound
			}
		}
	}
	p.conn.Emit(dbus.ObjectPath(p.path), propertiesInterface+".PropertiesChanged", iface, props, keys)
	return nil
}

//Lock 获取全局锁
func (p *property) Lock() *sync.Mutex {
	return &p.mut
}

//canSet是否可以进行属性设置
func (p *property) canSet(v string) (*prop.Prop, *dbus.Error) {
	item, has := p.props[v]
	if !has {
		return item, prop.ErrPropNotFound
	}
	if !item.Writable || item.Emit == prop.EmitConst {
		return item, prop.ErrReadOnly
	}
	return item, nil
}

//----------------------辅助函数------------------------
func calc(v interface{}) (map[string]*prop.Prop, reflect.Value, error) {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, reflect.Value{}, errors.New("")
	}
	result := map[string]*prop.Prop{}
	tag := ""
	tp := value.Type()
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		if !ast.IsExported(field.Name) {
			return nil, reflect.Value{}, errors.New("")
		}
		tag = string(field.Tag.Get("dbus"))
		if tag == "" || tag == "-" {
			continue
		}
		p := &prop.Prop{}
		p.Value = value.Field(i)
		p.Writable = writeable(tag)
		p.Emit = emittype(tag)
		result[name(tag, field.Name)] = p
	}
	if len(result) == 0 {
		return nil, reflect.Value{}, errors.New("")
	}
	return result, value, nil
}

func writeable(tag string) bool {
	items := strings.Split(tag, ",")
	for _, item := range items {
		if item == "writeable" {
			return true
		}
	}
	return false
}

func emittype(tag string) prop.EmitType {
	items := strings.Split(tag, ",")
	for _, item := range items {
		switch item {
		case "emit", "emittrue":
			return prop.EmitTrue
		case "emitinvalidates", "invalidates":
			return prop.EmitInvalidates
		case "emitconst", "const":
			return prop.EmitConst
		}
	}
	return prop.EmitFalse
}

func name(tag, name string) string {
	items := strings.Split(tag, ",")
	for _, item := range items {
		if strings.HasPrefix(item, "name:") {
			return strings.TrimLeft(item, "name:")
		}
	}
	return name
}

func interfaceInfo(v interface{}) introspect.Interface {
	return introspect.Interface{
		Name:    propertiesInterface,
		Methods: introspect.Methods(v),
		Signals: []introspect.Signal{
			{
				Name: "PropertiesChanged",
				Args: []introspect.Arg{
					{
						Name: "interface",
						Type: "s",
					}, {
						Name: "changed_properties",
						Type: "a{sv}",
					}, {
						Name: "invalidates_properties",
						Type: "as",
					},
				},
			},
		},
	}
}

//NewProperty ..
func NewProperty(value interface{}) (Property, error) {
	props, rv, err := calc(value)
	if err != nil {
		return nil, err
	}
	return &property{
		original: value,
		value:    rv,
		props:    props,
	}, nil
}

//MultiPropStruct 多属性结果导出
type MultiPropStruct interface {
	Add(string, interface{}) (Property, error)
	Export(conn *dbus.Conn, path string) error
	Interface() introspect.Interface
}

type multiProp struct {
	lock  sync.Mutex
	iface map[string]*property
}

func (mp *multiProp) Get(iface, key string) (dbus.Variant, *dbus.Error) {
	mp.lock.Lock()
	v, has := mp.iface[iface]
	mp.lock.Unlock()
	if !has {
		return dbus.Variant{}, &dbus.ErrMsgUnknownInterface
	}
	return v.Get(iface, key)
}

func (mp *multiProp) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	mp.lock.Lock()
	v, has := mp.iface[iface]
	mp.lock.Unlock()
	if !has {
		return nil, &dbus.ErrMsgUnknownInterface
	}
	return v.GetAll(iface)
}

func (mp *multiProp) Set(iface, key string, newv dbus.Variant) *dbus.Error {
	mp.lock.Lock()
	v, has := mp.iface[iface]
	mp.lock.Unlock() //在高并发场景下，有可能会出现乱序问题(后请求先执行，如果出现可以根据实际情况集成后，重写该方法)
	if !has {
		return &dbus.ErrMsgUnknownInterface
	}
	return v.Set(iface, key, newv)
}

func (mp *multiProp) SetBatch(iface string, keys []string, newv map[string]dbus.Variant) *dbus.Error {
	mp.lock.Lock()
	v, has := mp.iface[iface]
	mp.lock.Unlock()
	if !has {
		return &dbus.ErrMsgUnknownInterface
	}
	return v.SetBatch(iface, keys, newv)
}

func (mp *multiProp) Export(conn *dbus.Conn, path string) error {
	mp.lock.Lock()
	defer mp.lock.Unlock()
	for _, item := range mp.iface {
		item.conn = conn
		item.path = path
	}
	return conn.Export(mp, dbus.ObjectPath(path), propertiesInterface)
}

func (mp *multiProp) Add(iface string, value interface{}) (Property, error) {
	p, err := NewProperty(value)
	if err != nil {
		return nil, err
	}
	mp.lock.Lock()
	defer mp.lock.Unlock()
	mp.iface[iface] = p.(*property)
	return p, nil
}

func (mp *multiProp) Interface() introspect.Interface {
	return interfaceInfo(mp)
}

//NewMulti 创建多属性结构导出
func NewMulti() (MultiPropStruct, error) {
	return &multiProp{
		iface: map[string]*property{},
	}, nil
}
