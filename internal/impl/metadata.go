package impl

import (
	"os"

	"github.com/atomix/dazl"
	"github.com/open-edge-platform/orch-metadata-broker/internal/models"
	pb "github.com/open-edge-platform/orch-metadata-broker/pkg/api/v1"
)

var log = dazl.GetPackageLogger()

// TODO consider create a struct to hold these data
var _dataFolder string

// Init called at startup to load in persisted metadata
func Init(persistData string, persistFolder string) error {

	_dataFolder = persistFolder

	// "Migration project ID, required to migrate data at startup when upgrading from 24.08"
	migrationProjectId, doesMigrationProjectIdExist := os.LookupEnv("MIGRATION_PROJECT_ID")

	if doesMigrationProjectIdExist {
		log.Infof("MIGRATION_PROJECT_ID exists: %s", migrationProjectId)

		// Checks for older data and migrate to the latest format
		err := models.Migrate(persistData, persistFolder, migrationProjectId)
		if err != nil {
			return err
		}
	} else {
		log.Info("MIGRATION_PROJECT_ID does not exist. Migration skipped.")
	}

	return nil
}

func GetSystemMetadata(projectId *string) ([]*pb.StoredMetadata, error) {
	log.Infof("GetSystemMetadata (projectID: %v)", projectId)
	metadata, err := models.LoadMetadataV1(_dataFolder, *projectId)
	log.Debugf("Got Metadata: %+v", metadata)
	if err != nil {
		return nil, err
	}
	return metadata.GetKeyValues()
}

func CreateOrUpdate(projectId *string, k *pb.Metadata) ([]*pb.StoredMetadata, error) {
	log.Infof("CreateOrUpdate (projectID: %v): %+v", projectId, k)
	metadata, err := models.LoadMetadataV1(_dataFolder, *projectId)
	if err != nil {
		return nil, err
	}
	if err := metadata.CreateOrUpdate(k); err != nil {
		return nil, err
	}
	pbMeta, err := metadata.GetKeyValues()
	if err != nil {
		return nil, err
	}
	return pbMeta, models.SaveMetadataV1(metadata, _dataFolder, *projectId)
}

func Delete(projectId *string, k *pb.Metadata) ([]*pb.StoredMetadata, error) {
	log.Infof("Delete (projectID: %s): %+v", projectId, k)
	metadata, err := models.LoadMetadataV1(_dataFolder, *projectId)
	if err != nil {
		return nil, err
	}
	if err := metadata.Delete(k); err != nil {
		return nil, err
	}
	pbMeta, err := metadata.GetKeyValues()
	if err != nil {
		return nil, err
	}
	return pbMeta, models.SaveMetadataV1(metadata, _dataFolder, *projectId)
}

func DeleteProject(projectId *string) error {
	log.Infof("Delete (projectID: %s)", projectId)

	err := models.DeleteProject(_dataFolder, *projectId)

	if err != nil {
		return err
	}

	return nil
}
