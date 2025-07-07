package game

import (
	"fmt"
)

type Notifier interface {
	Notify(userID string, message string) error
}

type EmailNotifier struct{}

func (e *EmailNotifier) Notify(userID string, message string) error {
	// In a real application, this would send an email.
	// For now, we'll just print to console.
	fmt.Printf("Sending email to user %s: %s\n", userID, message)
	return nil
}
