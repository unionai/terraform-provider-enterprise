// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"unionai": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the scaffolding provider.
// It allows for testing assertions on data returned by an ephemeral resource during Open.
// The echoprovider is used to arrange tests by echoing ephemeral data into the Terraform state.
// This lets the data be referenced in test assertions with state checks.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"unionai": providerserver.NewProtocol6WithError(New("test")()),
	"echo":    echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
	// Check if required environment variables are set for acceptance tests
	if v := os.Getenv("UNIONAI_API_KEY"); v == "" {
		t.Skip("UNIONAI_API_KEY must be set for acceptance tests")
	}
}

// TestUnionaiProvider_New tests that the provider can be instantiated
func TestUnionaiProvider_New(t *testing.T) {
	providerFunc := New("test")
	if providerFunc == nil {
		t.Fatal("New() returned nil")
	}

	p := providerFunc()
	if p == nil {
		t.Fatal("Provider function returned nil")
	}

	// Verify it's the correct type
	if _, ok := p.(*UnionaiProvider); !ok {
		t.Fatalf("Expected *UnionaiProvider, got %T", p)
	}
}

// TestUnionaiProvider_Metadata tests the provider metadata
func TestUnionaiProvider_Metadata(t *testing.T) {
	p := New("1.0.0")()

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "unionai" {
		t.Errorf("Expected TypeName 'unionai', got '%s'", resp.TypeName)
	}

	if resp.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", resp.Version)
	}
}

// TestUnionaiProvider_Schema tests the provider schema
func TestUnionaiProvider_Schema(t *testing.T) {
	p := New("test")()

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned errors: %v", resp.Diagnostics.Errors())
	}

	// Check that api_key attribute exists
	if _, exists := resp.Schema.Attributes["api_key"]; !exists {
		t.Error("Expected 'api_key' attribute in schema")
	}
	if _, exists := resp.Schema.Attributes["org"]; !exists {
		t.Error("Expected 'org' attribute in schema")
	}
}

// TestUnionaiProvider_Resources tests that all expected resources are registered
func TestUnionaiProvider_Resources(t *testing.T) {
	p := New("test")()

	resources := p.Resources(context.Background())

	expectedResourceCount := 11
	if len(resources) != expectedResourceCount {
		t.Errorf("Expected %d resources, got %d", expectedResourceCount, len(resources))
	}

	// Test that each resource function returns a non-nil resource
	for i, resourceFunc := range resources {
		if resourceFunc == nil {
			t.Errorf("Resource function at index %d is nil", i)
			continue
		}

		resource := resourceFunc()
		if resource == nil {
			t.Errorf("Resource function at index %d returned nil", i)
		}
	}
}

// TestUnionaiProvider_DataSources tests that all expected data sources are registered
func TestUnionaiProvider_DataSources(t *testing.T) {
	p := New("test")()

	dataSources := p.DataSources(context.Background())

	expectedDataSourceCount := 12
	if len(dataSources) != expectedDataSourceCount {
		t.Errorf("Expected %d data sources, got %d", expectedDataSourceCount, len(dataSources))
	}

	// Test that each data source function returns a non-nil data source
	for i, dataSourceFunc := range dataSources {
		if dataSourceFunc == nil {
			t.Errorf("Data source function at index %d is nil", i)
			continue
		}

		dataSource := dataSourceFunc()
		if dataSource == nil {
			t.Errorf("Data source function at index %d returned nil", i)
		}
	}
}

// TestAccUnionaiProvider_Basic tests basic provider configuration
func TestAccUnionaiProvider_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUnionaiProviderConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					// This test just verifies the provider can be configured without errors
					resource.TestCheckNoResourceAttr("data.unionai_project.test", "nonexistent"),
				),
			},
		},
	})
}

// TestAccUnionaiProvider_WithApiKey tests provider configuration with explicit API key
func TestAccUnionaiProvider_WithApiKey(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("UNIONAI_API_KEY")
	if apiKey == "" {
		t.Skip("UNIONAI_API_KEY not set, skipping API key test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUnionaiProviderConfig_withApiKey(apiKey),
				Check: resource.ComposeTestCheckFunc(
					// This test verifies the provider can be configured with explicit API key
					resource.TestCheckNoResourceAttr("data.unionai_project.test", "nonexistent"),
				),
			},
		},
	})
}

// Helper function to generate basic provider configuration
func testAccUnionaiProviderConfig_basic() string {
	return `
provider "unionai" {
  # API key will be read from UNIONAI_API_KEY environment variable
}

data "unionai_project" "test" {
  id = "nelson"
}
`
}

// Helper function to generate provider configuration with explicit API key
func testAccUnionaiProviderConfig_withApiKey(apiKey string) string {
	return `
provider "unionai" {
  api_key = "` + apiKey + `"
}

data "unionai_project" "test" {
  id = "nelson"
}
`
}
