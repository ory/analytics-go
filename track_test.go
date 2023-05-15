package analytics

import "testing"

func TestTrackValid(t *testing.T) {
	page := Identify{
		Type:         2,
		DeploymentId: "TEST",
		InstanceId:   "TEST",
	}

	if err := page.Validate(); err != nil {
		t.Error("validating a valid track object failed:", page, err)
	}
}

func TestTrackMissingInstanceId(t *testing.T) {
	track := Track{
		Type:         2,
		DeploymentId: "TEST",
	}

	if err := track.Validate(); err == nil {
		t.Error("validating an invalid track object succeeded:", track)

	} else if e, ok := err.(FieldError); !ok {
		t.Error("invalid error type returned when validating track:", err)

	} else if e != (FieldError{
		Type:  "analytics.Track",
		Name:  "InstanceId",
		Value: "",
	}) {
		t.Error("invalid error value returned when validating track:", err)
	}
}

func TestTrackMissingDeploymentId(t *testing.T) {
	track := Track{
		Type:       2,
		InstanceId: "TEST",
	}

	if err := track.Validate(); err == nil {
		t.Error("validating an invalid track object succeeded:", track)

	} else if e, ok := err.(FieldError); !ok {
		t.Error("invalid error type returned when validating track:", err)

	} else if e != (FieldError{
		Type:  "analytics.Track",
		Name:  "DeploymentId",
		Value: "",
	}) {
		t.Error("invalid error value returned when validating track:", err)
	}
}
