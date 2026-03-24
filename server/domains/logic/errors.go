package logic

import (
	"fmt"
	"strings"
)

// HumanizeError takes a technical error and returns a human-friendly explanation and a potential fix.
func HumanizeError(err error) (explanation string, fix string) {
	if err == nil {
		return "", ""
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "dial tcp"):
		return "The system couldn't connect to the external service.", "Check if the service URL is correct and if the service is online."
	case strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized"):
		return "Access was denied by the external service.", "Verify that your API keys or credentials are correct."
	case strings.Contains(msg, "404"):
		return "The requested resource was not found on the external service.", "Double-check the URL or resource ID you provided."
	case strings.Contains(msg, "condition evaluation failed"):
		return "The system couldn't evaluate one of the flow conditions.", "Ensure all variables used in the condition (e.g., amount > 100) are actually defined in the process."
	case strings.Contains(msg, "no handler found"):
		return "The system doesn't know how to handle this node type.", "This might be a bug or an unsupported element type. Contact your administrator."
	case strings.Contains(msg, "variable not found"):
		return fmt.Sprintf("A required variable is missing: %s", msg), "Ensure the variable is set in a previous step or provided when starting the process."
	default:
		return msg, "Check the logs or contact support for more details."
	}
}
