package codegen

type ValueType string

const (
	ValueTypeBool   ValueType = "bool"
	ValueTypeInt    ValueType = "int"
	ValueTypeFloat  ValueType = "float"
	ValueTypeString ValueType = "string"
)

type ConstantValue struct {
	boolValue   bool
	intValue    int
	floatValue  float64
	stringValue string
	valueType   ValueType
}

func NewBoolConstant(val bool) ConstantValue {
	return ConstantValue{
		boolValue: val,
		valueType: ValueTypeBool,
	}
}

func NewIntConstant(val int) ConstantValue {
	return ConstantValue{
		intValue:  val,
		valueType: ValueTypeInt,
	}
}

func NewFloatConstant(val float64) ConstantValue {
	return ConstantValue{
		floatValue: val,
		valueType:  ValueTypeFloat,
	}
}

func NewStringConstant(val string) ConstantValue {
	return ConstantValue{
		stringValue: val,
		valueType:   ValueTypeString,
	}
}

func (cv ConstantValue) IsBool() bool {
	return cv.valueType == ValueTypeBool
}

func (cv ConstantValue) IsInt() bool {
	return cv.valueType == ValueTypeInt
}

func (cv ConstantValue) IsFloat() bool {
	return cv.valueType == ValueTypeFloat
}

func (cv ConstantValue) IsString() bool {
	return cv.valueType == ValueTypeString
}

func (cv ConstantValue) AsBool() (bool, bool) {
	if cv.valueType == ValueTypeBool {
		return cv.boolValue, true
	}
	return false, false
}

func (cv ConstantValue) AsInt() (int, bool) {
	if cv.valueType == ValueTypeInt {
		return cv.intValue, true
	}
	return 0, false
}

func (cv ConstantValue) AsFloat() (float64, bool) {
	if cv.valueType == ValueTypeFloat {
		return cv.floatValue, true
	}
	return 0.0, false
}

func (cv ConstantValue) AsString() (string, bool) {
	if cv.valueType == ValueTypeString {
		return cv.stringValue, true
	}
	return "", false
}
