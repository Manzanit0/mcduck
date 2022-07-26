package term

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
)

func Render(grouping string, calculations map[string]map[string]float32) {
	// get all unique categories
	categories := make(map[string]bool)
	for _, amountByCategory := range calculations {
		for category := range amountByCategory {
			if ok := categories[category]; !ok {
				categories[category] = true
			}
		}
	}

	nowMonthYear := time.Now().Format("2006-01")

	// first get the categories header
	chartCategories := []pterm.Bar{}
	tableCategoriesHeader := make([]string, len(categories)+1)
	tableCategoriesHeader[0] = ""
	i := 1
	for category := range categories {
		tableCategoriesHeader[i] = category
		i++

		if v := calculations[nowMonthYear][category]; v != 0 {
			chartCategories = append(chartCategories, pterm.Bar{Label: category, Value: int(calculations[nowMonthYear][category])})
		}
	}

	pterm.Println()
	pterm.NewRGB(15, 199, 209).Printf("This month per %s", grouping)
	pterm.Println()
	err := pterm.DefaultBarChart.WithBars(chartCategories).WithHorizontal(true).WithShowValue(true).Render()
	if err != nil {
		panic(err)
	}

	// then build each row by month
	rows := [][]string{}
	for month, amountByCategory := range calculations {
		row := []string{month}
		for _, category := range tableCategoriesHeader[1:] {
			row = append(row, fmt.Sprintf("%.2f€", amountByCategory[category]))
		}
		rows = append(rows, row)
	}

	data := pterm.TableData{tableCategoriesHeader}
	data = append(data, rows...)
	pterm.NewRGB(15, 199, 209).Printf("Overall per %s\n", grouping)
	pterm.Println()
	err = pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	if err != nil {
		panic(err)
	}
}
