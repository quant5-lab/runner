package codegen

/* Drives coercion decisions when namespace expressions feed into Series.Set(float64) */
type GoValueType int

const (
	GoFloat64 GoValueType = iota
	GoBool
	GoString
)

func (t GoValueType) String() string {
	switch t {
	case GoBool:
		return "bool"
	case GoString:
		return "string"
	default:
		return "float64"
	}
}
