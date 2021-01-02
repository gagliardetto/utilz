package utilz

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// CLIAskYesNo parses an input from the terminal and returns whether the
// response is affirmative or negative.
func CLIAskYesNo(message string) (bool, error) {
	fmt.Println()
	fmt.Println(message, "[y/n]")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		if err.Error() == "unexpected newline" {
			return CLIAskYesNo(message)
		}
		return false, err
	}

	// clean the input
	input = strings.ToLower(input)
	input = strings.TrimSpace(input)

	okayResponses := []string{"y", "yes"}
	nokayResponses := []string{"n", "no"}

	if SliceContains(okayResponses, input) {
		return true, nil
	} else if SliceContains(nokayResponses, input) {
		return false, nil
	} else {
		fmt.Println("Not recognized. Please type yes/no or y/n and then press enter.")
		return CLIAskYesNo(message)
	}
}

// CLIAskPassword prompts the user for a password input fro the CLI
func CLIAskPassword() ([]byte, error) {
	return terminal.ReadPassword(int(os.Stdin.Fd()))
}

// CLIAskString prompts the user for a string input from the CLI
func CLIAskString() (string, error) {
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		fmt.Println("fatal: ", err) //TODO: handle error
		return "", err
	}

	return input, nil
}

type FlagStringArray []string

func (i *FlagStringArray) String() string {
	return strings.Join(*i, ",")
}

func (i *FlagStringArray) Set(value string) error {
	value = strings.TrimSpace(value)
	if value != "" {
		*i = append(*i, value)
	}
	// TODO: sort by len here?
	return nil
}
func CLIMustConfirmYes(message string) {
	doContinue, err := CLIAskYesNo(message)
	if err != nil {
		panic(err)
	}
	if !doContinue {
		Ln(Orange("Aborting"))
		os.Exit(0)
	}
}
