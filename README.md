Struct validator
============
验证器实现了基于标签定义验证规则并且生成验证代码的验证。

### 如何使用
安装
```go
go get 
```

```go
go install 
```
### 数字类型:
| Tag | 表述   | 示例     |
|-----|------|--------|
| eq  | 等于   | eq 10  |
| ne  | 不等于  | ne 10  |
| lt  | 小于   | lt 10  |
| gt  | 大于   | gt 10  |
| lte | 小于等于 | lte 10 |
| gte | 大于等于 | gte 10 |

### 字符串类型:
| Tag      | 表述     | 示例       |
|----------|--------|----------|
| notEmpty | 不为空    | notEmpty |
| max      | 字符最大长度 | max 10   |
| min      | 字符最小长度 | min 10   |

### Format:
| Tag       | 表述     | 示例        |
|-----------|--------|-----------|
| uuid3     | uuid3  | uuid3     |
| uuid4     | uuid4  | uuid4     |
| uuid5     | uuid5  | uuid5     |
| uuid      | uuid   | uuid      |
| email     | email  | email     |
| base64    | base64 | base64    |
| latitude  | 纬度     | latitude  |
| longitude | 经度     | longitude |
| phone     | 手机号码   | phone     |



### 示例
1. 定义验证规则

默认使用`check` 标签验证，通过调用`NewGenDefinition().SetTag()`设置自定义标签验证
```go
type MyInt int

type Test struct {
	Id        *int           `check:"gt 0;lte 100"`
	MyInt     MyInt          `check:"lt 100;ne 10"`
	Name      string         `check:"notEmpty"`
	age       int            `check:"gte 0;lte 100"`
	Score     float32        `check:"gt 0.00"`
	Email     string         `check:"email"`
	Max       string         `check:"max 10"`
	Min       string         `check:"min 5"`
	MyUUID    string         `check:"required;uuid"`
	Slice     []int          `check:"required"`
	Map       map[string]int 
	Chan      chan int       `check:"required"`
	b.Address `check:"required"`
	Addr      *b.Address `check:"required"`
	Phone     string     `check:"phone"`
}
```
2. 运行验证代码生成：
```go
pkg.NewGenDefinition().Gen(Test{})
```
或者`go install` 本项目到`$GOPATH`中，在需要验证的`.go`文件中添加`//go:generate struct-validate validate .`，然后在你的项目根目录下执行
`go generate ./...`

则会根据验证规则在`Test` 同一包下生成验证代码：`test_validate.go`
如下：
```go
func (t *Nested) Validator() error {
	if *t.Id <= 0 {
		return errors.New("id必须 gt 0")
	}
	if *t.Id > 100 {
		return errors.New("id必须 lte 100")
	}
	if t.MyInt >= 100 {
		return errors.New("my_int必须 lt 100")
	}
	if t.MyInt == 10 {
		return errors.New("my_int必须 ne 10")
	}
	if t.Name == "" {
		return errors.New("name不能为空")
	}
	if t.Score <= 0.00 {
		return errors.New("score必须 gt 0.00")
	}
	if !regexp.MustCompile(`^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`).MatchString(t.Email) {
		return errors.New("email 的规则不匹配")
	}
	if len(t.Max) >= 10 {
		return errors.New("max必须 max 10")
	}
	if len(t.Min) < 5 {
		return errors.New("min必须 min 5")
	}
	
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).MatchString(t.MyUUID) {
		return errors.New("my_u_u_i_d 的规则不匹配")
	}
	if t.Slice == nil {
		return errors.New("Slice 不能为nil ")
	}
	
	if t.Chan == nil {
		return errors.New("Chan 不能为nil ")
	}
	
	if err := t.Address.Validator(); err != nil {
		return err
	}
	if t.Addr == nil {
		return errors.New("Addr 不能为nil ")
	}
	if err := t.Addr.Validator(); err != nil {
		return err
	}
	if !regexp.MustCompile(`^1[3456789]\d{9}$`).MatchString(t.Phone) {
		return errors.New("phone 的规则不匹配")
	}
	
	return nil
}
```
3. 调用验证器
```go
    t := Test{
	   // ....
    }
    if err := t.Validator(); err != nil {
        // TODO handler error
    }
```
### 自定义验证
1. 添加注解: `// @ext:check` 
2. 签名必须是: `func() error`
```go
// TestNested XXX
// @ext:check
func (t Test) customValidator() error {
	// TODO do something
	return nil
}
```

### 自定义验证代码生成路径
```go
// @path:PATH
type Test struct {}
```
### 自定义验证代码包名
```go
// @package:PACKAGE_NAME
type Test struct {}
```