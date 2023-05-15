package analytics

import "testing"

func TestIdentifyValid(t *testing.T) {
	page := Identify{
		Type:         1,
		DeploymentId: "TEST",
		InstanceId:   "TEST",
	}

	if err := page.Validate(); err != nil {
		t.Error("validating a valid identify object failed:", page, err)
	}
}

func TestIdentifyMissingInstanceId(t *testing.T) {
	identify := Identify{
		Type:         1,
		DeploymentId: "TEST",
	}

	if err := identify.Validate(); err == nil {
		t.Error("validating a valid identify object failed:", identify, err)

	} else if e, ok := err.(FieldError); !ok {
		t.Error("invalid error type returned when validating identify:", err)

	} else if e != (FieldError{
		Type:  "analytics.Identify",
		Name:  "InstanceId",
		Value: "",
	}) {
		t.Error("invalid error value returned when validating identify:", err)
	}
}

func TestIdentifyMissingDeploymentId(t *testing.T) {
	identify := Identify{
		Type:       1,
		InstanceId: "TEST",
	}

	if err := identify.Validate(); err == nil {
		t.Error("validating a valid identify object failed:", identify, err)

	} else if e, ok := err.(FieldError); !ok {
		t.Error("invalid error type returned when validating identify:", err)

	} else if e != (FieldError{
		Type:  "analytics.Identify",
		Name:  "DeploymentId",
		Value: "",
	}) {
		t.Error("invalid error value returned when validating identify:", err)
	}
}
