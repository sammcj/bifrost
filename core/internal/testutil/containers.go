// Package testutil provides container API test utilities for the Bifrost system.
package testutil

import (
	"context"
	"testing"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
)

// RunContainerCreateTest tests the container create functionality
func RunContainerCreateTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig ComprehensiveTestConfig) {
	if !testConfig.Scenarios.ContainerCreate {
		t.Logf("[SKIPPED] Container Create: Not supported by provider %s", testConfig.Provider)
		return
	}

	t.Run("ContainerCreate", func(t *testing.T) {
		t.Logf("[RUNNING] Container Create test for provider: %s", testConfig.Provider)

		request := &schemas.BifrostContainerCreateRequest{
			Provider: testConfig.Provider,
			Name:     "bifrost-test-container",
		}

		bfCtx := schemas.NewBifrostContext(ctx, schemas.NoDeadline)
		response, err := client.ContainerCreateRequest(bfCtx, request)

		if err != nil {
			// Check if this is an unsupported operation error
			if err.Error != nil && (err.Error.Code != nil && *err.Error.Code == "unsupported_operation") {
				t.Logf("[EXPECTED] Provider %s returned unsupported operation error", testConfig.Provider)
				return
			}
			t.Fatalf("❌ ContainerCreate failed: %v", GetErrorMessage(err))
		}

		if response == nil {
			t.Fatal("❌ ContainerCreate returned nil response")
		}

		if response.ID == "" {
			t.Fatal("❌ ContainerCreate returned empty container ID")
		}

		t.Logf("✅ Container Create test passed for provider: %s, container ID: %s", testConfig.Provider, response.ID)

		// Clean up: delete the created container
		deleteRequest := &schemas.BifrostContainerDeleteRequest{
			Provider:    testConfig.Provider,
			ContainerID: response.ID,
		}
		_, deleteErr := client.ContainerDeleteRequest(bfCtx, deleteRequest)
		if deleteErr != nil {
			t.Logf("[WARNING] Failed to clean up container %s: %v", response.ID, GetErrorMessage(deleteErr))
		}
	})
}

// RunContainerListTest tests the container list functionality
func RunContainerListTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig ComprehensiveTestConfig) {
	if !testConfig.Scenarios.ContainerList {
		t.Logf("[SKIPPED] Container List: Not supported by provider %s", testConfig.Provider)
		return
	}

	t.Run("ContainerList", func(t *testing.T) {
		t.Logf("[RUNNING] Container List test for provider: %s", testConfig.Provider)

		request := &schemas.BifrostContainerListRequest{
			Provider: testConfig.Provider,
			Limit:    10,
		}

		bfCtx := schemas.NewBifrostContext(ctx, schemas.NoDeadline)
		response, err := client.ContainerListRequest(bfCtx, request)

		if err != nil {
			// Check if this is an unsupported operation error
			if err.Error != nil && (err.Error.Code != nil && *err.Error.Code == "unsupported_operation") {
				t.Logf("[EXPECTED] Provider %s returned unsupported operation error", testConfig.Provider)
				return
			}
			t.Fatalf("❌ ContainerList failed: %v", GetErrorMessage(err))
		}

		if response == nil {
			t.Fatal("❌ ContainerList returned nil response")
		}

		t.Logf("✅ Container List test passed for provider: %s, found %d containers", testConfig.Provider, len(response.Data))
	})
}

// RunContainerRetrieveTest tests the container retrieve functionality
func RunContainerRetrieveTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig ComprehensiveTestConfig) {
	if !testConfig.Scenarios.ContainerRetrieve {
		t.Logf("[SKIPPED] Container Retrieve: Not supported by provider %s", testConfig.Provider)
		return
	}

	t.Run("ContainerRetrieve", func(t *testing.T) {
		t.Logf("[RUNNING] Container Retrieve test for provider: %s", testConfig.Provider)

		// First, create a container to retrieve
		createRequest := &schemas.BifrostContainerCreateRequest{
			Provider: testConfig.Provider,
			Name:     "bifrost-test-container-retrieve",
		}

		bfCtx := schemas.NewBifrostContext(ctx, schemas.NoDeadline)
		createResponse, createErr := client.ContainerCreateRequest(bfCtx, createRequest)

		if createErr != nil {
			if createErr.Error != nil && (createErr.Error.Code != nil && *createErr.Error.Code == "unsupported_operation") {
				t.Logf("[EXPECTED] Provider %s returned unsupported operation error", testConfig.Provider)
				return
			}
			t.Fatalf("❌ ContainerCreate (setup) failed: %v", GetErrorMessage(createErr))
		}

		if createResponse == nil || createResponse.ID == "" {
			t.Fatal("❌ ContainerCreate (setup) returned nil or empty response")
		}

		containerID := createResponse.ID
		defer func() {
			// Clean up
			deleteRequest := &schemas.BifrostContainerDeleteRequest{
				Provider:    testConfig.Provider,
				ContainerID: containerID,
			}
			_, _ = client.ContainerDeleteRequest(bfCtx, deleteRequest)
		}()

		// Now retrieve the container
		retrieveRequest := &schemas.BifrostContainerRetrieveRequest{
			Provider:    testConfig.Provider,
			ContainerID: containerID,
		}

		response, err := client.ContainerRetrieveRequest(bfCtx, retrieveRequest)

		if err != nil {
			t.Fatalf("❌ ContainerRetrieve failed: %v", GetErrorMessage(err))
		}

		if response == nil {
			t.Fatal("❌ ContainerRetrieve returned nil response")
		}

		if response.ID != containerID {
			t.Fatalf("❌ ContainerRetrieve returned wrong container ID: expected %s, got %s", containerID, response.ID)
		}

		t.Logf("✅ Container Retrieve test passed for provider: %s, container ID: %s", testConfig.Provider, response.ID)
	})
}

