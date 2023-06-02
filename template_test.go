package main

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func Test_ParseTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateString string
		res            Template
	}{
		{name: "none", templateString: "none", res: Template{{Content: "none"}}},
		{name: "bold", templateString: "[](BOLD)Bold", res: Template{{Command: Bold}, {Content: "Bold"}}},
		{name: "bold_reverse_bold", templateString: "[](BOLD)[](REVERSE)Bold[](BOLD)", res: Template{{Command: Bold}, {Command: Reverse}, {Content: "Bold"}, {Command: Bold}}},
		{name: "fontsize_with_content", templateString: "[](FONTSIZE,5)Text", res: Template{{Command: FontSize, Arguments: []any{5, 5}}, {Content: "Text"}}},
		{name: "fontsize_asymmetric_with_content", templateString: "[](FONTSIZE,5,1)Text", res: Template{{Command: FontSize, Arguments: []any{5, 1}}, {Content: "Text"}}},
		{name: "barcode_symmetric", templateString: "[CONTENT](BARCODE,UPCA,2)Text", res: Template{{Content: "CONTENT", Command: Barcode, Arguments: []any{"UPCA", 2, 2}}, {Content: "Text"}}},
		{name: "barcode_asymmetric", templateString: "[CONTENT](BARCODE,UPCA,1,2)Text", res: Template{{Content: "CONTENT", Command: Barcode, Arguments: []any{"UPCA", 1, 2}}, {Content: "Text"}}},
		{name: "qrcode", templateString: "[CONTENT](QRCODE,2,2,0)Text", res: Template{{Content: "CONTENT", Command: QRCode, Arguments: []any{2, 2, 0}}, {Content: "Text"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ParseTemplate(tt.templateString)
			if err != nil {
				t.Fatalf("parseTemplate(%q): %v", tt.templateString, err)
			}

			if diff := cmp.Diff(tt.res, res); diff != "" {
				t.Fatalf("parseTemplate(%q): %s", tt.templateString, diff)
			}
		})
	}
}
