/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"fmt"
	"os"
	"testing"

	pb "github.com/open-edge-platform/orch-metadata-broker/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

const jsonmetadataV0 = `{"keys":[{"name":"foo","values":["bar","baz"]}]}`

var expectedMetadataV0 = &Metadata{Keys: []Key{{
	Name:   "foo",
	Values: []string{"bar", "baz"},
}}}

const jsonmetadataV1 = `{"version":"v1","keys":[{"name":"foo","values":["bar","rab"]}]}`

var expectedMetadataV1 = &MetadataStoreV1{
	VersionedStore{Version: "v1"},
	Metadata{[]Key{{
		Name:   "foo",
		Values: []string{"bar", "rab"},
	}}},
}

const persistFolder = "/tmp"
const persistFile = "/tmp/meta.json"

var projectId = "123456"

func TestLoadMetadataV0(t *testing.T) {
	type args struct {
		setupContent []byte
		createFile   bool
	}
	tests := []struct {
		name    string
		args    args
		want    *Metadata
		wantErr assert.ErrorAssertionFunc
	}{
		{"missing-file", args{setupContent: []byte{}, createFile: false}, nil, assert.Error},
		{"empty-file", args{setupContent: []byte{}, createFile: true}, nil, assert.Error},
		{"no-content", args{setupContent: []byte("{}"), createFile: true}, &Metadata{}, assert.NoError},
		{"with-content", args{setupContent: []byte(jsonmetadataV0), createFile: true}, expectedMetadataV0, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Remove(persistFile)
			if tt.args.createFile {
				err := os.WriteFile(persistFile, tt.args.setupContent, 0644)
				assert.NoError(t, err)
			}

			got, err := LoadMetadataV0(persistFile)
			if !tt.wantErr(t, err, fmt.Sprintf("LoadMetadataV0(%v)", persistFile)) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetadata_createOrUpdate(t *testing.T) {

	type args struct {
		k []pb.Metadata
	}
	tests := []struct {
		name   string
		fields Metadata
		args   args
		wants  []Key
	}{
		{"add-one", Metadata{}, args{k: []pb.Metadata{{Key: "foo", Value: "bar"}}}, []Key{{Name: "foo", Values: []string{"bar"}}}},
		{"add-two", Metadata{}, args{k: []pb.Metadata{{Key: "foo", Value: "bar"}, {Key: "one", Value: "two"}}}, []Key{{Name: "foo", Values: []string{"bar"}}, {Name: "one", Values: []string{"two"}}}},
		{"add-lowercase", Metadata{}, args{k: []pb.Metadata{{Key: "Foo", Value: "Bar"}}}, []Key{{Name: "foo", Values: []string{"bar"}}}},
		{"add-idempotent", Metadata{}, args{k: []pb.Metadata{{Key: "Foo", Value: "Bar"}, {Key: "foo", Value: "bar"}}}, []Key{{Name: "foo", Values: []string{"bar"}}}},
		{"update", Metadata{Keys: []Key{{Name: "foo", Values: []string{"bar"}}}}, args{k: []pb.Metadata{{Key: "Foo", Value: "Bar"}, {Key: "foo", Value: "bar"}}}, []Key{{Name: "foo", Values: []string{"bar"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &tt.fields

			for _, pb := range tt.args.k { //nolint:all,copylocks
				m.createOrUpdate(&pb)
			}

			assert.Equal(t, tt.wants, m.Keys)
		})
	}
}

func TestMetadata_delete(t *testing.T) {

	type args struct {
		k []pb.Metadata
	}
	tests := []struct {
		name     string
		fields   Metadata
		args     args
		wants    []Key
		wantsErr assert.ErrorAssertionFunc
	}{
		{"missing-key", Metadata{Keys: []Key{{Name: "foo", Values: []string{}}}}, args{k: []pb.Metadata{{Key: "missing"}}}, []Key{{Name: "foo", Values: []string{}}}, assert.Error},
		{
			"remove-one",
			Metadata{[]Key{{Name: "foo", Values: []string{"bar", "rab"}}}},
			args{k: []pb.Metadata{{Key: "foo", Value: "bar"}}},
			[]Key{{"foo", []string{"rab"}}},
			assert.NoError,
		},
		{
			"remove-last",
			Metadata{[]Key{{Name: "foo", Values: []string{"bar"}}}},
			args{k: []pb.Metadata{{Key: "foo", Value: "bar"}}},
			[]Key{{"foo", []string{}}}, // FIXME NEX-1652 -> the entire key should be removed
			assert.NoError,
		},
		{
			"remove-lowercase",
			Metadata{[]Key{{Name: "foo", Values: []string{"bar"}}}},
			args{k: []pb.Metadata{{Key: "Foo", Value: "Bar"}}},
			[]Key{{"foo", []string{}}},
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &tt.fields

			for _, pb := range tt.args.k { //nolint:all,copylocks
				err := m.delete(&pb)
				if !tt.wantsErr(t, err, fmt.Sprintf("LoadMetadataV0(%v)", persistFile)) {
					return
				}

			}
			assert.Equal(t, tt.wants, m.Keys)
		})
	}
}

func TestSaveMetadataV1(t *testing.T) {
	type args struct {
		data *MetadataStoreV1
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"save",
			args{
				data: &MetadataStoreV1{VersionedStore{Version: "v1"}, Metadata{Keys: []Key{{Name: "foo", Values: []string{"bar", "rab"}}}}},
			},
			`{"version":"v1","keys":[{"name":"foo","values":["bar","rab"]}]}`,
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr(t, SaveMetadataV1(tt.args.data, persistFolder, projectId)) {
				return
			}

			got, err := loadFile(getFilename(persistFolder, projectId))
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestLoadMetadataV1(t *testing.T) {
	type args struct {
		setupContent []byte
		createFile   bool
	}
	tests := []struct {
		name    string
		args    args
		want    *MetadataStoreV1
		wantErr assert.ErrorAssertionFunc
	}{
		{"missing-file", args{setupContent: []byte{}, createFile: false}, &MetadataStoreV1{}, assert.NoError},
		{"empty-file", args{setupContent: []byte{}, createFile: true}, &MetadataStoreV1{}, assert.NoError},
		{"no-content", args{setupContent: []byte("{}"), createFile: true}, &MetadataStoreV1{}, assert.NoError},
		{"with-content", args{setupContent: []byte(jsonmetadataV1), createFile: true}, expectedMetadataV1, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := getFilename(persistFolder, projectId)
			_ = os.Remove(filename)
			if tt.args.createFile {
				err := os.WriteFile(filename, tt.args.setupContent, 0644)
				assert.NoError(t, err)
			}

			got, err := LoadMetadataV1(persistFolder, projectId)
			if !tt.wantErr(t, err, fmt.Sprintf("LoadMetadataV0(%v)", persistFile)) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_loadFile(t *testing.T) {
	type args struct {
		fileName             string
		createFileBeforeTest bool
		fileContent          string
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"file-exist",
			args{"/tmp/test.json", true, "content"},
			[]byte("content"),
			assert.NoError,
		},
		{
			"missing-file",
			args{"/tmp/test.json", false, ""},
			[]byte(""),
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.createFileBeforeTest {
				fmt.Println("writing to file")
				err := os.WriteFile(tt.args.fileName, []byte(tt.args.fileContent), 0644)
				assert.NoError(t, err)
			}

			got, err := loadFile(tt.args.fileName)
			if !tt.wantErr(t, err, fmt.Sprintf("loadFile(%v)", tt.args.fileName)) {
				return
			}
			assert.Equalf(t, tt.want, got, "loadFile(%v)", tt.args.fileName)

			t.Cleanup(func() {
				_ = os.Remove(tt.args.fileName)
			})
		})
	}
}

func TestDeleteProject(t *testing.T) {
	type args struct {
		projectId            string
		createFileBeforeTest bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"file-exists",
			args{"test-proj-1", true},
			false,
		},
		{
			"file-does-not-exist",
			args{"test-proj-1", false},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := getFilename(persistFolder, tt.args.projectId)

			if tt.args.createFileBeforeTest {
				fmt.Println("creating a file")

				_, err := os.Create(filename)
				assert.NoError(t, err)
			}

			if err := DeleteProject(persistFolder, tt.args.projectId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteProject() error = %v, wantErr %v", err, tt.wantErr)
			}

			fmt.Println("checking if file has been deleted")
			_, err := os.Stat(filename)
			assert.Error(t, err)
		})
	}
}
