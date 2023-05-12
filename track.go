package analytics

import "time"

var _ Message = (*Track)(nil)

type Track struct {
	// This field is exported for serialization purposes and shouldn't be set by
	// the application, its value is always overwritten by the library.
	Type           int    `json:"t"`
	PayloadVersion int    `json:"v"`
	Project        string `json:"p"`

	Timestamp    time.Time `json:"ts"`
	MessageId    string    `json:"mid"`
	InstanceId   string    `json:"iid"`
	DeploymentId string    `json:"did"`

	OsName         string `json:"osn,omitempty"`
	OsArchitecture string `json:"osa,omitempty"`
	CPU            int    `json:"cpu,omitempty"`

	// Alloc is bytes of allocated heap objects.
	Alloc uint64 `json:"alloc"`
	// TotalAlloc is cumulative bytes allocated for heap objects.
	TotalAlloc uint64 `json:"totalAlloc"`
	// Sys is the total bytes of memory obtained from the OS.
	Sys uint64 `json:"sys"`
	// Lookups is the number of pointer lookups performed by the
	// runtime.
	Lookups uint64 `json:"lookups"`
	// Mallocs is the cumulative count of heap objects allocated.
	// The number of live objects is Mallocs - Frees.
	Mallocs uint64 `json:"mallocs"`
	// Frees is the cumulative count of heap objects freed.
	Frees uint64 `json:"frees"`
	// HeapAlloc is bytes of allocated heap objects.
	HeapAlloc uint64 `json:"heapAlloc"`
	// HeapSys is bytes of heap memory obtained from the OS.
	HeapSys uint64 `json:"heapSys"`
	// HeapIdle is bytes in idle (unused) spans.
	HeapIdle uint64 `json:"heapIdle"`
	// HeapInuse is bytes in in-use spans.
	HeapInuse uint64 `json:"heapInuse"`
	// HeapReleased is bytes of physical memory returned to the OS.
	HeapReleased uint64 `json:"heapReleased"`
	// HeapObjects is the number of allocated heap objects.
	HeapObjects uint64 `json:"heapObjects"`
	// NumGC is the number of completed GC cycles.
	NumGC uint32 `json:"numGC"`
}

func (msg Track) internal() {
	panic(unimplementedError)
}

func (msg Track) Validate() error {
	if len(msg.InstanceId) == 0 {
		return FieldError{
			Type:  "analytics.Track",
			Name:  "InstanceId",
			Value: msg.InstanceId,
		}
	}

	if len(msg.DeploymentId) == 0 {
		return FieldError{
			Type:  "analytics.Track",
			Name:  "DeploymentId",
			Value: msg.DeploymentId,
		}
	}

	return nil
}
