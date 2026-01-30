package codegen

import (
	"fmt"
	"strings"
)

type SecurityCacheKeyBuilder struct{}

func NewSecurityCacheKeyBuilder() *SecurityCacheKeyBuilder {
	return &SecurityCacheKeyBuilder{}
}

type CacheKeyComponents struct {
	KeyPattern string
	FormatArgs string
}

func (b *SecurityCacheKeyBuilder) Build(symbolResult, timeframeResult *ExtractionResult) CacheKeyComponents {
	if symbolResult.IsRuntime && timeframeResult.IsRuntime {
		return CacheKeyComponents{
			KeyPattern: "%s:%s",
			FormatArgs: fmt.Sprintf("%s, %s", symbolResult.Code, timeframeResult.Code),
		}
	}

	if symbolResult.IsRuntime {
		tfLiteral := strings.Trim(timeframeResult.Code, `"`)
		return CacheKeyComponents{
			KeyPattern: fmt.Sprintf("%%s:%s", tfLiteral),
			FormatArgs: symbolResult.Code,
		}
	}

	if timeframeResult.IsRuntime {
		symbolLiteral := strings.Trim(symbolResult.Code, `"`)
		return CacheKeyComponents{
			KeyPattern: fmt.Sprintf("%s:%%s", symbolLiteral),
			FormatArgs: timeframeResult.Code,
		}
	}

	symbolLiteral := strings.Trim(symbolResult.Code, `"`)
	tfLiteral := strings.Trim(timeframeResult.Code, `"`)
	return CacheKeyComponents{
		KeyPattern: fmt.Sprintf("%s:%s", symbolLiteral, tfLiteral),
		FormatArgs: "",
	}
}
