package k8s

import (
	"testing"
)

func TestValidationWebhook(t *testing.T) {
	w := &webhookValidation{}
	w.startHTTPServer()
}
