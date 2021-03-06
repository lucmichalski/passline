package input

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	ucli "github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func ArgOrInput(args ucli.Args, index int, message string, defaultValue string) (string, error) {
	userInput := args.Get(index)

	if userInput == "" {
		message := fmt.Sprintf("Please enter a %s", message)
		if defaultValue != "" {
			message += " (%s): "
		} else {
			message += " []: "
		}

		var err error
		userInput, err = Default(message, defaultValue)
		if err != nil {
			return "", err
		}
	}

	return userInput, nil
}

func Default(message string, defaultValue string) (string, error) {
	// find if %s is in string
	rgx := regexp.MustCompile("%s")
	matches := rgx.FindAllStringIndex(message, -1)

	// print message
	if len(matches) == 0 {
		fmt.Print(message)
	} else {
		fmt.Printf(message, defaultValue)
	}

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	text = strings.TrimSuffix(text, "\n")
	// TODO Test if working for Linux
	text = strings.TrimSuffix(text, "\r")

	// if input empty assign default value also valid if defaultValue is empty
	if text == "" {
		text = defaultValue
	}
	return text, nil
}

func Confirmation(message string) bool {
	result := ""
	for result != "y" && result != "n" {
		var err error
		result, err = Default(message, "")
		if err != nil {
			return false
		}
	}

	return result == "y"
}

func Password(message string) []byte {
	// Now get the password.
	fmt.Print(message)
	p, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		if runtime.GOOS != "windows" {
			panic(err)
		}
	}

	// Return the password as a string.
	return p
}
