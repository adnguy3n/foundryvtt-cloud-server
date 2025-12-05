# foundryvtt-cloud-server
This project requires a license for Foundry VTT: https://foundryvtt.com/

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
