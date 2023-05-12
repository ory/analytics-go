package analytics

import "time"

var _ Message = (*Page)(nil)

type Page struct {
	// This field is exported for serialization purposes and shouldn't be set by
	// the application, its value is always overwritten by the library.
	Type           int    `json:"t"`
	PayloadVersion int    `json:"v"`
	Project        string `json:"p"`

	Timestamp    time.Time `json:"ts"`
	MessageId    string    `json:"mid"`
	InstanceId   string    `json:"iid"`
	DeploymentId string    `json:"did"`

	UrlHost        string  `json:"uh"`
	UrlPath        string  `json:"up"`
	UrlScheme      int     `json:"us,omitempty"`
	RequestCode    int     `json:"rc"`
	RequestLatency float64 `json:"rl"`
}

func (msg Page) internal() {
	panic(unimplementedError)
}

func (msg Page) Validate() error {
	if len(msg.InstanceId) == 0 {
		return FieldError{
			Type:  "analytics.Page",
			Name:  "InstanceId",
			Value: msg.InstanceId,
		}
	}

	if len(msg.DeploymentId) == 0 {
		return FieldError{
			Type:  "analytics.Page",
			Name:  "DeploymentId",
			Value: msg.DeploymentId,
		}
	}

	return nil
}
