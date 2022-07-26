package main

import (
	"encoding/csv"
	"log"
	"os"

	"github.com/manzanit0/mcduck/pkg/expense"
	"github.com/manzanit0/mcduck/pkg/term"
)

func main() {
	f, err := os.Open("./example_input.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = 4
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// skip header
	expenses, err := expense.NewExpenses(data[1:])
	if err != nil {
		log.Fatal(err)
	}

	categoryTotals := expense.CalculateTotalsPerCategory(expenses)
	err = term.Render("category", categoryTotals)
	if err != nil {
		log.Fatal(err)
	}

	subcategoryTotals := expense.CalculateTotalsPerSubCategory(expenses)
	err = term.Render("sub-category", subcategoryTotals)
	if err != nil {
		log.Fatal(err)
	}
}
