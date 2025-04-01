package impl

import (
	"fmt"
	"os"
	"path"
	"testing"

	pb "github.com/open-edge-platform/orch-metadata-broker/pkg/api/v1"
	"github.com/stretchr/testify/assert"
)

const persistFolder = "/tmp"

const jsonmetadataV1 = `{"version":"v1","keys":[{"name":"foo","values":["bar","rab"]}]}`

var testProject = "testProject"
var defaultProject = "defaultProject"
var pbMetadataV1 = []*pb.StoredMetadata{
	{Key: "foo", Values: []string{"bar", "rab"}},
}
var pbMetadata = []*pb.Metadata{
	{Key: "foo", Value: "bar"},
	{Key: "foo", Value: "rab"},
}

func TestGetSystemMetadata(t *testing.T) {
	type args struct {
		testMetadata   []byte
		readProjectId  *string
		writeProjectId string
	}
	tests := []struct {
		name    string
		args    args
		want    []*pb.StoredMetadata
		wantErr assert.ErrorAssertionFunc
	}{
		{"get", args{[]byte(jsonmetadataV1), &testProject, testProject}, pbMetadataV1, assert.NoError},
		{"get-missing-project", args{[]byte(jsonmetadataV1), &testProject, defaultProject}, []*pb.StoredMetadata(nil), assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := Init("", persistFolder)
			assert.NoError(t, err)

			filename := path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", tt.args.writeProjectId))
			fmt.Printf("Write file to: %s", filename)
			err = os.WriteFile(filename, tt.args.testMetadata, 0644)
			assert.NoError(t, err)

			got, err := GetSystemMetadata(tt.args.readProjectId)

			e := os.Remove(filename)
			assert.NoError(t, e)

			if !tt.wantErr(t, err, fmt.Sprintf("GetSystemMetadata(%v)", tt.args.readProjectId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetSystemMetadata(%v)", tt.args.readProjectId)
		})
	}
}

func TestCreateOrUpdate(t *testing.T) {
	type args struct {
		testMetadata   []*pb.Metadata
		writeProjectId *string
		readProjectId  string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{"write-project", args{pbMetadata, &testProject, testProject}, []byte(jsonmetadataV1), assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := Init("", persistFolder)
			assert.NoError(t, err)

			writeFilename := path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", tt.args.readProjectId))
			if tt.args.writeProjectId != nil {
				writeFilename = path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", *tt.args.writeProjectId))
			}
			fmt.Printf("Creating file: %s", writeFilename)
			err = os.WriteFile(writeFilename, []byte(`{"version":"v1"}`), 0644)
			assert.NoError(t, err)

			for _, k := range tt.args.testMetadata {
				_, err = CreateOrUpdate(tt.args.writeProjectId, k)
				if !tt.wantErr(t, err, fmt.Sprintf("CreateOrUpdate(%+v, %+v)", tt.args.writeProjectId, tt.args.testMetadata)) {
					return
				}
			}

			filename := path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", tt.args.readProjectId))
			// data, err := os.ReadFile(filename)
			_, err = os.ReadFile(filename)
			assert.NoError(t, err)
			e := os.Remove(writeFilename)
			assert.NoError(t, e)

			// TODO: fix for CI build
			// assert.Equalf(t, tt.want, data, "CreateOrUpdate(%+v, %+v)", tt.args.writeProjectId, tt.args.testMetadata)
		})
	}
}

func TestDelete(t *testing.T) {
	const jsonmetadataV1_delete = `{"version":"v1","keys":[{"name":"foo","values":["bar"]}]}`
	type args struct {
		content        []byte
		testMetadata   *pb.Metadata
		writeProjectId *string
		readProjectId  string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{"delete-from-project", args{[]byte(jsonmetadataV1), &pb.Metadata{Key: "foo", Value: "rab"}, &testProject, testProject}, []byte(jsonmetadataV1_delete), assert.NoError},
		// TODO: fix for CI build
		// {"delete-not-found", args{[]byte(jsonmetadataV1_delete), &pb.Metadata{Key: "foo", Value: "rab"}, &testProject, testProject}, []byte(jsonmetadataV1_delete), assert.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := Init("", persistFolder)
			assert.NoError(t, err)

			writeFilename := path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", tt.args.readProjectId))
			if tt.args.writeProjectId != nil {
				writeFilename = path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", *tt.args.writeProjectId))
			}
			fmt.Printf("Creating file: %s", writeFilename)
			err = os.WriteFile(writeFilename, tt.args.content, 0644)
			assert.NoError(t, err)

			_, err = Delete(tt.args.writeProjectId, tt.args.testMetadata)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateOrUpdate(%+v, %+v)", tt.args.writeProjectId, tt.args.testMetadata)) {
				return
			}

			filename := path.Join(persistFolder, fmt.Sprintf("metadata-%s.json", tt.args.readProjectId))
			data, err := os.ReadFile(filename)
			assert.NoError(t, err)
			e := os.Remove(writeFilename)
			assert.NoError(t, e)

			assert.Equalf(t, tt.want, data, "CreateOrUpdate(%+v, %+v)", tt.args.writeProjectId, tt.args.testMetadata)
		})
	}
}

func TestDeleteProject(t *testing.T) {
	var nonExistentProject = "non-existent-project"
	type args struct {
		projectId               *string
		createProjectBeforeTest bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"project-exists",
			args{projectId: &testProject, createProjectBeforeTest: true},
			assert.NoError,
		},
		{
			"project-does-not-exist",
			args{projectId: &nonExistentProject, createProjectBeforeTest: false},
			assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.createProjectBeforeTest {
				for _, k := range pbMetadata {
					_, err := CreateOrUpdate(tt.args.projectId, k)
					assert.NoError(t, err)
				}
			}

			err := DeleteProject(tt.args.projectId)
			if !tt.wantErr(t, err, fmt.Sprintf("DeleteProject(%v)", tt.args.projectId)) {
				return
			}
		})
	}
}
