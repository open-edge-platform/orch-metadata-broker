/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"context"
	"fmt"

	"github.com/open-edge-platform/orch-library/go/pkg/tenancy"
	"github.com/open-edge-platform/orch-metadata-broker/internal/impl"
)

// tenancyHandler implements tenancy.Handler for the metadata-broker.
// It handles project lifecycle events: create is a no-op, delete removes
// the project's metadata file.
type tenancyHandler struct{}

// HandleEvent processes a single tenancy event. Only project events are
// handled; org events are ignored.
func (h *tenancyHandler) HandleEvent(_ context.Context, event tenancy.Event) error {
	if event.ResourceType != "project" {
		return nil // only project events are relevant
	}

	switch event.EventType {
	case "created":
		log.Infof("Runtime Project created (no action): %s", event.ResourceName)
		return nil

	case "deleted":
		return h.handleProjectDeleted(event)

	default:
		log.Infof("Ignoring unknown event type %q for project %s", event.EventType, event.ResourceName)
		return nil
	}
}

// handleProjectDeleted removes the metadata file for the deleted project.
// Uses the resource UUID (not name) since metadata files are keyed by the
// ActiveProjectID header, which contains the UUID.
func (h *tenancyHandler) handleProjectDeleted(event tenancy.Event) error {
	projectID := event.ResourceID.String()
	log.Infof("Project %s (%s) marked for deletion in metadata broker", event.ResourceName, projectID)

	if err := impl.DeleteProject(&projectID); err != nil {
		return fmt.Errorf("failed to delete project %s metadata: %w", projectID, err)
	}

	log.Infof("Deleted metadata for project %s (%s)", event.ResourceName, projectID)
	return nil
}