// RunContainerDeleteTest tests the container delete functionality
func RunContainerDeleteTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig ComprehensiveTestConfig) {
	if !testConfig.Scenarios.ContainerDelete {
		t.Logf("[SKIPPED] Container Delete: Not supported by provider %s", testConfig.Provider)
		return
	}

	t.Run("ContainerDelete", func(t *testing.T) {
		t.Logf("[RUNNING] Container Delete test for provider: %s", testConfig.Provider)

		// First, create a container to delete
		createRequest := &schemas.BifrostContainerCreateRequest{
			Provider: testConfig.Provider,
			Name:     "bifrost-test-container-delete",
		}

		bfCtx := schemas.NewBifrostContext(ctx, schemas.NoDeadline)
		createResponse, createErr := client.ContainerCreateRequest(bfCtx, createRequest)

		if createErr != nil {
			if createErr.Error != nil && (createErr.Error.Code != nil && *createErr.Error.Code == "unsupported_operation") {
				t.Logf("[EXPECTED] Provider %s returned unsupported operation error", testConfig.Provider)
				return
			}
			t.Fatalf("❌ ContainerCreate (setup) failed: %v", GetErrorMessage(createErr))
		}

		if createResponse == nil || createResponse.ID == "" {
			t.Fatal("❌ ContainerCreate (setup) returned nil or empty response")
		}

		containerID := createResponse.ID

		// Now delete the container
		deleteRequest := &schemas.BifrostContainerDeleteRequest{
			Provider:    testConfig.Provider,
			ContainerID: containerID,
		}

		response, err := client.ContainerDeleteRequest(bfCtx, deleteRequest)

		if err != nil {
			t.Fatalf("❌ ContainerDelete failed: %v", GetErrorMessage(err))
		}

		if response == nil {
			t.Fatal("❌ ContainerDelete returned nil response")
		}

		if !response.Deleted {
			t.Fatal("❌ ContainerDelete returned deleted=false")
		}

		t.Logf("✅ Container Delete test passed for provider: %s, container ID: %s", testConfig.Provider, containerID)
	})
}

// RunContainerUnsupportedTest tests that providers correctly return unsupported operation errors
func RunContainerUnsupportedTest(t *testing.T, client *bifrost.Bifrost, ctx context.Context, testConfig ComprehensiveTestConfig) {
	// Only run this test if none of the container operations are supported
	if testConfig.Scenarios.ContainerCreate || testConfig.Scenarios.ContainerList ||
		testConfig.Scenarios.ContainerRetrieve || testConfig.Scenarios.ContainerDelete {
		t.Logf("[SKIPPED] Container Unsupported: Provider %s supports container operations", testConfig.Provider)
		return
	}

	t.Run("ContainerUnsupported", func(t *testing.T) {
		t.Logf("[RUNNING] Container Unsupported test for provider: %s", testConfig.Provider)

		bfCtx := schemas.NewBifrostContext(ctx, schemas.NoDeadline)

		// Test ContainerCreate returns unsupported
		createRequest := &schemas.BifrostContainerCreateRequest{
			Provider: testConfig.Provider,
			Name:     "test-container",
		}
		_, createErr := client.ContainerCreateRequest(bfCtx, createRequest)

		if createErr == nil {
			t.Fatal("❌ Expected unsupported operation error for ContainerCreate, got nil")
		}

		if createErr.Error == nil || createErr.Error.Code == nil || *createErr.Error.Code != "unsupported_operation" {
			t.Fatalf("❌ Expected unsupported_operation error code, got: %v", createErr)
		}

		t.Logf("✅ Container Unsupported test passed for provider: %s", testConfig.Provider)
	})
}
