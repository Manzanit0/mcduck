package controllers_test

import (
	"fmt"
	"slices"
	"testing"

	api "github.com/manzanit0/mcduck/cmd/api/controllers"
	"github.com/manzanit0/mcduck/internal/expense"
)

func TestGroupSubcategoriesByCategory(t *testing.T) {
	testCases := []struct {
		expenses []expense.Expense
		result   map[string][]string
	}{
		{
			expenses: []expense.Expense{
				{Category: "a", Subcategory: "1"},
				{Category: "a", Subcategory: "2"},
				{Category: "b", Subcategory: "3"},
				{Category: "c", Subcategory: "4"},
				{Category: "c", Subcategory: "5"},
				{Category: "d", Subcategory: "1"},
			},
			result: map[string][]string{
				"a": {"1", "2"},
				"b": {"3"},
				"c": {"4", "5"},
				"d": {"1"},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			grouped := api.GroupSubcategoriesByCategory(tc.expenses)
			if len(grouped) != len(tc.result) {
				t.Fatalf("expected %d results, got %d", len(tc.result), len(grouped))
			}

			for category, subcategories := range grouped {
				for _, sub := range subcategories {
					if !slices.Contains(tc.result[category], sub) {
						t.Errorf("expected %s to contain %s", category, sub)
					}
				}
			}
		})
	}
}
