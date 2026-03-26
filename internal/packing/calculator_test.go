package packing

import (
	"reflect"
	"testing"
)

func TestCalculatorCalculate(t *testing.T) {
	calculator, err := NewCalculator([]int{250, 500, 1000, 2000, 5000})
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	tests := []struct {
		name      string
		ordered   int
		total     int
		expected  []PackSummary
	}{
		{
			name:    "minimum order rounds up to one smallest pack",
			ordered: 1,
			total:   250,
			expected: []PackSummary{{PackSize: 250, Count: 1}},
		},
		{
			name:    "exact single pack match",
			ordered: 250,
			total:   250,
			expected: []PackSummary{{PackSize: 250, Count: 1}},
		},
		{
			name:    "prefers fewer packs once total is minimized",
			ordered: 251,
			total:   500,
			expected: []PackSummary{{PackSize: 500, Count: 1}},
		},
		{
			name:    "sample 501",
			ordered: 501,
			total:   750,
			expected: []PackSummary{{PackSize: 500, Count: 1}, {PackSize: 250, Count: 1}},
		},
		{
			name:    "sample 12001",
			ordered: 12001,
			total:   12250,
			expected: []PackSummary{{PackSize: 5000, Count: 2}, {PackSize: 2000, Count: 1}, {PackSize: 250, Count: 1}},
		},
		{
			name:    "avoids greedy overfill",
			ordered: 2750,
			total:   2750,
			expected: []PackSummary{{PackSize: 2000, Count: 1}, {PackSize: 500, Count: 1}, {PackSize: 250, Count: 1}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := calculator.Calculate(test.ordered)
			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}

			if result.TotalItems != test.total {
				t.Fatalf("TotalItems = %d, want %d", result.TotalItems, test.total)
			}

			if !reflect.DeepEqual(result.Packs, test.expected) {
				t.Fatalf("Packs = %#v, want %#v", result.Packs, test.expected)
			}
		})
	}
}

func TestCalculatorCalculateRejectsInvalidOrder(t *testing.T) {
	calculator, err := NewCalculator([]int{250, 500})
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	if _, err := calculator.Calculate(0); err == nil {
		t.Fatal("Calculate() expected error for zero quantity")
	}
}

func TestNewCalculatorRejectsDuplicatePackSizes(t *testing.T) {
	if _, err := NewCalculator([]int{250, 250}); err == nil {
		t.Fatal("NewCalculator() expected duplicate pack size error")
	}
}
