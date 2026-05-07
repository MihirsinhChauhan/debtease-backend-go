package models

import "github.com/google/uuid"

// SnapshotAttachResponse is returned by POST /api/v1/snapshot/attach.
type SnapshotAttachResponse struct {
	Attached   bool       `json:"attached"`
	SnapshotID *uuid.UUID `json:"snapshot_id,omitempty"`
}
