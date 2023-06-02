package main

import (
	"fmt"
	"github.com/hennedo/escpos"
	"strconv"
	"strings"
)

type templateState int

const (
	None templateState = iota
	Content
	AfterContent
	Command
)

type command string

const (
	Text     command = ""
	Bold     command = "BOLD"
	Reverse  command = "REVERSE"
	FontSize command = "FONTSIZE"
	Barcode  command = "BARCODE"
	QRCode   command = "QRCODE"
	Cut      command = "CUT"
)

var barcodeTypes = map[string]func(*escpos.Escpos, string) (int, error){
	"UPCA":  (*escpos.Escpos).UPCA,
	"UPCE":  (*escpos.Escpos).UPCE,
	"EAN13": (*escpos.Escpos).EAN13,
	"EAN8":  (*escpos.Escpos).EAN8,
}

func (c command) validate(args []string, text string) (cmdArgs []any, err error) {
	switch c {
	case Bold, Reverse, FontSize, Cut:
		if text != "" {
			return nil, fmt.Errorf("command %q does not allow text argument", c)
		}
	case Barcode, QRCode, Text:
		if text == "" {
			return nil, fmt.Errorf("command %q requires text argument", c)
		}
	default:
		return nil, fmt.Errorf("unknown command: %q", c)
	}

	switch c {
	case Bold, Reverse, Text, Cut:
		if len(args) != 0 {
			return nil, fmt.Errorf("command %q does not allow arguments", c)
		}
	case FontSize:
		for i, arg := range args {
			switch i {
			case 0: //x-size 0-5
				n, err := strconv.Atoi(arg)
				if err != nil {
					return nil, err
				}
				cmdArgs = append(cmdArgs, n)
			case 1: //y-size 0-5
				n, err := strconv.Atoi(arg)
				if err != nil {
					return nil, err
				}
				cmdArgs = append(cmdArgs, n)
			default:
				return nil, fmt.Errorf("invalid argument count for %q", c)
			}
		}

		// if there is only one size argument given, copy it to make the font symmetric
		if len(cmdArgs) == 1 {
			cmdArgs = append(cmdArgs, cmdArgs[0])
		}

	case Barcode:
		for i, arg := range args {
			switch i {
			case 0: //type UPCA, UPCE, EAN13, EAN8
				t := strings.ToUpper(arg)
				if _, ok := barcodeTypes[t]; !ok {
					return nil, fmt.Errorf("unknown barcode type: %s", t)
				}

				cmdArgs = append(cmdArgs, t)
			case 1: //x-size 2-6
				n, err := strconv.Atoi(arg)
				if err != nil {
					return nil, err
				}
				cmdArgs = append(cmdArgs, n)
			case 2: //y-size 2-6
				n, err := strconv.Atoi(arg)
				if err != nil {
					return nil, err
				}
				cmdArgs = append(cmdArgs, n)
			default:
				return nil, fmt.Errorf("invalid argument count for %q", c)
			}
		}

		// if there is only one size argument given, copy it to make the barcode symmetric
		if len(cmdArgs) == 2 {
			cmdArgs = append(cmdArgs, cmdArgs[1])
		}

	case QRCode:
		if len(args) != 3 {
			return nil, fmt.Errorf("%s requires three arguments", c)
		}

		for i, arg := range args {
			switch i {
			case 0: // model 1, 2
				n, err := strconv.Atoi(arg)
				if err != nil {
					return nil, err
				}
				cmdArgs = append(cmdArgs, n)
			case 1: // size 1-16
				n, err := strconv.Atoi(arg)
				if err != nil {
					return nil, err
				}
				cmdArgs = append(cmdArgs, n)
			case 2: // correction 48-51
				n, err := strconv.Atoi(arg)
				if err != nil {
					return nil, err
				}
				cmdArgs = append(cmdArgs, n)
			default:
				return nil, fmt.Errorf("invalid argument count for %q", c)
			}
		}

	default:
		return nil, fmt.Errorf("command %q arguments not implemented", c)
	}

	return
}

type Template []templatePart

func (t Template) Execute(p *escpos.Escpos) error {
	for _, part := range t {
		switch part.Command {
		case Text:
			// Nothing
		case Bold:
			p.Style.Bold = !p.Style.Bold
			continue
		case Reverse:
			p.Style.Reverse = !p.Style.Reverse
			continue
		case FontSize:
			p.Style.Height = uint8(part.Arguments[0].(int))
			p.Style.Width = uint8(part.Arguments[1].(int))
			continue
		case Barcode:
			fn := barcodeTypes[part.Arguments[0].(string)]
			if _, err := p.BarcodeWidth(uint8(part.Arguments[1].(int))); err != nil {
				return err
			}
			if _, err := p.BarcodeHeight(uint8(part.Arguments[2].(int))); err != nil {
				return err
			}
			if _, err := fn(p, part.Content); err != nil {
				return err
			}

			continue
		case QRCode:
			model := false
			if part.Arguments[0] == 2 {
				model = true
			}
			size := uint8(part.Arguments[1].(int))
			errCorr := uint8(part.Arguments[2].(int))
			if _, err := p.QRCode(part.Content, model, size, errCorr); err != nil {
				return err
			}

			continue
		case Cut:
			if err := p.PrintAndCut(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown command: %s", part.Command)
		}

		for _, s := range strings.Split(part.Content, "\n") {
			if _, err := p.WriteWEU(s); err != nil {
				return err
			}

			if _, err := p.LineFeed(); err != nil {
				return err
			}
		}
	}

	if err := p.Print(); err != nil {
		return err
	}

	return nil
}

type templatePart struct {
	Command   command
	Arguments []any
	Content   string
}

type parsePart struct {
	command string
	text    string
	params  []string
}

func (p parsePart) empty() bool {
	return p.command == "" && p.text == "" && len(p.params) == 0
}

func (p parsePart) toTemplatePart() (tp templatePart, err error) {
	cmdParts := strings.Split(p.command, ",")
	tp.Command = command(cmdParts[0])
	tp.Arguments, err = tp.Command.validate(cmdParts[1:], p.text)
	tp.Content = p.text
	return
}

type templateParser struct {
	curr  parsePart
	parts []templatePart
	state templateState
}

func (tp *templateParser) next() error {
	if tp.state != None {
		return fmt.Errorf("invalid state: %v", tp.state)
	}

	if tp.curr.empty() {
		return nil
	}

	t, err := tp.curr.toTemplatePart()
	if err != nil {
		return err
	}

	tp.parts = append(tp.parts, t)
	tp.curr = parsePart{}

	return nil
}

func (tp *templateParser) parse(s string) error {
	for _, char := range s {
		switch char {
		case '[':
			if tp.state != None {
				panic("Invalid Template")
			}

			if err := tp.next(); err != nil {
				return err
			}

			tp.state = Content
			continue
		case ']':
			if tp.state != Content {
				panic("Invalid Template")
			}

			tp.state = AfterContent
			continue
		case '(':
			if tp.state != AfterContent {
				panic("Invalid Template")
			}

			tp.state = Command
			continue
		case ')':
			if tp.state != Command {
				panic("Invalid Template")
			}

			tp.state = None
			if err := tp.next(); err != nil {
				return err
			}

			continue
		}

		switch tp.state {
		case None, Content:
			tp.curr.text += string(char)
		case Command:
			tp.curr.command += string(char)
		default:
			return fmt.Errorf("invalid state: %v", tp.state)
		}
	}

	return tp.next()
}

func ParseTemplate(s string) (Template, error) {
	var tp templateParser
	return tp.parts, tp.parse(s)
}
