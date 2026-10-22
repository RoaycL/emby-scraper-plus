package router

import (
	"testing"
)

func TestClassifyNumber(t *testing.T) {
	tests := []struct {
		code     string
		expected ContentType
	}{
		{"SSIS-057", TypeCensored},
		{"ABW-358", TypeCensored},
		{"MIDV-123", TypeCensored},
		{"SIRO-1234", TypeAmateur},
		{"300MIUM-123", TypeAmateur},
		{"259LUXU-456", TypeAmateur},
		{"010120-001", TypeUncensored},
		{"HEYZO-1234", TypeUncensored},
		{"FC2-PPV-1234567", TypeFC2},
		{"FC-1234567", TypeFC2},
	}

	for _, tt := range tests {
		actual := ClassifyNumber(tt.code)
		if actual != tt.expected {
			t.Errorf("ClassifyNumber(%q) = %v; want %v", tt.code, actual, tt.expected)
		}
	}
}
