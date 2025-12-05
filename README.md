# foundryvtt-cloud-server
This project requires a license for Foundry VTT: https://foundryvtt.com/

It also uses the following Cloud Services:
Compute Engine
Cloud Storage
Pub/Sub
Cloud Scheduler
Monitoring

The project ID at the time I made this was term-project-478209.

## VM Instance Setup
An E2-medium instance running Debian 12 was used for this project. 2 vCPU's and 4gb of memory meet the hosting requirements for Foundry VTT. Make sure to allow http and https traffic; SSH also needs to be enabled to set up Foundry on the Instance. Foundry also needs port 30000 as well as ports 80 and 443 if a reverse proxy is used (highly recommended).

A service account for the instance: foundry-vtt-server@term-project-478209.iam.gserviceaccount.com was also created and given the Storage Object Admin Role; this is to allow Foundry to be able to make use of Cloud Storage.

### Foundry Installation

https://foundryvtt.wiki/en/setup/linux-installation

The linux installation guide on the Foundry VTT community wiki provides the vast majority of steps needed to set up Foundry on the VM Instance.

https://www.duckdns.org/

Duck DNS was used as a reverse proxy step and requires additional steps not found in the community wiki. On the VM Instance, create a script (I named mine update_duckdns.sh) that runs: 
```
curl "https://www.duckdns.org/update?domains=DOMAIN&token=DUCK_DNS_TOKEN&> 
```
Replace DOMAIN and DUCK_DNS_TOKEN with the appropriate values. This script updates the IP address on Duck DNS for the domain to match the VM Instance's IP address, so regardless of what the Instance's IP address changes to, you can always access the server with the same url.

In order for the script to be ran whenever the server starts up, a cron process is made: 
```
crontab -e.
```
Include the following in the crontab:
```
@reboot ~/update_duckdns.sh
```
https://www.duckdns.org/install.jsp

Duck DNS suggests a more elaborate setup, however, I found that running the update script every five minutes is superfluous. The log file could be useful, and is something to be considered however.

## Cloud Storage Setup

https://foundryvtt.com/article/aws-s3/

Foundry has a tutorial on setting up S3 File Storage Integration on their website, but it was made with AWS S3 in mind so some adjustments have to be made to work with Google Cloud Storage.

When making the bucket, make sure to allow Public Access. Also, you must turn on Fine Grain Access Control; from my understanding Foundry relies on setting individual file permissions when uploading to Cloud Storage and uniform access control causes issues with it.

### CORS

The following CORS policy also needs to be applied to the bucket:

```
[
    {
      "origin": ["*"],
      "method": ["GET", "HEAD", "OPTIONS"],
      "responseHeader": ["*"],
      "maxAgeSeconds": 3600
    }
]
```
It is slightly different from the one the tutorial uses. Where as AWS uses AllowedOrigins, AllowedMethods, AllowedHeaders, Google Cloud uses origin, method, and responseHeader.
```
gcloud storage buckets update gs://${BUCKET_NAME} --cors-file=cors.json
```

### Bucket Policy

On the VM Instance running Foundry, you must create a bucket policy. The tutorial details some examples, however, to make Foundry work with Google Cloud Storage, you need to add an additional line for ACL.
```
{
  "buckets": ["nguyeant-foundry-bucket"],
  "region": "us-central-1",
  "endpoint": "https://storage.googleapis.com",
  "acl": "public-read",
  "credentials": {
    "accessKeyId": "{access key}",
    "secretAccessKey": "{secret key}"
  }
}
```
To get the access key and secret key, go to Cloud Storage -> Settings and scroll down to Access Keys for Service Account. Create a key pair for the VM Instance's service account.

<img width="833" height="780" alt="image" src="https://github.com/user-attachments/assets/a1a548ff-005e-4dce-bd89-9df9035af2ac" />

I recommend storing the policy in foundryuserdata/config that would have been made if you followed the community wiki tutorial for setting up Foundry. On Foundry, you would go to the settings and place the file path to the policy in the AWS Configuration path location.

<img width="712" height="1121" alt="image" src="https://github.com/user-attachments/assets/ca84f918-5eb5-454a-9a86-df18f6d28d56" />

