package nubank

import (
	"strings"
)


var headerMapping = map[string]string{
    "Data":          "date",
    "Descrição":     "description",
    "Valor":         "amount",
    "Identificador": "external_id", //This is wrong, the type depends of the sinal of the amount, and the indentification is exaternal identification
}

type NubankParserIMPL struct {
    header *[]string
    columns map[string]int
}

func NewNubankParserIMPL(header *[]string) (*NubankParserIMPL, error) {
    
    return &NubankParserIMPL{
        header: header,
        columns: make(map[string]int),
    }, nil
}



func (p *NubankParserIMPL) Supports() bool {
    expected := map[string]bool{
        "date":        true,
        "description": true,
        "amount":      true,
        "external_id":        true,
    }


    if len(*p.header) != len(expected) {
        return false
    }

    for i, header := range *p.header {
        normalized := headerMapping[strings.TrimSpace(header)]

        if !expected[normalized] {
            return false
        }

        if _, exists := p.columns[normalized]; exists {
            return false
        }

        p.columns[normalized] = i
    }

    return len(p.columns) == len(expected)
}


func (p *NubankParserIMPL) GetColumnIndex(columnName string) (int, bool) {
    index, exists := p.columns[columnName]
    return index, exists
}
