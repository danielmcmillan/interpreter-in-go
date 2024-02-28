package repl

import (
	"bufio"
	"fmt"
	"io"

	"danielmcm.com/interpreterbook/lexer"
	"danielmcm.com/interpreterbook/token"
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

		for nextToken := lexer.NextToken(); nextToken.Type != token.EOF; nextToken = lexer.NextToken() {
			fmt.Fprintf(out, "%+v\n", nextToken)
		}
	}
}
