# MLLP Adapter

The MLLP(Short for "Minimal Lower Layer Protocol") adapter is a component that
runs on [GKE](https://cloud.google.com/kubernetes-engine/), receives HL7v2
messages via MLLP/TCP, and forwards received messages to HL7v2 API.

## Requirements

*   A [Google Cloud project](https://cloud.google.com).
*   A [Docker](https://docs.docker.com/) repository. The following instructions
    assume the use of
    [Google Container Registry](https://cloud.google.com/container-registry/).
*   Installed [gcloud](https://cloud.google.com/sdk/gcloud/) and
    [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command
    line tools.

## Initial Pull

The prebuilt docker image is staged on Google Container Registry (GCR). It is
strongly recommended to use the prebuilt docker images:

```bash
docker pull gcr.io/cloud-healthcare-containers/mllp-adapter:latest
```

Note that the 'latest' image tag is convenient for development and testing, but
production deployments should use a stable image tag to ensure that a new image
isn't unintentionally deployed whenever the mllp-adapter image is updated.

## Build (Optional)

We use bazel as the build tool. Please refer to
[the bazel documentation](https://docs.bazel.build/versions/master/getting-started.html)
to get started.

Run the following commands to build the MLLP adapter binary:

```bash
cd mllp_adapter
bazel build :mllp_adapter
```

Or to build and load a docker image that contains the binary:

```bash
bazel run :mllp_adapter_docker
```

If docker is installed on your system, you should be able to see the image:

```bash
docker images
```

## Test Locally

To run the image locally:

```bash
docker run -p 127.0.0.1:2575:2575 -v ~/.config:/root/.config gcr.io/cloud-healthcare-containers/mllp-adapter /usr/mllp_adapter/mllp_adapter --hl7_v2_project_id=<PROJECT_ID> --hl7_v2_location_id=<LOCATION_ID> --hl7_v2_dataset_id=<DATASET_ID> --hl7_v2_store_id=<STORE_ID> --export_stats=false --receiver_ip=0.0.0.0 --pubsub_project_id=<PUBSUB_PROJECT_ID> --pubsub_subscription=<PUBSUB_SUBSCRIPTION_ID>
```

In the command above:
* `-p 127.0.0.1:2575:2575` is used to publish the port `2575` of the MLLP
container to host port `2575`. By default MLLP adapter listen to port `2575`;
* `-v ~/.config:/root/.config` is used to give the container access
to gcloud credentials;

Also note that:
* `PUBSUB_PROJECT_ID` and `PUBSUB_SUBSCRIPTION_ID` are available
by creating a pubsub topic and a subscription on Google Cloud;

You should be able to send HL7v2 messages now:

```bash
# This will fail because the format is invalid.
echo -n -e '\x0btestmessage\x1c\x0d' | telnet localhost 2575
```

> **_NOTE:_** Older versions of the MLLP adapter subscribed to the single
> Pub/Sub topic configured in an HL7v2 store's `notification_config` field, and
> sent an outgoing message if a Pub/Sub notification had the "publish" attribute
> set (i.e. notification sent by the messages.create method). The new
> recommended method is to use the HL7v2 store's configurable
> `notification_configs` field to route messages to separate Pub/Sub topics and
> configure the MLLP adapter to subscribe to a topic that contains only messages
> it should send.
>
> To continue using the old behavior, set the `--legacy_publish_attribute=true`
> flag when running the image. This functionality is deprecated and the publish
> attribute will be removed in a future release.

## Deployment

### Use Customized Service Account

If you want to use a custom service account instead of the default GCE service
account, make sure the custom service account is passed as `--service-account`
parameter while creating the cluster (See the
[doc](https://cloud.google.com/sdk/gcloud/reference/container/clusters/create)).

The custom service account will require the following permissions to work
properly: * roles/pubsub.subscriber. This is required only if you need the MLLP
adapter to listen for pubsub notifications and forward the HL7v2 messages to
HL7v2 stores. This role only needs to be set on a per-subscription basis. *
roles/healthcare.hl7V2Ingest. * roles/monitoring.metricWriter. This is for
writing metrics to Stackdriver. You don't need this if `--exportStats` is set to
`false`.

Before deploying the docker image to GKE you need to publish the image to a
registry. First modify `BUILD.bazel` to replace `my-project` and `my-image` with
real values, then run:

```bash
bazel run :image_push
```

If this fails with the error message "ModuleNotFoundError: No module named
'urlparse'", you are hitting a python 3 bug in the google/containerregistry repo
(https://github.com/google/containerregistry/issues/42). Fix this by setting:

```bash
BAZEL_PYTHON=python2.7
```

If the push fails with "Permission denied", try running the following the
command in the console and retry:

```bash
gcloud auth configure-docker
```

Next replace the placeholders in the
[deployment config](config/mllp_adapter.yaml) for your use case.

Note that the <IMAGE_LABEL> should be a stable image tag that identifies the
version of mllp-adapter image you want deployed. Do not use the 'latest' image
tag, as this points to a new image each time there is an update.

You can retrieve the list of mllp-adapter image tags using the following
command:

```bash
gcloud container images list-tags gcr.io/cloud-healthcare-containers/mllp-adapter

DIGEST        TAGS         TIMESTAMP
edea3487df8a               1969-12-31T19:00:00
231b073df13d  blue,latest  1969-12-31T19:00:00
```

In the above example, the image tag 'blue' will be the stable image tag for the
latest mllp-adapter image at the time of execution.

Deploy to a GKE cluster:

```bash
# To use default service account.
gcloud container clusters create mllp-adapter --zone=<ZONE_ID> --scopes https://www.googleapis.com/auth/pubsub
# Or to use custom service account. Read the documentation here: https://cloud.google.com/sdk/gcloud/reference/container/clusters/create
gcloud container clusters create mllp-adapter --zone=<ZONE_ID> --service-account <SERVICE_ACCOUNT>

# See the documentation here: https://kubernetes.io/docs/reference/kubectl/kubectl/
kubectl create -f config/mllp_adapter.yaml
kubectl create -f config/mllp_adapter_service.yaml
```

### Use Binary Authorization Service
[Google Binary Authorization Service](https://cloud.google.com/binary-authorization/) can be applied as
a deploy-time security control to ensure only trusted container images can be deployed. Please refer to
[Enable Binary Authorization with MLLP Adapter Deployment](docs/binary_authz.md) for details of setup.

## VPN

*Use E2E VPN setup if want your data to be encrypted end-to-end. See
[the "encryption in transit" doc](https://cloud.google.com/security/encryption-in-transit/)
for more details.*

### Cloud VPN

[Cloud VPN](https://cloud.google.com/vpn/docs/) creates a secure tunnel to
ensure data is encrypted in transit.

First create a static IP address for the VPN gateway (GATEWAY_IP):

```bash
gcloud --project <PROJECT_ID> compute addresses create <IP_NAME> --region=<REGION_ID>
```

Then set up the VPN:

```bash
gcloud --project <PROJECT_ID> compute target-vpn-gateways create mllp-vpn --region <REGION_ID> --network <VPC_NAME>
gcloud --project <PROJECT_ID> compute forwarding-rules create "mllp-vpn-rule-esp" --region <REGION_ID> --address <GATEWAY_IP> --ip-protocol "ESP" --target-vpn-gateway "mllp-vpn"
gcloud --project <PROJECT_ID> compute forwarding-rules create "mllp-vpn-rule-udp500" --region <REGION_ID> --address <GATEWAY_IP> --ip-protocol "UDP" --ports "500" --target-vpn-gateway "mllp-vpn"
gcloud --project <PROJECT_ID> compute forwarding-rules create "mllp-vpn-rule-udp4500" --region <REGION_ID> --address <GATEWAY_IP> --ip-protocol "UDP" --ports "4500" --target-vpn-gateway "mllp-vpn"
```

Next set up the tunnels (also in the opposite direction which isn't shown here):

```bash
gcloud --project <PROJECT_ID>  compute vpn-tunnels create "tunnel-1" --region <REGION_ID>  --peer-address <PEER_IP_ADDRESS> --shared-secret <SHARED_SECRET> --ike-version "2" --target-vpn-gateway "mllp-vpn" --local-traffic-selector="10.100.0.0/24"
gcloud --project <PROJECT_ID>  compute routes create "tunnel-1-route-1" --network <VPC_NAME> --next-hop-vpn-tunnel "tunnel-1" --next-hop-vpn-tunnel-region <REGION_ID> --destination-range "10.101.0.0/24"
```

Finally configure the firewall to allow traffic if necessary:

```bash
gcloud --project <PROJECT_ID> compute firewall-rules create allow-mllp-over-vpn --network <VPC_NAME> --allow tcp:32577,icmp --source-ranges "10.101.0.0/24"
```

### E2E VPN

The docker VPN image used in this section is
[here on github](https://github.com/kitten/docker-strongswan). First check the
configurations in the repo to see if you need to make any changes for your use
case (you probabaly do).

Pull the image from docker hub and upload it to gcr.io:

```bash
# Build the image from source code instead of pulling it if you made changes in previous step.
docker pull philplckthun/strongswan
docker tag philplckthun/strongswan:latest gcr.io/<GCR_PROJECT_ID>/<IMAGE_NAME>:<IMAGE_LABEL>
gcloud docker -- push gcr.io/<GCR_PROJECT_ID>/<IMAGE_NAME>:<IMAGE_LABEL>
```

Replace the placeholders in the
[deployment config](config/mllp_adapter_e2e.yaml).

Load necessary kernel modules before applying the configuration changes. Then
apply the changes:

```bash
kubectl apply -f config/mllp_adapter_e2e.yaml
```

(Optional) Allocate a static IP address for the load balancer:

```bash
gcloud --project <PROJECT_ID> compute addresses create vpn-load-balancer --region=us-central1
```

Update the [service config](config/mllp_adapter_service_e2e.yaml) to expose it
as an external load balancer.

Apply the change as well by running `kubectl apply -f
config/mllp_adapter_service_e2e.yaml`

This will also update the firewall rules automatically.

Now connect your client to the VPN server and test if you can access the
adapter:

```bash
kubectl describe pods | grep IP: # Get pod IP.
echo -n -e '\x0btestmessage\x1c\x0d' | telnet <POD_IP> 2575
```

## Debug

To view the running status and logs of the pod:

```bash
kubectl get pods
kubectl logs <POD_ID>
```
