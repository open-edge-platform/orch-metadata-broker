/*
* SPDX-FileCopyrightText: (C) 2023 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"context"
	"fmt"
	"github.com/go-playground/assert/v2"
	client "github.com/open-edge-platform/orch-metadata-broker/pkg/restClient"
	"testing"
)

func TestRestClientGet(t *testing.T) {
	c, err := client.NewClientWithResponses("http://localhost:5900")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	res, err := c.MetadataServiceGetMetadataWithResponse(context.TODO())
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	fmt.Println(res.JSON200.Metadata)
	assert.Equal(t, res.StatusCode(), 200)
}
func TestRestClientPost(t *testing.T) {
	c, err := client.NewClientWithResponses("http://localhost:5900")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	created, err := c.MetadataServiceCreateOrUpdateMetadataWithResponse(context.TODO(), client.MetadataServiceCreateOrUpdateMetadataJSONRequestBody{
		Metadata: []client.Metadata{
			{Key: "customer", Value: "culvers"},
			{Key: "customer", Value: "menards"},
		},
	})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	assert.Equal(t, created.StatusCode(), 200)
}

func TestRestClientDelete(t *testing.T) {
	c, err := client.NewClientWithResponses("http://localhost:5900")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	key := "customer"
	value := "menards"
	created, err := c.MetadataServiceDeleteWithResponse(context.TODO(), &client.MetadataServiceDeleteParams{
		Key: &key, Value: &value,
	})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	assert.Equal(t, created.StatusCode(), 200)
}
