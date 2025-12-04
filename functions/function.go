package function

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

func init() {
	functions.CloudEvent("StopFoundryServer", StopFoundryServer)
}

func StopFoundryServer(ctx context.Context, e event.Event) error {
	// To identify which VM Instance will be shut down.
	projectID := "term-project-478209"
	instanceName := "foundry-vtt-server"
	zone := "us-central1-c"

	// Inits a compute engine client.
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("Error with Client Init: %w", err)
	}

	defer client.Close()

	// Create the Stop Request.
	req := &computepb.StopInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: instanceName,
	}

	// Stop the Foundry Server.
	if _, err := client.Stop(ctx, req); err != nil {
		return fmt.Errorf("Error with stopping instance: %w", err)
	}

	return nil
}
