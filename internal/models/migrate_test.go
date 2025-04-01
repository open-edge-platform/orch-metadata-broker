package models

//
//import (
//	"fmt"
//	"github.com/open-edge-platform/orch-metadata-broker/test"
//	"github.com/stretchr/testify/assert"
//	"testing"
//)
//
//var emptyFile = []byte("")
//var v0 = []byte(`{"keys":[{"name":"foo","values":["bar","baz"]}]}`)
//var v1 = []byte(`
//{
//	"version":"v1",
//	"projects":{
//		"sampleProject":{
//			"keys":[
//				{
//					"name":"foo",
//					"values":["bar","baz"]
//				}
//			]
//		}
//	}
//}`)
//
//func TestMigrate(t *testing.T) {
//	type args struct {
//		data []byte
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    string
//		wantErr assert.ErrorAssertionFunc
//	}{
//		{"empty-file", args{data: emptyFile}, `{"version":"v1","projects":{"sampleProject":{"keys":null}}}`, assert.NoError},
//		{"migrate", args{data: v0}, string(v1), assert.NoError},
//		{"no-migration", args{data: v1}, "", assert.NoError},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			res, err := Migrate(tt.args.data, projectId)
//			tt.wantErr(t, err, fmt.Sprintf("UnexpectedError in Migrate(%v): %s", tt.args.data, err))
//			assert.Equal(t, test.RemoveAllSpaces(tt.want), test.RemoveAllSpaces(string(res)))
//		})
//	}
//}
