package codegen

import (
	"fmt"
	"strings"
)

type SimpleStringGenerator struct{}

func (g *SimpleStringGenerator) CanGenerate(funcName string) bool {
	simple := []string{
		"str.lower", "str.upper", "str.trim",
		"str.length", "str.tonumber",
	}
	for _, name := range simple {
		if funcName == name {
			return true
		}
	}
	return false
}

func (g *SimpleStringGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("%s requires at least 1 argument", funcName)
	}

	switch funcName {
	case "str.lower":
		return fmt.Sprintf("strings.ToLower(%s)", args[0]), nil
	case "str.upper":
		return fmt.Sprintf("strings.ToUpper(%s)", args[0]), nil
	case "str.trim":
		return fmt.Sprintf("strings.TrimSpace(%s)", args[0]), nil
	case "str.length":
		return fmt.Sprintf("len(%s)", args[0]), nil
	case "str.tonumber":
		return fmt.Sprintf("strconv.ParseFloat(%s, 64)", args[0]), nil
	default:
		return "", fmt.Errorf("unsupported function: %s", funcName)
	}
}

type TwoArgStringGenerator struct{}

func (g *TwoArgStringGenerator) CanGenerate(funcName string) bool {
	twoArg := []string{
		"str.contains", "str.startswith", "str.endswith",
		"str.pos", "str.split", "str.match",
	}
	for _, name := range twoArg {
		if funcName == name {
			return true
		}
	}
	return false
}

func (g *TwoArgStringGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("%s requires 2 arguments", funcName)
	}

	switch funcName {
	case "str.contains":
		return fmt.Sprintf("strings.Contains(%s, %s)", args[0], args[1]), nil
	case "str.startswith":
		return fmt.Sprintf("strings.HasPrefix(%s, %s)", args[0], args[1]), nil
	case "str.endswith":
		return fmt.Sprintf("strings.HasSuffix(%s, %s)", args[0], args[1]), nil
	case "str.pos":
		return fmt.Sprintf("strings.Index(%s, %s)", args[0], args[1]), nil
	case "str.split":
		return fmt.Sprintf("strings.Split(%s, %s)", args[0], args[1]), nil
	case "str.match":
		return fmt.Sprintf("regexp.MustCompile(%s).FindString(%s)", args[1], args[0]), nil
	default:
		return "", fmt.Errorf("unsupported function: %s", funcName)
	}
}

type SubstringGenerator struct{}

func (g *SubstringGenerator) CanGenerate(funcName string) bool {
	return funcName == "str.substring"
}

func (g *SubstringGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("str.substring requires at least 2 arguments")
	}

	if len(args) == 2 {
		return fmt.Sprintf("%s[%s:]", args[0], args[1]), nil
	}
	return fmt.Sprintf("%s[%s:%s]", args[0], args[1], args[2]), nil
}

type ReplaceGenerator struct{}

func (g *ReplaceGenerator) CanGenerate(funcName string) bool {
	return funcName == "str.replace" || funcName == "str.replace_all"
}

func (g *ReplaceGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("%s requires at least 3 arguments", funcName)
	}

	if funcName == "str.replace_all" {
		return fmt.Sprintf("strings.ReplaceAll(%s, %s, %s)", args[0], args[1], args[2]), nil
	}

	count := "-1"
	if len(args) >= 4 {
		count = args[3]
	}
	return fmt.Sprintf("strings.Replace(%s, %s, %s, %s)", args[0], args[1], args[2], count), nil
}

type RepeatGenerator struct{}

func (g *RepeatGenerator) CanGenerate(funcName string) bool {
	return funcName == "str.repeat"
}

func (g *RepeatGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("str.repeat requires at least 2 arguments")
	}

	separator := `""`
	if len(args) >= 3 {
		separator = args[2]
	}
	return fmt.Sprintf("strings.Repeat(%s + %s, %s)", args[0], separator, args[1]), nil
}

type ToStringGenerator struct{}

func (g *ToStringGenerator) CanGenerate(funcName string) bool {
	return funcName == "str.tostring"
}

func (g *ToStringGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("str.tostring requires at least 1 argument")
	}

	format := `"%.10g"`
	if len(args) >= 2 {
		format = args[1]
	}
	return fmt.Sprintf("fmt.Sprintf(%s, %s)", format, args[0]), nil
}

type FormatGenerator struct{}

func (g *FormatGenerator) CanGenerate(funcName string) bool {
	return funcName == "str.format"
}

func (g *FormatGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("str.format requires at least 1 argument")
	}

	if len(args) == 1 {
		return args[0], nil
	}

	formatArgs := strings.Join(args[1:], ", ")
	return fmt.Sprintf("fmt.Sprintf(%s, %s)", args[0], formatArgs), nil
}

type FormatTimeGenerator struct{}

func (g *FormatTimeGenerator) CanGenerate(funcName string) bool {
	return funcName == "str.format_time"
}

func (g *FormatTimeGenerator) Generate(gen *generator, funcName string, args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("str.format_time requires at least 2 arguments")
	}

	timezone := `"UTC"`
	if len(args) >= 3 {
		timezone = args[2]
	}

	return fmt.Sprintf(
		"time.Unix(%s/1000, 0).In(time.LoadLocation(%s)).Format(%s)",
		args[0], timezone, args[1],
	), nil
}
