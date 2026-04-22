package wizard

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Pimatis/mavetis/src/config"
)

func RunInteractive(reader *bufio.Reader, project Project) ConfigTemplate {
	template := ConfigTemplate{
		Profile:    project.Profile,
		Severity:   "low",
		FailOn:     "high",
		Output:     "text",
		Ignore:     project.Ignore,
		Critical:   project.Critical,
		Restricted: project.Restricted,
	}
	fmt.Println("Press Enter to accept defaults.")

	profile := prompt(reader, "Profile (auth, backend, frontend, fintech)", project.Profile)
	if config.ValidateProfile(profile) == nil {
		template.Profile = profile
	}

	failOn := prompt(reader, "Fail-on threshold (low, medium, high, critical)", "high")
	if config.ValidateSeverity(failOn, "fail-on") == nil {
		template.FailOn = failOn
	}

	output := prompt(reader, "Output format (text, json, sarif)", "text")
	if config.ValidateOutput(output) == nil {
		template.Output = output
	}

	if len(project.Critical) > 0 {
		fmt.Println("\nDetected critical zones:")
		for _, z := range project.Critical {
			fmt.Printf("  - %s\n", z)
		}
		if !promptYesNo(reader, "Use detected critical zones?", true) {
			template.Critical = nil
		}
	}

	if len(project.Restricted) > 0 {
		fmt.Println("\nDetected restricted zones:")
		for _, z := range project.Restricted {
			fmt.Printf("  - %s\n", z)
		}
		if !promptYesNo(reader, "Use detected restricted zones?", true) {
			template.Restricted = nil
		}
	}

	if len(project.Ignore) > 0 {
		fmt.Println("\nSuggested ignore patterns:")
		for _, i := range project.Ignore {
			fmt.Printf("  - %s\n", i)
		}
		if !promptYesNo(reader, "Use suggested ignore patterns?", true) {
			template.Ignore = nil
		}
	}

	return template
}

func prompt(reader *bufio.Reader, question, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", question, defaultValue)
	} else {
		fmt.Printf("%s: ", question)
	}
	text, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return defaultValue
		}
		return defaultValue
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultValue
	}
	return text
}

func promptYesNo(reader *bufio.Reader, question string, defaultYes bool) bool {
	defaultStr := "Y/n"
	if !defaultYes {
		defaultStr = "y/N"
	}
	fmt.Printf("%s [%s]: ", question, defaultStr)
	text, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return defaultYes
		}
		return defaultYes
	}
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return defaultYes
	}
	return text == "y" || text == "yes"
}
