// Code generated by woco, DO NOT EDIT.

package msg

import (
	"encoding/json"
	"fmt"
	"time"
)

type AlertStatus struct {
	InhibitedBy []string         `binding:"required" json:"inhibitedBy"`
	SilencedBy  []int            `binding:"required" json:"silencedBy"`
	State       AlertStatusState `binding:"required,oneof=unprocessed active suppressed" json:"state"`
}

type ClusterStatus struct {
	Name   string              `json:"name,omitempty"`
	Peers  []*PeerStatus       `json:"peers,omitempty"`
	Status ClusterStatusStatus `binding:"required,oneof=ready settling disabled" json:"status"`
}

type GettableAlert struct {
	*Alert `json:",inline"`
	// Annotations A set of labels. Labels are key/value pairs that are attached to
	// alerts. Labels are used to specify identifying attributes of alerts,
	// such as their tenant, user , instance, and job.
	// tenant: specific tenant id.
	// user: specific user id. the user is the notify target. Some notification need info from user, such as email address.
	// alertname: the name of alert.it is also the event name.
	Annotations LabelSet    `binding:"required" json:"annotations"`
	EndsAt      time.Time   `binding:"required" json:"endsAt" time_format:"2006-01-02T15:04:05Z07:00"`
	Fingerprint string      `binding:"required" json:"fingerprint"`
	Receivers   []*Receiver `binding:"required" json:"receivers"`
	StartsAt    time.Time   `binding:"required" json:"startsAt" time_format:"2006-01-02T15:04:05Z07:00"`
	Status      AlertStatus `binding:"required" json:"status"`
	UpdatedAt   time.Time   `binding:"required" json:"updatedAt" time_format:"2006-01-02T15:04:05Z07:00"`
}

// LabelSet A set of labels. Labels are key/value pairs that are attached to
// alerts. Labels are used to specify identifying attributes of alerts,
// such as their tenant, user , instance, and job.
// tenant: specific tenant id.
// user: specific user id. the user is the notify target. Some notification need info from user, such as email address.
// alertname: the name of alert.it is also the event name.
type LabelSet map[string]string

type Matcher struct {
	IsEqual bool   `json:"isEqual,omitempty"`
	IsRegex bool   `binding:"required" json:"isRegex"`
	Name    string `binding:"required" json:"name"`
	Value   string `binding:"required" json:"value"`
}

type PeerStatus struct {
	Address string `binding:"required" json:"address"`
	Name    string `binding:"required" json:"name"`
}

type Receiver struct {
	Name string `binding:"required" json:"name"`
}

type Silence struct {
	Comment   string    `binding:"required" json:"comment"`
	CreatedBy int       `binding:"required" json:"createdBy"`
	EndsAt    time.Time `binding:"required,gt" json:"endsAt" time_format:"2006-01-02T15:04:05Z07:00"`
	Matchers  Matchers  `binding:"required,min=1" json:"matchers"`
	StartsAt  time.Time `binding:"required,ltfield=EndsAt" json:"startsAt" time_format:"2006-01-02T15:04:05Z07:00"`
	TenantID  int       `binding:"required" json:"tenantID"`
}

type SilenceStatus struct {
	State SilenceStatusState `binding:"required,oneof=expired active pending" json:"state"`
}

type VersionInfo struct {
	Branch    string `binding:"required" json:"branch"`
	BuildDate string `binding:"required" json:"buildDate"`
	BuildUser string `binding:"required" json:"buildUser"`
	GoVersion string `binding:"required" json:"goVersion"`
	Revision  string `binding:"required" json:"revision"`
	Version   string `binding:"required" json:"version"`
}

type Alert struct {
	GeneratorURL string `binding:"omitempty,uri" json:"generatorURL,omitempty"`
	// Labels A set of labels. Labels are key/value pairs that are attached to
	// alerts. Labels are used to specify identifying attributes of alerts,
	// such as their tenant, user , instance, and job.
	// tenant: specific tenant id.
	// user: specific user id. the user is the notify target. Some notification need info from user, such as email address.
	// alertname: the name of alert.it is also the event name.
	Labels LabelSet `binding:"required" json:"labels"`
}

type AlertGroup struct {
	Alerts []*GettableAlert `binding:"required" json:"alerts"`
	// Labels A set of labels. Labels are key/value pairs that are attached to
	// alerts. Labels are used to specify identifying attributes of alerts,
	// such as their tenant, user , instance, and job.
	// tenant: specific tenant id.
	// user: specific user id. the user is the notify target. Some notification need info from user, such as email address.
	// alertname: the name of alert.it is also the event name.
	Labels   LabelSet `binding:"required" json:"labels"`
	Receiver Receiver `binding:"required" json:"receiver"`
}

