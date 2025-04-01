package manager

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/atomix/dazl"
	"github.com/open-edge-platform/orch-library/go/pkg/northbound"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"github.com/open-edge-platform/orch-metadata-broker/internal/grpc"
	"github.com/open-edge-platform/orch-metadata-broker/internal/impl"
	"github.com/open-edge-platform/orch-metadata-broker/internal/rest"
)

var log = dazl.GetPackageLogger()

// Config is a manager configuration
type Config struct {
	CAPath             string
	KeyPath            string
	CertPath           string
	GRPCPort           int
	RestPort           int
	OPAPort            int
	BasePath           string
	AllowedCorsOrigins string
	BackupFile         string
	BackupFolder       string
	OpenapiSpecFile    string
}

// Manager single point of entry for the provisioner
type Manager struct {
	Config Config
	doneCh chan bool
	wg     *sync.WaitGroup
}

// NewManager initializes the application manager
func NewManager(doneCh chan bool, wg *sync.WaitGroup, cfg Config) *Manager {
	return &Manager{
		Config: cfg,
		doneCh: doneCh,
		wg:     wg,
	}
}

func (m *Manager) Run() {
	log.Info("Starting manager")
	if err := m.Start(); err != nil {
		log.Fatalw("Unable to start manager", dazl.Error(err))
	}
}

func (m *Manager) Start() error {
	err := impl.Init(m.Config.BackupFile, m.Config.BackupFolder)
	if err != nil {
		log.Info("unable to initialize data store from backup %v assuming new installation\n", err)
	}

	m.wg.Add(1)
	go func() {
		if err = m.startNorthboundServer(); err != nil {
			log.Fatalf("cannot-start-grpc-server", err.Error())
		}
	}()

	m.wg.Add(1)
	go func() {
		if err = m.startRestServer(); err != nil {
			log.Fatalf("cannot-start-rest-server", err.Error())
		}
	}()

	log.Info("Subscribing to Nexus")

	nexusHook := NewNexusHook()
	err = nexusHook.Subscribe()

	if err != nil {
		log.Errorf("Unable to subscribe to Nexus hook %v", err)
	}

	m.wg.Wait()
	return nil
}

const OIDCServerURL = "OIDC_SERVER_URL"

// startNorthboundServer starts the northbound gRPC server
func (m *Manager) startNorthboundServer() error {
	serverConfig := northbound.NewInsecureServerConfig(int16(m.Config.GRPCPort))

	if oidcURL := os.Getenv(OIDCServerURL); oidcURL != "" {
		serverConfig.SecurityCfg = &northbound.SecurityConfig{
			AuthenticationEnabled: true,
			AuthorizationEnabled:  true,
		}
		log.Infof("Authentication enabled. %s=%s", OIDCServerURL, oidcURL)
		// OIDCServerURL is also referenced in jwt.go (from github.com/open-edge-platform/orch-library/go)
	} else {
		log.Infof("Authentication not enabled %s not set", OIDCServerURL)
	}

	s := northbound.NewServer(serverConfig)

	serverAddr := fmt.Sprintf("http://localhost:%d", m.Config.OPAPort)

	var opaClient openpolicyagent.ClientWithResponsesInterface
	var err error
	if serverConfig.SecurityCfg.AuthorizationEnabled {
		opaClient, err = openpolicyagent.NewClientWithResponses(serverAddr)
		if err != nil {
			log.Fatalf("OPA server cannot be created %v", err)
		}
	}

	s.AddService(grpc.NewService(opaClient))

	doneCh := make(chan error)
	go func() {
		err := s.Serve(func(started string) {
			log.Info("Started NBI on ", started)
			close(doneCh)
		})
		if err != nil {
			doneCh <- err
		}
	}()
	m.wg.Done()
	return <-doneCh
}

func (m *Manager) startRestServer() error {
	s := rest.NewServer(m.Config.RestPort, m.Config.GRPCPort, m.Config.BasePath, m.Config.AllowedCorsOrigins, m.Config.OpenapiSpecFile)
	// start server
	log.Infow("Starting REST proxy Server", dazl.Int("address", m.Config.RestPort))
	err := s.ListenAndServe()
	if err != nil {
		return err
	}

	x := <-m.doneCh
	if x {
		// if the API channel is closed, stop the REST server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			log.Error("Cannot stop Rest server")
		}
		log.Info("Rest server stopped")
	}
	m.wg.Done()
	return nil
}
