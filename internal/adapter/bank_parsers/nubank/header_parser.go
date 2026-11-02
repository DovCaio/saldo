package nubank

import (
	"strings"
)


var headerMapping = map[string]string{
    "Data":          "date",
    "Descrição":     "description",
    "Valor":         "amount",
    "Identificador": "type",
}

type NubankParserIMPL struct {
    content *[][]string
    columns map[string]int
}

func NewNubankParserIMPL(content *[][]string) (*NubankParserIMPL, error) {
    
    return &NubankParserIMPL{
        content: content,
        columns: make(map[string]int),
    }, nil
}



func (p *NubankParserIMPL) Supports() bool {
    expected := map[string]bool{
        "date":        true,
        "description": true,
        "amount":      true,
        "type":        true,
    }

    header := (*p.content)[0]

    if len(header) != len(expected) {
        return false
    }

    for i, header := range header {
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

func (p *NubankParserIMPL) Parse() error {
    header := (*p.content)[0]

    for i, h := range header {
        header[i] = headerMapping[strings.TrimSpace(h)]
    }

    return nil
}