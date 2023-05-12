package analytics

import "time"

var _ Message = (*Identify)(nil)

type Identify struct {
	// This field is exported for serialization purposes and shouldn't be set by
	// the application, its value is always overwritten by the library.
	Type           int    `json:"t"`
	PayloadVersion int    `json:"v"`
	Project        string `json:"p"`

	Timestamp    time.Time `json:"ts"`
	MessageId    string    `json:"mid"`
	InstanceId   string    `json:"iid"`
	DeploymentId string    `json:"did"`

	DatabaseDialect  string  `json:"dbd,omitempty"`
	ProductVersion   string  `json:"pv,omitempty"`
	ProductBuild     string  `json:"pb,omitempty"`
	UptimeDeployment float64 `json:"dup"`
	UptimeInstance   float64 `json:"iup"`
	IsDevelopment    bool    `json:"isd"`
	IsOptOut         bool    `json:"iso"`
	Startup          bool    `json:"s"`
}

func (msg Identify) internal() {
	panic(unimplementedError)
}

func (msg Identify) Validate() error {
	if len(msg.InstanceId) == 0 {
		return FieldError{
			Type:  "analytics.Identify",
			Name:  "InstanceId",
			Value: msg.InstanceId,
		}
	}

	if len(msg.DeploymentId) == 0 {
		return FieldError{
			Type:  "analytics.Identify",
			Name:  "DeploymentId",
			Value: msg.DeploymentId,
		}
	}

	return nil
}
