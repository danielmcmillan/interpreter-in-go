package repl

import (
	"bufio"
	"fmt"
	"io"

	"danielmcm.com/interpreterbook/lexer"
	"danielmcm.com/interpreterbook/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		parser := parser.New(lexer)

		program := parser.ParseProgram()
		errors := parser.Errors()
		if len(errors) > 0 {
			printParserErrors(out, errors)
		} else {
			fmt.Fprint(out, program.String())
		}
	}
}

func printParserErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		fmt.Fprintf(out, "Error: %s\n", err)
	}
}
