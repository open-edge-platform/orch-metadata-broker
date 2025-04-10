/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
*/

package models

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/atomix/dazl"
	pb "github.com/open-edge-platform/orch-metadata-broker/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var log = dazl.GetPackageLogger()
var lock sync.Mutex

type Key struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func (k *Key) AddValue(v string) {
	k.Values = append(k.Values, v)
}

type Metadata struct {
	Keys []Key `json:"keys"`
}

func (m *Metadata) GetJson() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Metadata) GetKeyValues() ([]*pb.StoredMetadata, error) {
	var keyValues []*pb.StoredMetadata
	for i := 0; i < len(m.Keys); i++ {
		k := m.Keys[i].Name
		vals := m.Keys[i].Values
		kv := pb.StoredMetadata{Key: k, Values: vals}
		log.Debugf("Found Metadata entry: {Key: %s, values: %s}", k, vals)
		keyValues = append(keyValues, &kv)
	}

	return keyValues, nil
}

func (m *Metadata) createOrUpdate(k *pb.Metadata) {

	// make sure that we only store lowercase metadata to avoid confusion
	md := &pb.Metadata{
		Key:   strings.ToLower(k.Key),
		Value: strings.ToLower(k.Value),
	}

	for i := 0; i < len(m.Keys); i++ {
		key := &m.Keys[i]
		if key.Name == md.Key {
			//Key exists so check if value already present
			for j := 0; j < len(key.Values); j++ {
				if md.Value == key.Values[j] {
					// Already exists so just exit quietly
					return
				}
			}
			//append value to slice
			key.AddValue(md.Value)
			log.Debugf("Adding Value %s\n", md.Value)
			return
		}
	}
	log.Debugf("Adding Key %s with Value %s", md.Key, md.Value)
	m.Keys = append(m.Keys, Key{
		Name:   md.Key,
		Values: []string{md.Value},
	})
}

func (m *Metadata) delete(k *pb.Metadata) error {

	md := &pb.Metadata{
		Key:   strings.ToLower(k.Key),
		Value: strings.ToLower(k.Value),
	}

	for i := 0; i < len(m.Keys); i++ {
		key := &m.Keys[i]
		if key.Name == md.Key {
			//Key exists so check if value exists
			for j := 0; j < len(key.Values); j++ {
				if md.Value == key.Values[j] {
					// Value exists, so remove it
					remaining := append(key.Values[:j], key.Values[j+1:]...)
					key.Values = remaining
					return nil
				}
			}
		}
	}
	return fmt.Errorf("not-found")
}

type VersionedStore struct {
	Version string `json:"version"`
}

func (s *MetadataStoreV1) GetJson() ([]byte, error) {
	return json.Marshal(s)
}

func (s *MetadataStoreV1) GetKeyValues() ([]*pb.StoredMetadata, error) {
	return s.Metadata.GetKeyValues()
}

func (s *MetadataStoreV1) CreateOrUpdate(k *pb.Metadata) error {
	lock.Lock()
	defer lock.Unlock()

	s.Metadata.createOrUpdate(k)

	return nil
}

func (s *MetadataStoreV1) Delete(k *pb.Metadata) error {
	lock.Lock()
	defer lock.Unlock()

	return s.Metadata.delete(k)
}

// MetadataStoreV1 Supports Project isolation.
type MetadataStoreV1 struct {
	VersionedStore
	Metadata
}

func loadFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(fileName)

			if err != nil {
				log.Errorf("Error while creating persistDataFile %s", fileName)
				return nil, err
			}
		} else {
			log.Errorf("Error while reading persistDataFile %s", fileName)
			return nil, err
		}
	}

	defer file.Close()
	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, err
}

// LoadMetadataV0 Loads data from a v0 file
func LoadMetadataV0(fileName string) (*Metadata, error) {
	bytes, err := loadFile(fileName)
	if err != nil {
		return nil, err
	}
	m := Metadata{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func getFilename(folder, projectId string) string {
	fileName := fmt.Sprintf("metadata-%s.json", projectId)
	return path.Join(folder, fileName)
}

// SaveMetadataV1 Saves data to a v1 file
func SaveMetadataV1(data *MetadataStoreV1, persistFolder, defaultProjectId string) error {
	bytes, err := data.GetJson()
	if err != nil {
		return err
	}

	filename := getFilename(persistFolder, defaultProjectId)
	log.Infof("Saving metadata to file (%s): %+v", filename, data)
	return os.WriteFile(filename, bytes, 0644)
}

// LoadMetadataV1 Loads data from a v1 file
func LoadMetadataV1(persistFolder, defaultProjectId string) (*MetadataStoreV1, error) {
	bytes, err := loadFile(getFilename(persistFolder, defaultProjectId))
	if err != nil {
		return nil, err
	}
	m := MetadataStoreV1{}
	_ = json.Unmarshal(bytes, &m)
	return &m, nil
}

func DeleteProject(persistFolder, projectId string) error {
	fileName := getFilename(persistFolder, projectId)

	err := os.Remove(fileName)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Debugf("Project %s doesn't exist", projectId)
			return nil
		}

		if errors.Is(err, os.ErrPermission) {
			log.Debugf("Permission denied while deleting project %s", projectId)
			return status.Error(codes.PermissionDenied, "permission denied at OS level")
		}

		log.Errorf("Failed to delete project: %v", err)
		return status.Error(codes.Unknown, err.Error())
	}

	log.Infof("Successfully deleted project %s", projectId)
	return nil
}
