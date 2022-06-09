package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

func Success(message ...string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s - %v\n", green("PASS"), strings.Join(message, " "))
}

func Warn(message ...string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("%s - %v\n", yellow("WARN"), strings.Join(message, " "))
}

func Error(message error) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Printf("%s  %v\n", red("ERR "), message)
}

func Fatal(message error) {
	Error(message)
	os.Exit(1)
}
