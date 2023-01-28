package internal

import (
	"SJT/struct-validate/utils"
	"SJT/struct-validate/utils/slice"
	"fmt"
)

type Operator string

const (
	// NotEmpty notEmpty不为空
	NotEmpty Operator = "notEmpty"
	// Eq eq等于
	Eq Operator = "eq"
	// Ne ne不等于
	Ne Operator = "ne"
	// Lt lt 小于
	Lt Operator = "lt"
	// Gt gt 大于
	Gt Operator = "gt"
	// Lte lte 小于等于
	Lte Operator = "lte"
	// Gte gte 大于等于
	Gte Operator = "gte"
	// Max 字符最大长度
	Max Operator = "max"
	// Min 字符最小长度
	Min Operator = "min"
	// UUID3 uuid3
	UUID3     Operator = "uuid3"
	UUID4     Operator = "uuid4"
	UUID5     Operator = "uuid5"
	UUID      Operator = "uuid"
	Email     Operator = "email"
	Base64    Operator = "base64"
	Latitude  Operator = "latitude"
	Longitude Operator = "longitude"
	Phone     Operator = "phone"
)

func (s Operator) String() string {
	return string(s)
}

// 所有支持的规则
var roles = map[Operator]struct{}{
	NotEmpty:  {},
	Eq:        {},
	Ne:        {},
	Lt:        {},
	Gt:        {},
	Lte:       {},
	Gte:       {},
	UUID3:     {},
	UUID4:     {},
	UUID5:     {},
	UUID:      {},
	Email:     {},
	Base64:    {},
	Latitude:  {},
	Longitude: {},
	Max:       {},
	Min:       {},
	Phone:     {},
}

var normalRoles = map[Operator]string{
	NotEmpty: `==`,
	Eq:       "!=",
	Ne:       "==",
	Lt:       ">=",
	Gt:       "<=",
	Lte:      ">",
	Gte:      "<",
	Max:      "",
	Min:      "",
}

var regexpRoles = map[Operator]string{
	UUID3:     `^[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	UUID4:     `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	UUID5:     `^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	UUID:      `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	Email:     `^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`,
	Base64:    `^(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{2}==|[A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{4})$`,
	Latitude:  `^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?)$`,
	Longitude: `^[-+]?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$`,
	Phone:     `^1[3456789]\d{9}$`,
}

func (t Tag) Check(operator string) bool {
	_, ok := roles[Operator(operator)]
	return ok
}

func (t Tag) GeError(field, operator string, value any) string {
	field = utils.UnderscoreName(field)
	if _, ok := normalRoles[Operator(operator)]; ok {
		if Operator(operator) == NotEmpty {
			return field + "不能为空"
		}

		return fmt.Sprintf("%s必须 %s %v", field, Operator(operator).String(), value)
	}

	if _, ok := regexpRoles[Operator(operator)]; ok {
		return fmt.Sprintf("%s 的规则不匹配", field)
	}
	return ""
}

// Expression 表达式策略
type Expression interface {
	get(field, star, operator string, value any, realType string) string
}

func (*Tag) GetExp(field, star, operator string, value any, realType string) string {
	if ot, ok := normalRoles[Operator(operator)]; ok {
		switch operator {
		case NotEmpty.String():
			if realType == "string" {
				return fmt.Sprintf(`%st.%s %s ""`, star, field, ot)
			}
		case Max.String():
			if realType == "string" {
				return fmt.Sprintf(`len(%st.%s) >= %s`, star, field, value)
			}
		case Min.String():
			if realType == "string" {
				return fmt.Sprintf(`len(%st.%s) < %s`, star, field, value)
			}
		default:
			// 数字类型
			if slice.Contains[string](numeric, realType) {
				return fmt.Sprintf("%st.%s %s %s", star, field, ot, value)
			}
		}

	}

	if _regexp, ok := regexpRoles[Operator(operator)]; ok {
		if realType == "string" {
			return fmt.Sprintf("!regexp.MustCompile(`%s`).MatchString(t.%s)", _regexp, field)
		}
	}
	return ""
}

// GetStarType returns * when field is pointer
func (n *Node) GetStarType() string {
	if n.Kind == "ptr" {
		return "*"
	}
	return ""
}

var ptrType = map[string]struct{}{
	"struct": {},
	"slice":  {},
	"chan":   {},
	"map":    {},
}

// ShouldValidateNil 应该校验是否为nil
func (n *Node) ShouldValidateNil() bool {
	_, ok := ptrType[n.RealType]
	if ok {
		if n.RealType == "struct" {
			return n.Kind == "ptr"
		}
		return true
	}
	return false
}

// numeric 数字类型
var numeric = []string{"int", "uint", "int8", "uint8", "int32", "uint32", "int64", "uint64", "float32", "float64"}
