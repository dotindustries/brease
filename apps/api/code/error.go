package code

import (
	"fmt"
	"github.com/fatih/color"
	"regexp"
	"strconv"
	"strings"
)

func extractCodeSection(errorMessage string, codeBlock string, numLines int) string {
	lines := strings.Split(errorMessage, "\n")
	lineNumber, charNumber := extractLineCharNumber(lines[len(lines)-1])
	codeSection := ""

	if lineNumber > 0 {
		codeLines := strings.Split(codeBlock, "\n")
		startLine := lineNumber - numLines - 1
		endLine := lineNumber + numLines - 1

		for i := startLine; i <= endLine; i++ {
			if i >= 0 && i < len(codeLines) {
				lnPrefix := fmt.Sprintf("%d: ", i+1)
				line := fmt.Sprintf("%s%s", lnPrefix, codeLines[i])
				codeSection += line + "\n"

				if i == lineNumber-1 {
					// Mark the error position with a '^' symbol
					if charNumber > 0 && charNumber <= len(codeLines[i])+1 {
						errorLineMarker := strings.Repeat(" ", len(lnPrefix)+charNumber-1) + color.RedString("^ %s", lines[0])
						codeSection += errorLineMarker + "\n"
					}
				}
			}
		}
	}

	return codeSection
}

func extractLineCharNumber(errorLine string) (lineNumber int, charNumber int) {
	re := regexp.MustCompile(`:(\d+):(\d+)`)
	match := re.FindStringSubmatch(errorLine)
	if len(match) >= 3 {
		lineNumber, _ = strconv.Atoi(match[1])
		charNumber, _ = strconv.Atoi(match[2])
	}
	return
}
