package analytics

import "testing"

func TestPageValid(t *testing.T) {
	page := Page{
		Type:         3,
		DeploymentId: "TEST",
		InstanceId:   "TEST",
	}

	if err := page.Validate(); err != nil {
		t.Error("validating a valid page object failed:", page, err)
	}
}

func TestPageMissingInstanceId(t *testing.T) {
	page := Page{
		Type:         3,
		DeploymentId: "TEST",
	}

	if err := page.Validate(); err == nil {
		t.Error("validating an invalid page object succeeded:", page)

	} else if e, ok := err.(FieldError); !ok {
		t.Error("invalid error type returned when validating page:", err)

	} else if e != (FieldError{
		Type:  "analytics.Page",
		Name:  "InstanceId",
		Value: "",
	}) {
		t.Error("invalid error value returned when validating page:", err)
	}
}

func TestPageMissingDeploymentId(t *testing.T) {
	page := Page{
		Type:       3,
		InstanceId: "TEST",
	}

	if err := page.Validate(); err == nil {
		t.Error("validating an invalid page object succeeded:", page)

	} else if e, ok := err.(FieldError); !ok {
		t.Error("invalid error type returned when validating page:", err)

	} else if e != (FieldError{
		Type:  "analytics.Page",
		Name:  "DeploymentId",
		Value: "",
	}) {
		t.Error("invalid error value returned when validating page:", err)
	}
}
