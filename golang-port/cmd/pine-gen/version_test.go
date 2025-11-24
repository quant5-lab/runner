package main

import (
	"testing"
)

func TestDetectPineVersion_ValidVersions(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "version 5",
			content:  "//@version=5\nindicator(\"test\")",
			expected: 5,
		},
		{
			name:     "version 4",
			content:  "//@version=4\nstudy(\"test\")",
			expected: 4,
		},
		{
			name:     "version 3",
			content:  "//@version=3\nstudy(\"test\")",
			expected: 3,
		},
		{
			name:     "version 2",
			content:  "//@version=2\nstudy(\"test\")",
			expected: 2,
		},
		{
			name:     "version 1",
			content:  "//@version=1\nstudy(\"test\")",
			expected: 1,
		},
		{
			name:     "version with spaces",
			content:  "//@version  =  5\nindicator(\"test\")",
			expected: 5,
		},
		{
			name:     "version in middle of file",
			content:  "// Comment\n//@version=5\nindicator(\"test\")",
			expected: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detectPineVersion(tc.content)
			if result != tc.expected {
				t.Errorf("Expected version %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestDetectPineVersion_NoVersion(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "no version comment",
			content: "study(\"test\")\nma = sma(close, 20)",
		},
		{
			name:    "empty file",
			content: "",
		},
		{
			name:    "only comments",
			content: "// This is a comment\n// Another comment",
		},
		{
			name:    "malformed version comment",
			content: "// version=5\nstudy(\"test\")",
		},
		{
			name:    "version without @",
			content: "//version=5\nstudy(\"test\")",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detectPineVersion(tc.content)
			if result != 4 {
				t.Errorf("Expected default version 4, got %d", result)
			}
		})
	}
}

func TestDetectPineVersion_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "multiple version comments (first wins)",
			content:  "//@version=4\n//@version=5\nstudy(\"test\")",
			expected: 4,
		},
		{
			name:     "version 10 (future version)",
			content:  "//@version=10\nindicator(\"test\")",
			expected: 10,
		},
		{
			name:     "version 0",
			content:  "//@version=0\nstudy(\"test\")",
			expected: 0,
		},
		{
			name:     "version in string (regex finds it anyway)",
			content:  "study(\"//@version=5\")\nma = sma(close, 20)",
			expected: 5, // Simple regex will match even in strings
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detectPineVersion(tc.content)
			if result != tc.expected {
				t.Errorf("Expected version %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestDetectPineVersion_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "non-numeric version",
			content:  "//@version=five\nstudy(\"test\")",
			expected: 4, // Should default to 4 when parsing fails
		},
		{
			name:     "version with decimal",
			content:  "//@version=5.0\nstudy(\"test\")",
			expected: 5, // Should parse as 5
		},
		{
			name:     "negative version",
			content:  "//@version=-1\nstudy(\"test\")",
			expected: 4, // Sscanf will fail, default to 4
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detectPineVersion(tc.content)
			if result != tc.expected {
				t.Errorf("Expected version %d, got %d", tc.expected, result)
			}
		})
	}
}
