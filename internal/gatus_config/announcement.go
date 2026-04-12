package gatusconfig

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GatusAnnouncementConfig struct {
	Timestamp metav1.Time      `json:"timestamp"`
	Type      AnnouncementType `json:"type,omitempty"`
	Message   string           `json:"message"`
	Archived  *bool            `json:"archived,omitempty"`
}

type AnnouncementType string

const (
	AnnouncementTypeOutage      AnnouncementType = "outage"
	AnnouncementTypeWarning     AnnouncementType = "warning"
	AnnouncementTypeInformation AnnouncementType = "information"
	AnnouncementTypeOperational AnnouncementType = "operational"
	AnnouncementTypeNone        AnnouncementType = "none"
)
