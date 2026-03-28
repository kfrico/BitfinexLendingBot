package strategy

import (
	"reflect"
	"testing"
)

func TestBuildOrderAmounts(t *testing.T) {
	tests := []struct {
		name            string
		totalFunds      float64
		requestedSplits int
		minLoan         float64
		maxLoan         float64
		expected        []float64
	}{
		{
			name:            "caps each order at max loan and leaves remainder",
			totalFunds:      5000,
			requestedSplits: 3,
			minLoan:         150,
			maxLoan:         1000,
			expected:        []float64{1000, 1000, 1000},
		},
		{
			name:            "uses requested split count when evenly feasible",
			totalFunds:      2500,
			requestedSplits: 3,
			minLoan:         150,
			maxLoan:         1000,
			expected:        []float64{833.34, 833.33, 833.33},
		},
		{
			name:            "falls back to average split when below max loan",
			totalFunds:      400,
			requestedSplits: 3,
			minLoan:         150,
			maxLoan:         1000,
			expected:        []float64{200, 200},
		},
		{
			name:            "distributes all funds without leaving remainder idle",
			totalFunds:      1000,
			requestedSplits: 4,
			minLoan:         150,
			maxLoan:         300,
			expected:        []float64{250, 250, 250, 250},
		},
		{
			name:            "never rounds above available funds",
			totalFunds:      197.539,
			requestedSplits: 1,
			minLoan:         150,
			maxLoan:         1000,
			expected:        []float64{197.53},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildOrderAmounts(tt.totalFunds, tt.requestedSplits, tt.minLoan, tt.maxLoan)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}
