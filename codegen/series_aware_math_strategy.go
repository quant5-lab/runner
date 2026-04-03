package codegen

import (
	"fmt"
	"strings"
)

type SeriesAwareMathStrategy interface {
	GenerateFromExtracted(args []string) (string, error)
	ValidateArgCount(count int) error
}

type SeriesAwareUnaryGenerator struct {
	goFunc string
}

func NewSeriesAwareUnaryGenerator(goFunc string) *SeriesAwareUnaryGenerator {
	return &SeriesAwareUnaryGenerator{goFunc: goFunc}
}

func (s *SeriesAwareUnaryGenerator) ValidateArgCount(count int) error {
	if count != 1 {
		return fmt.Errorf("unary function requires 1 argument, got %d", count)
	}
	return nil
}

func (s *SeriesAwareUnaryGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s)", s.goFunc, args[0]), nil
}

type SeriesAwareBinaryGenerator struct {
	goFunc string
}

func NewSeriesAwareBinaryGenerator(goFunc string) *SeriesAwareBinaryGenerator {
	return &SeriesAwareBinaryGenerator{goFunc: goFunc}
}

func (s *SeriesAwareBinaryGenerator) ValidateArgCount(count int) error {
	if count != 2 {
		return fmt.Errorf("binary function requires 2 arguments, got %d", count)
	}
	return nil
}

func (s *SeriesAwareBinaryGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s, %s)", s.goFunc, args[0], args[1]), nil
}

type SeriesAwareSignGenerator struct{}

func NewSeriesAwareSignGenerator() *SeriesAwareSignGenerator {
	return &SeriesAwareSignGenerator{}
}

func (s *SeriesAwareSignGenerator) ValidateArgCount(count int) error {
	if count != 1 {
		return fmt.Errorf("sign requires 1 argument, got %d", count)
	}
	return nil
}

func (s *SeriesAwareSignGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	return fmt.Sprintf("func() float64 { if %s > 0 { return 1 } else if %s < 0 { return -1 } else { return 0 } }()", args[0], args[0]), nil
}

type SeriesAwareToDegreesGenerator struct{}

func NewSeriesAwareToDegreesGenerator() *SeriesAwareToDegreesGenerator {
	return &SeriesAwareToDegreesGenerator{}
}

func (s *SeriesAwareToDegreesGenerator) ValidateArgCount(count int) error {
	if count != 1 {
		return fmt.Errorf("todegrees requires 1 argument, got %d", count)
	}
	return nil
}

func (s *SeriesAwareToDegreesGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s * 180.0 / math.Pi)", args[0]), nil
}

type SeriesAwareToRadiansGenerator struct{}

func NewSeriesAwareToRadiansGenerator() *SeriesAwareToRadiansGenerator {
	return &SeriesAwareToRadiansGenerator{}
}

func (s *SeriesAwareToRadiansGenerator) ValidateArgCount(count int) error {
	if count != 1 {
		return fmt.Errorf("toradians requires 1 argument, got %d", count)
	}
	return nil
}

func (s *SeriesAwareToRadiansGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s * math.Pi / 180.0)", args[0]), nil
}

type SeriesAwareAvgGenerator struct{}

func NewSeriesAwareAvgGenerator() *SeriesAwareAvgGenerator {
	return &SeriesAwareAvgGenerator{}
}

func (s *SeriesAwareAvgGenerator) ValidateArgCount(count int) error {
	if count < 1 {
		return fmt.Errorf("avg requires at least 1 argument, got %d", count)
	}
	return nil
}

func (s *SeriesAwareAvgGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	sum := strings.Join(args, " + ")
	return fmt.Sprintf("((%s) / %d.0)", sum, len(args)), nil
}

type SeriesAwareRandomGenerator struct{}

func NewSeriesAwareRandomGenerator() *SeriesAwareRandomGenerator {
	return &SeriesAwareRandomGenerator{}
}

func (s *SeriesAwareRandomGenerator) ValidateArgCount(count int) error {
	if count > 3 {
		return fmt.Errorf("random accepts at most 3 arguments, got %d", count)
	}
	return nil
}

func (s *SeriesAwareRandomGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	switch len(args) {
	case 0:
		return "rand.Float64()", nil
	case 1:
		return fmt.Sprintf("rand.Float64() * %s", args[0]), nil
	case 2:
		return fmt.Sprintf("(%s + rand.Float64() * (%s - %s))", args[0], args[1], args[0]), nil
	case 3:
		return fmt.Sprintf("(%s + rand.Float64() * (%s - %s))", args[0], args[1], args[0]), nil
	default:
		return "rand.Float64()", nil
	}
}

type SeriesAwareRoundToMintickGenerator struct{}

func NewSeriesAwareRoundToMintickGenerator() *SeriesAwareRoundToMintickGenerator {
	return &SeriesAwareRoundToMintickGenerator{}
}

func (s *SeriesAwareRoundToMintickGenerator) ValidateArgCount(count int) error {
	if count != 1 {
		return fmt.Errorf("round_to_mintick requires 1 argument, got %d", count)
	}
	return nil
}

func (s *SeriesAwareRoundToMintickGenerator) GenerateFromExtracted(args []string) (string, error) {
	if err := s.ValidateArgCount(len(args)); err != nil {
		return "", err
	}
	return fmt.Sprintf("(math.Round(%s / ctx.Mintick) * ctx.Mintick)", args[0]), nil
}
