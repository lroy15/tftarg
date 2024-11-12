package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

func formatString(input string) string {
	noQuotes := strings.ReplaceAll(input, "\"", "")
	formatted := strings.ReplaceAll(noQuotes, " ", ".")
	return formatted
}

func gitBranchDiff() (string, error) {
	var outBuffer bytes.Buffer

	masterDiff := exec.Command("git", "diff", "master...HEAD")
	masterDiffOut, err := masterDiff.Output()
	if err != nil {
		return "", err
	}
	outBuffer.Write(masterDiffOut)

	unstagedDiff := exec.Command("git", "diff")
	unstagedDiffOut, err := unstagedDiff.Output()
	if err != nil {
		return "", err
	}
	outBuffer.Write(unstagedDiffOut)

	return outBuffer.String(), nil
}

func extractTopLevel(diff string) []string {
	re := regexp.MustCompile(`\b(resource|module)\b\s+"([^"]+)"(?:\s+"([^"]+)")?`)
	matches := re.FindAllStringSubmatch(diff, -1)

	uniqueResources := make(map[string]struct{})
	for _, match := range matches {
		if match[3] == "" {
			prepared := match[1] + "." + match[2]
			uniqueResources[prepared] = struct{}{}
		} else {
			prepared := match[1] + "." + match[2] + "." + match[3]
			uniqueResources[prepared] = struct{}{}
		}
	}
	var tfResources []string
	for tfResource := range uniqueResources {
		tfResources = append(tfResources, tfResource)
	}

	return tfResources
}

func main() {
	diff, err := gitBranchDiff()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	tfResources := extractTopLevel(diff)

	selectedResources := []string{}
	command := ""

	options := huh.NewOptions(tfResources...)
	form := huh.NewForm(
		huh.NewGroup(huh.NewMultiSelect[string]().
			Title("Choose an option").
			Options(options...).
			Value(&selectedResources),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Command").
				Options(
					huh.NewOption("plan", "plan"),
					huh.NewOption("apply", "apply"),
					huh.NewOption("print", "print"),
				).
				Value(&command),
		),
	)
	err = form.Run()
	if err != nil {
		log.Fatalf("aborted the prompt")
	}

	commandString := []string{}

	for _, item := range selectedResources {
		commandString = append(commandString, "-target="+formatString(item))
	}

	fullCommandString := append([]string{command}, commandString...)

	if command == "print" {
		anySlice := make([]any, len(commandString))
		for i, v := range commandString {
			anySlice[i] = v
		}
		fmt.Println(anySlice...)
		os.Exit(0)
	}
	tfcommand := exec.Command("terraform", fullCommandString...)
	fmt.Printf("Executing command: %s %s\n", tfcommand.Path, strings.Join(tfcommand.Args[1:], " "))

	tfcommand.Dir = "./"
	tfcommand.Stdin = os.Stdin
	tfcommand.Stdout = os.Stdout
	tfcommand.Stderr = os.Stderr

	err = tfcommand.Run()
	if err != nil {
		log.Fatalf("Error running terraform command: %v", err)
	}
}