In Foundry, when selecting images, there should now be an Amazon S3 tab where you can upload and select images in the bucket.

<img width="876" height="1000" alt="image" src="https://github.com/user-attachments/assets/5ff4beb4-c0e1-4e48-bffd-f156e6d007e0" />

While it does say Amazon S3, the bucket being used is the Cloud Storage Bucket we created on Google Cloud.

## Cloud Function Setup

The following cloud function was used to automatically shut off of the VM Instance, triggered by a Pub/Sub Topic.

```
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
```

It can be found, along with its go.mod and go.sum files in the functions folder of this repository. There is also a deploy_functions.sh script, however the Pub/Sub Topic must be made first before the script will work.

```
gcloud functions deploy stop-foundry-server \
--gen2 \
--runtime=go125 \
--region=us-central1 \
--source=. \
--entry-point=StopFoundryServer \
--trigger-topic=stop-foundryvtt-server
```

## Pub/Sub Setup

The payload does not matter; only that something is sent. A simple command creating the topic is all that is neeeded for the Cloud Function to run.
```
gcloud pubsub topics create stop-foundry-server
```

## Monitoring Alert Setup

An alert can be setup to send a notification to a Pub/Sub Topic and have it trigger a Cloud Function. However, the service account for Monitoring: service-35701440766@gcp-sa-monitoring-notification.iam.gserviceaccount.com must be given Pub/Sub Publisher role to do so.

### Create the Alert Policy

On Monitoring, create an Alert policy that uses the Sent Bytes from the Foundry VM Instance as the metric. Set the Rolling Window to 1 hour so the metric must meet the required threshold for 1 hour before the Alert sends a notification to Pub/Sub to shut off the server.

<img width="2288" height="926" alt="image" src="https://github.com/user-attachments/assets/6cc04d16-f742-41cc-ae19-d75ceda0a4ef" />

Now we need to configure the Condition Trigger. Set the alert trigger to Anytime the series violates; the Rolling Window will not let it trigger until the condition is met for at least one continuous hour. Now set the threshold to 50000 for an activity threshold of 50,000 kb/s. With this anytime the amount of sent bytes by the VM instance is lower than 50,000 kb/s for a continuous duration of 1 hour, an alert notifcation will be made.

<img width="812" height="875" alt="image" src="https://github.com/user-attachments/assets/cb09d978-d3e3-4cc7-948c-f5e4ae8b7152" />

Now we need to set the alert notification to be sent to the Pub/Sub topic we made earlier so enable the usage of the Notifications Channel. Then go to Manage Notifications Channels -> Pub/Sub and click Add New. Name it and select the topic you made earlier. It will ask for a path to the topic which will be:
```
projects/{Project ID}/topics/{Topic Name}
```
For me, this was:
```
projects/term-project-478209/topics/stop-foundryvtt-server
```
<img width="554" height="797" alt="image" src="https://github.com/user-attachments/assets/d483cfc9-6858-45ea-b2e5-c8da0938b5a1" />

Then hit Add Channel. Afterwards go to Notifications Channels for the Alert Policy and select the Notification Channel that you just made.

Finally, name the Alert Policy and hit Create Policy.

## Cloud Scheduler Setup

The last service to be set up is Cloud Scheduler. This is to create a backup Shutoff Trigger in the case of a player accidentally forgetting to leave the server or a malicious bot access the server.

```
gcloud scheduler jobs create pubsub foundry-midnight-shutdown \
  --schedule="0 0 * * *" \
  --timezone={Time Zone} \
  --topic={Topic Name} \
  --message-body="Midnight Shutdown" \
  --location={Region} \
```

Replace the Time Zone, Topic Name, and Region with the appropriate values. Alternatively, you can do it in the Console. The important part is picking the right time zone and when the job is triggered; "0 0 * * *" is set for Every night at midnight. The message body itself can be anything you want, but something like Midnight Shutdown would probably make more sense if you read the logs and see the job being triggered in it.

<img width="578" height="1129" alt="image" src="https://github.com/user-attachments/assets/7e8ce6a8-f34a-4a30-9479-8261aa23b57c" />

