gcloud functions deploy stop-foundry-server \
--gen2 \
--runtime=go125 \
--region=us-central1 \
--source=. \
--entry-point=StopFoundryServer \
--trigger-topic=stop-foundryvtt-server
