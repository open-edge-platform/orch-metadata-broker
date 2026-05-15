// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/open-edge-platform/orch-library/go/pkg/tenancy"
	"github.com/open-edge-platform/orch-metadata-broker/internal/impl"
)

const (
	appName                 = "metadata-broker"
	tenantManagerURLEnvVar  = "TENANT_MANAGER_URL"
	defaultTenantManagerURL = "http://tenancy-manager.orch-iam:8080"
)

// TenancyHook replaces the former Nexus-based project lifecycle listener.
// It consumes project events from the Tenant Manager REST API via the shared
// orch-library tenancy poller.
type TenancyHook struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

// NewTenancyHook creates a TenancyHook.
func NewTenancyHook() *TenancyHook {
	return &TenancyHook{}
}

// Subscribe starts the tenancy poller in a background goroutine.
// If already subscribed, it returns an error.
func (h *TenancyHook) Subscribe() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cancel != nil {
		return fmt.Errorf("tenancy hook already subscribed")
	}

	tenantManagerURL := os.Getenv(tenantManagerURLEnvVar)
	if tenantManagerURL == "" {
		tenantManagerURL = defaultTenantManagerURL
	}

	handler := &metadataHandler{}
	poller, err := tenancy.NewPoller(tenantManagerURL, appName, handler,
		func(cfg *tenancy.PollerConfig) {
			cfg.OnError = func(err error, msg string) {
				log.Warnf("tenancy poller: %s: %v", msg, err)
			}
		},
	)
	if err != nil {
		return fmt.Errorf("create tenancy poller: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	go func() {
		if err := poller.Run(ctx); err != nil && ctx.Err() == nil {
			log.Errorf("tenancy poller stopped unexpectedly: %v", err)
		}
	}()

	log.Infof("Tenancy hook subscribed: controller=%s url=%s", appName, tenantManagerURL)
	return nil
}

// Unsubscribe cancels the background poller.
func (h *TenancyHook) Unsubscribe() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cancel != nil {
		h.cancel()
		h.cancel = nil
	}
}

// metadataHandler implements tenancy.Handler for the metadata-broker.
// On project deletion it removes all metadata for the project.
// On project creation it does nothing — metadata storage is created on first use.
type metadataHandler struct{}

func (h *metadataHandler) HandleEvent(_ context.Context, event tenancy.Event) error {
	switch {
	case event.ResourceType == tenancy.ResourceTypeProject && event.EventType == tenancy.EventTypeDeleted:
		return h.handleProjectDeleted(event)
	default:
		// Org events and project-created events require no action.
		return nil
	}
}

func (h *metadataHandler) handleProjectDeleted(event tenancy.Event) error {
	projectID := event.ResourceID.String()
	log.Infof("Deleting metadata for project %s (%s)", event.ResourceName, projectID)

	if err := impl.DeleteProject(&projectID); err != nil {
		return fmt.Errorf("delete project %s metadata: %w", event.ResourceName, err)
	}

	log.Infof("Metadata deleted for project %s", event.ResourceName)
	return nil
}
