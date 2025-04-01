package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/open-edge-platform/orch-metadata-broker/internal/impl"
	projectActiveWatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectactivewatcher.edge-orchestrator.intel.com/v1"
	projectwatcherv1 "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/apis/projectwatcher.edge-orchestrator.intel.com/v1"
	nexus "github.com/open-edge-platform/orch-utils/tenancy-datamodel/build/nexus-client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	appName = "metadata-broker"

	// Allow only certain time for interacting with Nexus server
	nexusTimeout = 5 * time.Second
)

type NexusHook struct {
	nexusClient *nexus.Clientset
}

// NewNexusHook creates a new hook for receiving project lifecycle events from Nexus.
func NewNexusHook() *NexusHook {
	return &NexusHook{}
}

// Subscribe issues all required subscriptions for receiving project lifecycle events.
func (h *NexusHook) Subscribe() error {
	// Initialize Nexus SDK, by pointing it to the K8s API endpoint where CRD's are to be stored.
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf("Unable to load in-cluster configuration: %+v", err)
		return err
	}

	h.nexusClient, err = nexus.NewForConfig(cfg)
	if err != nil {
		log.Errorf("Unable to create Nexus configuration: %+v", err)
		return err
	}

	// Register the configuration provisioner watcher node in the configuration subtree.
	if err := h.setupConfigProvisionerWatcherConfig(); err != nil {
		return err
	}

	// Subscribe to Multi-Tenancy graph.
	// Subscribe() api empowers subscription to objects from datamodel.
	// What subscription does is to keep the local cache in sync with datamodel changes.
	// This sync is done in the background.
	h.nexusClient.SubscribeAll()

	if _, err := h.nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").Folders("*").Projects("*").RegisterAddCallback(h.projectCreated); err != nil {
		log.Errorf("Unable to register project creation callback: %+v", err)
		return err
	}

	if _, err := h.nexusClient.TenancyMultiTenancy().Runtime().Orgs("*").Folders("*").Projects("*").RegisterUpdateCallback(h.projectUpdated); err != nil {
		log.Errorf("Unable to register project deletion callback: %+v", err)
		return err
	}

	log.Info("Nexus hook successfully subscribed")

	return nil
}

func (h *NexusHook) setupConfigProvisionerWatcherConfig() error {
	tenancy := h.nexusClient.TenancyMultiTenancy()

	ctx, cancel := context.WithTimeout(context.Background(), nexusTimeout)
	defer cancel()

	projWatcher, err := tenancy.Config().AddProjectWatchers(ctx, &projectwatcherv1.ProjectWatcher{ObjectMeta: metav1.ObjectMeta{
		Name: appName,
	}})

	if nexus.IsAlreadyExists(err) {
		log.Warnf("Project watcher already exist: appName=%s, projWatcher=%v", appName, projWatcher)
	} else if err != nil {
		log.Errorf("Failed to create project watcher: appName=%s", appName)
		return err
	}
	log.Infof("Created project watcher: appName=%s, projWatcher=%v", appName, projWatcher)
	return nil
}

// Callback function to be invoked when Project is added.
func (h *NexusHook) projectCreated(project *nexus.RuntimeprojectRuntimeProject) {
	log.Infof("Runtime Project: %+v created", *project)

	ctx, cancel := context.WithTimeout(context.Background(), nexusTimeout)
	defer cancel()

	// Register this app as an active watcher for this project.
	watcherObj, err := project.AddActiveWatchers(ctx, &projectActiveWatcherv1.ProjectActiveWatcher{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: projectActiveWatcherv1.ProjectActiveWatcherSpec{
			StatusIndicator: projectActiveWatcherv1.StatusIndicationIdle,
			Message:         fmt.Sprintf("Added active project watcher for %v", project.DisplayName()),
			TimeStamp:       uint64(time.Now().Unix()),
		},
	})

	if nexus.IsAlreadyExists(err) {
		log.Warnf("Watch %s already exists for project %s", watcherObj.DisplayName(), project.DisplayName())
	} else if err != nil {
		log.Errorf("Error %+v while creating watch %s for project %s", err, appName, project.DisplayName())
	}
	log.Infof("Active watcher %s created for Project %s", watcherObj.DisplayName(), project.DisplayName())
}

// Callback function to be invoked when Project is deleted.
func (h *NexusHook) projectUpdated(_, project *nexus.RuntimeprojectRuntimeProject) {
	projectName := project.DisplayName()
	if project.Spec.Deleted {
		log.Infof("Project: %+v marked for deletion in metadata broker", projectName)

		err := impl.DeleteProject(&projectName)

		if err != nil {
			watcher, err := project.GetActiveWatchers(context.Background(), appName)

			if err != nil {
				log.Errorf("Error while getting the active watcher: %v", err)
			}

			watcher.Spec.StatusIndicator = projectActiveWatcherv1.StatusIndicationError
			watcher.Spec.TimeStamp = uint64(time.Now().Unix())

			err = watcher.Update(context.Background())

			if err != nil {
				log.Errorf("Error while updating Spec for the project: %v", err)
			}

			log.Errorf("Error while deleting project inside metadata broker: %v", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), nexusTimeout)
		defer cancel()

		// Stop watching the project as it is marked for deletion.
		err = project.DeleteActiveWatchers(ctx, appName)
		if nexus.IsChildNotFound(err) {
			// This app has already stopped watching the project.
			log.Warnf("App %s DOES NOT watch project %s", appName, projectName)
			return
		} else if err != nil {
			log.Errorf("Error %+v while deleting watch %s for project %s", err, appName, projectName)
			return
		}
		log.Infof("Active watcher %s deleted for project %s", appName, projectName)
	}
}
