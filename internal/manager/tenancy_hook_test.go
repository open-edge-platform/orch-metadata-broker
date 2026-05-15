// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/open-edge-platform/orch-library/go/pkg/tenancy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/orch-metadata-broker/internal/impl"
)

func newEvent(resourceType, eventType string, id uuid.UUID, name string) tenancy.Event {
	return tenancy.Event{
		ID:           1,
		ResourceType: resourceType,
		EventType:    eventType,
		ResourceID:   id,
		ResourceName: name,
	}
}

// TestMetadataHandler_HandleEvent_ProjectDeleted verifies that a project-deleted event
// causes impl.DeleteProject to be called for the matching project.
func TestMetadataHandler_HandleEvent_ProjectDeleted(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, impl.Init("", dir))

	projectID := uuid.New()

	// Pre-create a metadata file so DeleteProject has something to remove.
	filename := filepath.Join(dir, fmt.Sprintf("metadata-%s.json", projectID))
	require.NoError(t, os.WriteFile(filename, []byte(`{"version":"v1","keys":[]}`), 0o600))

	h := &metadataHandler{}
	err := h.HandleEvent(context.Background(), newEvent(
		tenancy.ResourceTypeProject,
		tenancy.EventTypeDeleted,
		projectID,
		"test-project",
	))
	assert.NoError(t, err)

	// File should be gone.
	_, statErr := os.Stat(filename)
	assert.True(t, errors.Is(statErr, os.ErrNotExist), "metadata file should have been deleted")
}

// TestMetadataHandler_HandleEvent_ProjectCreated verifies that a project-created event
// is a no-op (metadata is created on first use, not eagerly).
func TestMetadataHandler_HandleEvent_ProjectCreated(t *testing.T) {
	h := &metadataHandler{}
	err := h.HandleEvent(context.Background(), newEvent(
		tenancy.ResourceTypeProject,
		tenancy.EventTypeCreated,
		uuid.New(),
		"new-project",
	))
	assert.NoError(t, err)
}

// TestMetadataHandler_HandleEvent_OrgEvents verifies that org events are ignored.
func TestMetadataHandler_HandleEvent_OrgEvents(t *testing.T) {
	h := &metadataHandler{}
	for _, eventType := range []string{tenancy.EventTypeCreated, tenancy.EventTypeDeleted} {
		err := h.HandleEvent(context.Background(), newEvent(
			tenancy.ResourceTypeOrg,
			eventType,
			uuid.New(),
			"some-org",
		))
		assert.NoError(t, err, "org %s event should be a no-op", eventType)
	}
}

// TestMetadataHandler_HandleEvent_DeleteProjectMissingDir verifies that deleting a
// project whose data directory does not exist is still treated as success (idempotent).
func TestMetadataHandler_HandleEvent_DeleteProjectMissingDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, impl.Init("", dir))

	h := &metadataHandler{}
	// No file created — project never had data.
	err := h.HandleEvent(context.Background(), newEvent(
		tenancy.ResourceTypeProject,
		tenancy.EventTypeDeleted,
		uuid.New(),
		"ghost-project",
	))
	assert.NoError(t, err)
}

// TestTenancyHook_SubscribeUnsubscribe verifies that Subscribe/Unsubscribe can be called
// without crashing when the Tenant Manager is unreachable.
func TestTenancyHook_SubscribeUnsubscribe(t *testing.T) {
	// Point at a local address that is not listening so the poller will fail
	// gracefully after cancellation.
	t.Setenv("TENANT_MANAGER_URL", "http://127.0.0.1:19999")

	h := NewTenancyHook()
	err := h.Subscribe()
	// NewPoller itself should succeed (network errors occur in the background goroutine).
	assert.NoError(t, err)

	// Cancelling should not panic.
	h.Unsubscribe()
}

// TestTenancyHook_UnsubscribeWithoutSubscribe verifies that calling Unsubscribe before
// Subscribe does not panic.
func TestTenancyHook_UnsubscribeWithoutSubscribe(t *testing.T) {
	h := NewTenancyHook()
	assert.NotPanics(t, func() { h.Unsubscribe() })
}

// TestTenancyHook_DoubleSubscribeReturnsError verifies that a second Subscribe call
// before Unsubscribe returns an error instead of leaking a goroutine.
func TestTenancyHook_DoubleSubscribeReturnsError(t *testing.T) {
	t.Setenv("TENANT_MANAGER_URL", "http://127.0.0.1:19999")

	h := NewTenancyHook()
	require.NoError(t, h.Subscribe())
	defer h.Unsubscribe()

	err := h.Subscribe()
	assert.ErrorContains(t, err, "already subscribed")
}
