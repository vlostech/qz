package ioext

import (
	"bufio"
	"os"
	"strings"
)

// GetMultilineString returns a string from stdin that might consist of multiple lines.
func GetMultilineString() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	var strList []string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "\\end" {
			break
		}

		line, hasSeparator := strings.CutSuffix(line, "\\")

		strList = append(strList, line)

		if !hasSeparator {
			break
		}
	}

	return strings.Join(strList, "\n"), nil
}
