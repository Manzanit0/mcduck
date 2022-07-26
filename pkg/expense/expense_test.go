package expense_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/manzanit0/mcduck/pkg/expense"
)

func TestCalculateTotalsPerCategory(t *testing.T) {
	testCases := []struct {
		expenses []expense.Expense
		result   map[string]map[string]float32
	}{
		{
			expenses: []expense.Expense{
				{Date: time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC), Category: "a", Amount: 1},
				{Date: time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC), Category: "a", Amount: 1},
				{Date: time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC), Category: "b", Amount: 1},
				{Date: time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC), Category: "c", Amount: 1},
				{Date: time.Date(2006, 2, 1, 0, 0, 0, 0, time.UTC), Category: "c", Amount: 2},
			},
			result: map[string]map[string]float32{
				"2006-01": {"a": 1},
				"2006-02": {"a": 1, "b": 1, "c": 3},
			},
		},
		{
			expenses: []expense.Expense{
				{Date: time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC), Category: "a", Amount: 1.3},
				{Date: time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC), Category: "a", Amount: 1.2},
			},
			result: map[string]map[string]float32{
				"2006-01": {"a": 1.3},
				"2008-01": {"a": 1.2},
			},
		},
	}
	for x, tC := range testCases {
		t.Run(fmt.Sprintf("case %d", x), func(t *testing.T) {
			totals := expense.CalculateTotalsPerCategory(tC.expenses)
			if len(totals) != len(tC.result) {
				t.Fatalf("expected %d results, got %d", len(tC.result), len(totals))
			}

			for month, monthTotals := range totals {
				for category, total := range monthTotals {
					if tC.result[month][category] != total {
						t.Errorf("expected %f for %s-%s, got %f", tC.result[month][category], month, category, total)
					}
				}
			}
		})
	}
}
