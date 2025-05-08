/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
*/

package models

import (
	"os"
)

type migration func(*string, *string, *string) error

func migrateV0(persistData, persistFolder, defaultProjectId *string) error {
	log.Info("Migrating data from v0 to v1")
	metadata, err := LoadMetadataV0(*persistData)
	if err != nil {
		return err
	}

	log.Debugf("Existing data: %+v", metadata)

	newData := &MetadataStoreV1{
		VersionedStore{Version: "v1"},
		*metadata,
	}
	log.Infof("New Data: %+v", newData)

	err = SaveMetadataV1(newData, *persistFolder, *defaultProjectId)
	if err != nil {
		return err
	}

	log.Infof("Removing old persistent file %s", *persistData)
	err = os.Remove(*persistData)
	if err != nil {
		return err
	}
	return nil
}

var migrations = map[string]migration{
	"v0": migrateV0,
}

// Migrate receives the content of the backup file.
// If the content does not match the latest format it applies the required migration(s).
func Migrate(persistData, persistFolder, defaultProjectId string) error {
	log.Infof("Migrating (persistData: %s, persistFolder: %s, defaultProjectId: %s)", persistData, persistFolder, defaultProjectId)

	if _, e := os.Stat(persistData); e == nil {
		// else read the data, convert them in the new format and write them back into a file
		// note that the file is suffixed with the defaultProjectId
		return migrations["v0"](&persistData, &persistFolder, &defaultProjectId)
	} else {
		// if there is no persistData file we're starting fresh, nothing to do for now
		log.Info("There are no persisted data in the 24.08 format, nothing to do", persistData)

		// in the future iterate over all files and check that the version is as expected
		// for file in range ...
		// versioned := VersionedStore{}
		//	_ = json.Unmarshal(data, &versioned)
		// if versioned.Version == "v1" { ...
	}

	// no migration required
	return nil
}
