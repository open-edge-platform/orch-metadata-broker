package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/labstack/gommon/log"

	"github.com/open-edge-platform/orch-metadata-broker/internal/manager"
)

func main() {
	caPath := flag.String("caPath", "", "path to CA certificate")
	keyPath := flag.String("keyPath", "", "path to client private key")
	certPath := flag.String("certPath", "", "path to client certificate")
	backupFile := flag.String("backupFile", "/data/metadata.json", "file that metadata is persisted to and loaded from at startup")
	backupFolder := flag.String("backupFolder", "/data", "Folder used to store backup files")
	specFilePath := flag.String("openapiSpec", "/opt/openapi.yaml", "The location of the OpenAPI spec file")
	allowedCorsOrigins := flag.String("allowedCorsOrigins", "", "Comma separated list of allowed CORS origins")
	basePath := flag.String("basePath", "", "The rest server basePath (REST API prefix)")
	restPort := flag.Int("restPort", 9988, "port that REST service runs on")
	grpcPort := flag.Int("grpcPort", 9987, "The endpoint of the gRPC server")
	opaPort := flag.Int("opaPort", 9986, "The endpoint of the Open Policy Agent")
	flag.Parse()

	// create a channel to manage the servers lifecycle
	doneChannel := make(chan bool)

	// create a Waitgroup to wait for the server to exit before shutting down
	wg := sync.WaitGroup{}

	// listen on SIGTERM to signal to the servers to shutdown (via doneChannel)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		<-sigs
		close(doneChannel)
	}()

	cfg := manager.Config{
		CAPath:             *caPath,
		KeyPath:            *keyPath,
		CertPath:           *certPath,
		GRPCPort:           *grpcPort,
		RestPort:           *restPort,
		OPAPort:            *opaPort,
		BasePath:           *basePath,
		AllowedCorsOrigins: *allowedCorsOrigins,
		BackupFile:         *backupFile,
		OpenapiSpecFile:    *specFilePath,
		BackupFolder:       *backupFolder,
	}

	log.Infof("Metadata Broker starting with config: %+v", cfg)

	mgr := manager.NewManager(doneChannel, &wg, cfg)
	mgr.Run()
	wg.Wait()
}
