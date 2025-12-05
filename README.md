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