type AlertGroups []*AlertGroup

type AlertmanagerConfig struct {
	Original string `binding:"required" json:"original"`
}

type AlertmanagerStatus struct {
	Cluster     ClusterStatus      `binding:"required" json:"cluster"`
	Config      AlertmanagerConfig `json:"config"`
	Uptime      time.Time          `binding:"required" json:"uptime" time_format:"2006-01-02T15:04:05Z07:00"`
	VersionInfo VersionInfo        `binding:"required" json:"versionInfo"`
}

type GettableAlerts []*GettableAlert

type GettableSilence struct {
	*Silence  `json:",inline"`
	ID        int           `binding:"required" json:"id"`
	Status    SilenceStatus `binding:"required" json:"status"`
	UpdatedAt time.Time     `binding:"required" json:"updatedAt" time_format:"2006-01-02T15:04:05Z07:00"`
}

type GettableSilences []*GettableSilence

type Matchers []*Matcher

type PostableAlert struct {
	*Alert `json:",inline"`
	// Annotations A set of labels. Labels are key/value pairs that are attached to
	// alerts. Labels are used to specify identifying attributes of alerts,
	// such as their tenant, user , instance, and job.
	// tenant: specific tenant id.
	// user: specific user id. the user is the notify target. Some notification need info from user, such as email address.
	// alertname: the name of alert.it is also the event name.
	Annotations LabelSet   `binding:"required" json:"annotations,omitempty"`
	EndsAt      *time.Time `json:"endsAt,omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
	StartsAt    *time.Time `json:"startsAt,omitempty" time_format:"2006-01-02T15:04:05Z07:00"`
}

type PostableAlerts []*PostableAlert

type PostableSilence struct {
	*Silence `json:",inline"`
	ID       int `json:"id,omitempty"`
}

// PushData Push data is for notify clients.
type PushData json.RawMessage

// AlertStatusState defines the type for the state.state enum field.
type AlertStatusState string

// AlertStatusState values.
const (
	AlertStatusStateUnprocessed AlertStatusState = "unprocessed"
	AlertStatusStateActive      AlertStatusState = "active"
	AlertStatusStateSuppressed  AlertStatusState = "suppressed"
)

func (s AlertStatusState) String() string {
	return string(s)
}

// AlertStatusStateValidator is a validator for the AlertStatusState field enum values.
func AlertStatusStateValidator(s AlertStatusState) error {
	switch s {
	case AlertStatusStateUnprocessed, AlertStatusStateActive, AlertStatusStateSuppressed:
		return nil
	default:
		return fmt.Errorf("AlertStatusState does not allow the value '%s'", s)
	}
}

// SilenceStatusState defines the type for the state.state enum field.
type SilenceStatusState string

// SilenceStatusState values.
const (
	SilenceStatusStateExpired SilenceStatusState = "expired"
	SilenceStatusStateActive  SilenceStatusState = "active"
	SilenceStatusStatePending SilenceStatusState = "pending"
)

func (s SilenceStatusState) String() string {
	return string(s)
}

// SilenceStatusStateValidator is a validator for the SilenceStatusState field enum values.
func SilenceStatusStateValidator(s SilenceStatusState) error {
	switch s {
	case SilenceStatusStateExpired, SilenceStatusStateActive, SilenceStatusStatePending:
		return nil
	default:
		return fmt.Errorf("SilenceStatusState does not allow the value '%s'", s)
	}
}

// ClusterStatusStatus defines the type for the status.status enum field.
type ClusterStatusStatus string

// ClusterStatusStatus values.
const (
	ClusterStatusStatusReady    ClusterStatusStatus = "ready"
	ClusterStatusStatusSettling ClusterStatusStatus = "settling"
	ClusterStatusStatusDisabled ClusterStatusStatus = "disabled"
)

func (s ClusterStatusStatus) String() string {
	return string(s)
}

// ClusterStatusStatusValidator is a validator for the ClusterStatusStatus field enum values.
func ClusterStatusStatusValidator(s ClusterStatusStatus) error {
	switch s {
	case ClusterStatusStatusReady, ClusterStatusStatusSettling, ClusterStatusStatusDisabled:
		return nil
	default:
		return fmt.Errorf("ClusterStatusStatus does not allow the value '%s'", s)
	}
}
