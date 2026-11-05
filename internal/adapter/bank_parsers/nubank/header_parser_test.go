package nubank

import "testing"



func TestNubankParserIMPL_Supports(t *testing.T) {


	testCases := []struct {
		name   string
		header []string
		expectedColumns map[string]int
	}{
		{
			name:   "supports Nubank header",
			header: []string{"Data", "Descrição", "Valor", "Identificador"},
			expectedColumns: map[string]int{
				"date":        0,
				"description": 1,
				"amount":      2,
				"external_id": 3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewNubankParserIMPL(&tc.header)
			if err != nil {
				t.Fatalf("Failed to create NubankParserIMPL: %v", err)
			}

			if !parser.Supports() {
				t.Errorf("Expected Supports() to return true, but got false")
			}

			if len(parser.columns) != len(tc.expectedColumns) {
				t.Errorf("Expected columns length to be %d, but got %d", len(tc.expectedColumns), len(parser.columns))
			}

			for key, expectedIndex := range tc.expectedColumns {
				if index, exists := parser.columns[key]; !exists || index != expectedIndex {
					t.Errorf("Expected column '%s' to have index %d, but got %d", key, expectedIndex, index)
				}
			}
		})
	}

}


func TestNubankParserIMPL_GetColumns(t *testing.T) {
	testCases := []struct {
		name   string
		header []string
		expectedColumns map[string]int
	}{
		{
			name:   "get columns from Nubank header",
			header: []string{"Data", "Descrição", "Valor", "Identificador"},
			expectedColumns: map[string]int{
				"date":        0,
				"description": 1,
				"amount":      2,
				"external_id": 3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewNubankParserIMPL(&tc.header)
			if err != nil {
				t.Fatalf("Failed to create NubankParserIMPL: %v", err)
			}

			if !parser.Supports() {
				t.Errorf("Expected Supports() to return true, but got false")
			}

			for key, expectedIndex := range tc.expectedColumns {
				if index, exists := parser.GetColumnIndex(key); !exists || index != expectedIndex {
					t.Errorf("Expected column '%s' to have index %d, but got %d", key, expectedIndex, index)
				}
			}
		})
	}
}