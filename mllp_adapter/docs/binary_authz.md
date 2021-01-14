# Enable Binary Authorization with MLLP Adapter Deployment

## Goal
* A codelab demonstrating how to use the GCP (Google Cloud Platform)
[Binary Authorization](https://cloud.google.com/binary-authorization/) service as part of MLLP
Deployment.
  - Binary Authorization adds deploy-time policy enforcement to users' Kubernetes Engine Cluster,
  i.e. only approved images attested by trusted parties (called "attestors") can be deployed.
  Cluster managers can use this service to prevent untrusted images being deployed.
* A multi-project setup of using GCP Binary Authorization service.
  - The official [Binary Authorization Codelab](https://codelabs.developers.google.com/codelabs/cloud-binauthz-intro/)
  assumes a single project setup, which might not be suitable for most real-world uses of Binary
  Authorization. There is a multi-project setup in the
  [Binary Authorization Document](https://cloud.google.com/binary-authorization/docs/multi-project-setup-cli),
  however, it turns out some GCP projects can be split further to fit even finer granularity.

## Background
* [MLLP Adapter](https://github.com/GoogleCloudPlatform/mllp)
  * The [Readme](https://github.com/GoogleCloudPlatform/mllp/blob/master/README.md) has the
  instructions about deploying a MLLP Adapter cluster **without** Binary Authorization.
  * Public Container Registry: [gcr.io/cloud-healthcare-containers/mllp-adapter](https://console.cloud.google.com/gcr/images/cloud-healthcare-containers/GLOBAL/mllp-adapter)

* [Binary Authorization References](https://cloud.google.com/binary-authorization/docs/)
  * Make sure you are familiar with the concept and use of Binary Authorization.
    * [Official CodeLab of Binary Authorization (Single Project Setup)](https://codelabs.developers.google.com/codelabs/cloud-binauthz-intro/#0)
    * [Official Multi-Project Setup Document](https://cloud.google.com/binary-authorization/docs/multi-project-setup-cli)

## Components
Here is the list of all the interacting components in this codelab. The details of the interaction
will be demonstrated in Setup and Deployment section.
### GCP Projects
This codelab tries to split the responsibilities at a fine granularity to allow maximum flexibility
for different security and governance practices by creating a project for each role.
However, this does not mean a setup must have that many projects. Details of merging projects are
demonstrated in the [Combining Project section](#combining-gcp-projects).

#### Deployer Project

The project that owns the GKE cluster. The binary authorization policy will be imported to and
stored in this project.
  * Project ID: `${DEPLOYER_PROJ}`
  * Project Number: `${DEPLOYER_PROJ_NUM}`

#### Container Project

The project that holds the container registry, in our case, the project holds MLLP adapter images.
Note that in real world there might be more than one container registry to be used by the cluster.
Here we take just one project as an example.
  * Project ID: `${CONTAINER_PROJ}`
  * Project Number: `${CONTAINER_PROJ_NUM}`

Using Binary Authorization service does not require any specific modification to Container
Project(s). Using or not using the Binary Authorization service is totally transparent to Container
Project(s).

#### Note Project

The project that owns the container analysis note.
  * Project ID: `${NOTE_PROJ}`
  * Project Number: `${NOTE_PROJ_NUM}`

#### Attestor Project

The project that stores the attestors.
  * Project ID: `${ATTESTOR_PROJ}`
  * Project Number: `${ATTESTOR_PROJ_NUM}`

#### Attestation Project

The project that stores attestations.
  * Project ID: `${ATTESTATION_PROJ}`
  * Project ID: `${ATTESTATION_PROJ_NUM}`

#### KMS Project
(optional)

The project that provides PKIX signature with Google Cloud Key Management Service. For a setup not
using Google Cloud Key Management Service (e.g. manage private/public key pairs locally), this
project is not required.
  * Project ID: `${KMS_PROJ}`
  * Project ID: `${KMS_PROJ_NUM}`

### Resources, Concepts
#### Container Analysis Note

Created and stored in the [Note Project](#note-project), identified by `${NOTE_ID}`.

#### Attestor

Created and stored in the [Attestor Project](#attestor-project), identified by `${ATTESTOR_ID}`.

#### Attestation

Created and stored in the [Attestation Project](#attestation-project), identified by `${ATTESTATION_ID}`.

#### Cluster Name/Zone

GKE cluster terminology, assuming their values are `${CLUSTER_NAME}`, `${CLUSTER_ZONE}`.

The cluster is managed in the [Deployer Project](#deployer-project).

#### Keyring/Key/KeyVersion/KeyringLocation (optional)

Managed in the [KMS Project](#kms-project). Only necessary if using Google Cloud Key Management Service.
Assuming their values to be:
  * `${KEY_RING}`
  * `${KEY}`
  * `${KEY_VERSION}`
  * `${KEY_RING_LOCATION}`
In this codelab, we assume the key ring location to be “global” (i.e. `KEY_RING_LOCATION=global`)

## Preparation
### 1. GCP Projects
Create the projects listed in GCP Projects section.
Reference: [Creating and Managing Projects](https://cloud.google.com/resource-manager/docs/creating-managing-projects)

### 2. MLLP Adapter Image
In this codelab we will be using the existing MLLP Adapter Image at
[gcr.io/cloud-healthcare-containers/mllp-adapter](https://console.cloud.google.com/gcr/images/cloud-healthcare-containers/GLOBAL/mllp-adapter).
`${CONTAINER_PROJ}` will be `cloud-healthcare-containers`.

The link above lists all versions of the MLLP Adapter images, identified by a piece of the digest of
each version. One can copy the full image names here by clicking the button shows up when hovering
on the image digest.

At the time this document is created, we have following images in the registry. We will take them as
example images in this document.
* gcr.io/cloud-healthcare-containers/mllp-adapter@sha256:231b073df13db0c65e57b0e1d526ab6816a73c37262e25c18bcca99bf4b4b185
  - Take this one as `${IMAGE_SIGNED}` in this document
* gcr.io/cloud-healthcare-containers/mllp-adapter@sha256:edea3487df8a00e75a471cfe1ee3c3c8032f0339421ab00e376acf841d1408a1
  - Take this one as `${IMAGE_UNSIGNED}` in this document

One can also checkout MLLP Adapter code, customize it and build a docker image. Docker image
building and publishing is out of the scope of this document.

## Setup Attestor on Attestor Project

**Use of `--project=` flag**

Each of the following steps with gcloud command will have a `--project=` flag set, to indicate on
which project the operation is done. When `--project` is not specified, `gcloud` will use the
default project set in the current environment. To check the current default project, use
`gcloud config get-value project`. However, in this multi-project setup codelab, we will work on
different projects, so it is recommended to use `--project=` flag for each gcloud command. In the
real world, those projects might be managed separately, in such cases, saving the `--project=` flag
and rely on the default project might do a better job.

**Use of Google KMS service**

In this codelab, we assume using Google KMS service to provide the PKIX keys, however, it is not
mandatory to use Google KMS service.

### 1. Enable Container Analysis API on the Note Project
```shell
gcloud --project=${NOTE_PROJ} services enable containeranalysis.googleapis.com
```

### 2. Create Container Analysis Note in the Note Project
```shell
## Assume we dump the note payload to ./note_payload.json
cat > ./note_payload.json << EOM
{
  "name": "projects/${NOTE_PROJ}/notes/${NOTE_ID}",
  "attestation": {
    "hint": {
      "human_readable_name": "Attestor Note"
    }
  }
}
EOM

## Note the noteId should match with the note ID used in the payload.
curl -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $(gcloud auth print-access-token)"  \
    --data @./note_payload.json  \
    "https://containeranalysis.googleapis.com/v1/projects/${NOTE_PROJ}/notes/?noteId=${NOTE_ID}"

## Verify the note has been created
curl \
-H "Authorization: Bearer $(gcloud auth print-access-token)" \
"https://containeranalysis.googleapis.com/v1/projects/${NOTE_PROJ}/notes/${NOTE_ID}"
```

### 3. Enable Google Key Management Service (KMS) on the KMS Project
```shell
gcloud --project=${KMS_PROJ} services enable cloudkms.googleapis.com
```

### 4. Create Keyring, Key pair and verify the Keyversion on the KMS Project
```shell
## Create key ring
gcloud --project=${KMS_PROJ} kms keyrings create ${KEY_RING} --location=${KEY_RING_LOCATION}
## Create key pair
gcloud --project=${KMS_PROJ} kms keys create ${KEY} \
--keyring=${KEY_RING} \
--location=${KEY_RING_LOCATION} \
--purpose asymmetric-signing \
--default-algorithm="ec-sign-p256-sha256"
## Verify the key version, it should be 1 (i.e. KEY_VERSION=1)
gcloud --project=${KMS_PROJ} kms keys versions list --location=${KEY_RING_LOCATION} --key=${KEY} --keyring=${KEY_RING}
```

### 5. Enable Binary Authorization API on the Attestor Project
```
gcloud --project=${ATTESTOR_PROJ} services enable binaryauthorization.googleapis.com
```

### 6. Create Attestor on the Attestor Project, the attestor should use the note created in the Note Project for attestation.
```shell
gcloud --project=${ATTESTOR_PROJ} beta container binauthz attestors create ${ATTESTOR_ID} \
  --attestation-authority-note=${NOTE_ID} \
  --attestation-authority-note-project=${NOTE_PROJ}
## Verify the created attestor
gcloud --project=${ATTESTOR_PROJ} beta container binauthz attestors list
```

### 7. Set the IAM permission so that the Attestor Project’s Binary Authorization service account can read the Container Analysis Note Occurrences in the Note Project.
```shell
cat > iam_request.json << EOM
{
  "resource": "projects/${NOTE_PROJ}/notes/${NOTE_ID}",
  "policy": {
    "bindings": [
      {
        "role": "roles/containeranalysis.notes.occurrences.viewer",
        "members": [
          "serviceAccount:service-${ATTESTOR_PROJ_NUM}@gcp-sa-binaryauthorization.iam.gserviceaccount.com"
        ]
      }
    ]
  }
}
EOM

curl -X POST  \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $(gcloud auth print-access-token)" \
    --data @./iam_request.json \
"https://containeranalysis.googleapis.com/v1/projects/${NOTE_PROJ}/notes/${NOTE_ID}:setIamPolicy"
```

### 8. Set IAM permission so that the Binary Authorization Service Account on the Deployer Project can access the attestor for attestation verification.
```shell
gcloud --project=${ATTESTOR_PROJ} \
    beta container binauthz attestors add-iam-policy-binding \
    "projects/${ATTESTOR_PROJ}/attestors/${ATTESTOR_ID}" \
    --member="serviceAccount:service-${DEPLOYER_PROJ_NUM}@gcp-sa-binaryauthorization.iam.gserviceaccount.com" \
    --role=roles/binaryauthorization.attestorsVerifier
```

### 9. Enable Google KMS service on the Attestor Project. This is necessary if the attestor is to use key pairs managed by Google KMS service (stored in the KMS project). Otherwise this step can be skipped.
```shell
gcloud --project=${ATTESTOR_PROJ} services enable binaryauthorization.googleapis.com
```

### 10. Add key generated in KMS project to the attestor
```shell
gcloud --project=${ATTESTOR_PROJ} container binauthz attestors public-keys add  \
    --attestor=${ATTESTOR_ID}  \
    --keyversion-project=${KMS_PROJ}  \
    --keyversion-location=${KEY_RING_LOCATION} \
    --keyversion-keyring=${KEY_RING} \
    --keyversion-key=${KEY} \
    --keyversion=${KEY_VERSION}
```

## Sign and Create Attestation on Attestation Project
### 1. Enable Binary Authorization API on the Attestation Project
```shell
gcloud --project=${ATTESTATION_PROJ} services enable binaryauthorization.googleapis.com
```

### 2. Sign and create attestation
```shell
## --artifact-url= points to the mllp-adapter image to be signed, following the form:
##      gcr.io/${CONTAINER_PROJ}/mllp-adapter@{DIGEST}
## which is exactly the full name of the image copied from GCP console.
gcloud beta container binauthz attestations sign-and-create \
    --project=${ATTESTATION_PROJ} \
    --artifact-url=${IMAGE_SIGNED} \
    --attestor=${ATTESTOR_ID} \
    --attestor-project=${ATTESTOR_PROJ} \
    --keyversion-project=${KMS_PROJ} \
    --keyversion-location=${KEY_RING_LOCATION} \
    --keyversion-keyring=${KEY_RING} \
    --keyversion-key=${KEY} \
    --keyversion=${KEY_VERSION}
```

## Deploying MLLP Adapter
### 1. Enable Binary Authorization API on the Deployer Project
```shell
gcloud --project=${DEPLOYER_PROJ} services enable binaryauthorization.googleapis.com
```

### 2. Create a cluster with --enable-binauthz in the Deployer Project
```shell
gcloud beta --project=${DEPLOYER_PROJ} container clusters create --enable-binauthz --zone ${CLUSTER_ZONE} ${CLUSTER_NAME}
```

### 3. Create a deployment policy and import it to the Deployer Project
```shell
## Dump the following policy to ./policy.yaml, which will be imported later.
## This policy puts some image sources to the whitelist, so that any images from them
## are allowed to be deployed.
## This policy also sets a project scope default rule which will block the images from
## non-whitelisted sources that have not been attested by the attestor we created before.
cat > policy.yaml << EOM
    admissionWhitelistPatterns:
    - namePattern: gcr.io/google_containers/*
    - namePattern: gcr.io/google-containers/*
    - namePattern: k8s.gcr.io/*
    - namePattern: gcr.io/stackdriver-agents/*
    defaultAdmissionRule:
      evaluationMode: REQUIRE_ATTESTATION
      enforcementMode: ENFORCED_BLOCK_AND_AUDIT_LOG
      requireAttestationsBy:
        - projects/${ATTESTOR_PROJ}/attestors/${ATTESTOR_ID}
    name: projects/${DEPLOYER_PROJ}/policy
EOM

## import the policy
gcloud --project=${DEPLOYER_PROJ} beta container binauthz policy import policy.yaml
```

The policy can be checked and modified  in the GCP console sidebar:
${DEPLOYER_PROJ} -> Security -> Binary Authorization.
More example policies can be found [here](https://cloud.google.com/binary-authorization/docs/example-policies).

### 4. Checkout the GKE cluster credentials
```shell
gcloud --project=${DEPLOYER_PROJ} container clusters get-credentials --zone ${CLUSTER_ZONE} ${CLUSTER_NAME}
```

### 5. Try to create a deployment with the unattested image to verify that it is rejected
Sample deployment YAML file
```yaml
# Not a complete file, needs to provide image's full URL
#   (form: gcr.io/cloud-healthcare-containers/mllp-adapter@${digest})
# hl7_v2_project_id, hl7_v2_location_id, and other arguments are required for having the
# MLLP adapter actually working, but not strictly necessary to be correct in this codelab.
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mllp-adapter-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mllp-adapter
  template:
    metadata:
      labels:
        app: mllp-adapter
    spec:
      containers:
        - name: mllp-adapter
          imagePullPolicy: Always
          image: ${IMAGE_UNSIGNED}
          ports:
            - containerPort: 2575
              protocol: TCP
              name: "port"
          command:
            - "/usr/mllp_adapter/mllp_adapter"
            - "--hl7_v2_project_id=<your-hl7v2-store-project-id>"
            - "--hl7_v2_location_id=<your-hl7v2-store-location-id>"
            - "--hl7_v2_dataset_id=<your-hl7v2-dataset-id>"
            - "--hl7_v2_store_id=<your-hl7v2-store-id>"
            - "--logtostderr"
            - "--receiver_ip=0.0.0.0"
```

Run following commands:
```shell
## With image set to ${IMAGE_UNSIGNED} in the deployment.yaml
kubectl create -f deployment.yaml
## Get pods should returns no running pods
kubectl get pods
## Event should show the deployment is denied
kubectl get event
```
On the GCP console, Kubernetes Engine -> Workloads page, the deployment: 'mllp-adapter-deployment'
should have an error status showing the deployment is denied due to unverified image.

### 6. Try to create a deployment with the attested image
Change the image used in the deployment.yaml to ${IMAGE_SIGNED}, then:
```shell
## delete the previous deployment
kubectl delete deployment mllp-adapter-deployment
```
Repeat the commands in step 5:
`kubectl get pods` should show one running pod.

`kubectl get event` should show ‘Scaled up replica set mllp-adapter-deployment-xxxx to 1’

GCP console Kubernetes Engine -> Workloads page should show the deployment status is ‘OK’.

### 7. Try to create a deployment with image specified by tag
Change the image field value to `gcr.io/cloud-healthcare-containers/mllp-adapter:latest`

The only change to the ${IMAGE_SIGNED} should be replacing the digest string with ‘:latest’.

Then delete the deployment (like in step 6) and repeat the commands in step 5.

The deployment should be denied, `kubectl get event` and console workload page should show the error.

The reason is, ‘tag’ is not allowed to be used to identify an image to pass binary authorization,
because the version identified by that tag could change.

### 8. Override a policy
Binary Authorization allows users to override an authorization policy when deploying a container
image, a.k.a. *break-glass* feature. The feature behaves consistently with the recommendations
in the Kubernetes [admission controller](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
specification.

Below shows the change to the deployment YAML file, to override the policy and allow the unsigned
image to be deployed.
```yaml
    <unchanged content>
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mllp-adapter
      ## Begin of change
      annotations:
        alpha.image-policy.k8s.io/break-glass: "true"
      ## End of change
    spec:
      containers:
        - name: mllp-adapter
          imagePullPolicy: Always
          image: ${IMAGE_UNSIGNED}
          ports:
            - containerPort: 2575
              protocol: TCP
              name: "port"
    <unchanged content>
```

Note that, unlike the [official document](https://cloud.google.com/binary-authorization/docs/deploying-containers#override_a_policy),
the annotation is added at `spec.templates.metadata.annotations`, not at `metadata.annotations`,
because this is a **Deployment** config file (specified by `kind`), not a **Pod** config file.

Now `kubectl create -f deployment.yaml` should be able to spin up the pods.

*break-glass* deployment will write audit log for the Kubernetes Cluster (k8s_cluster) resource,
**if** the image **fails** to pass the required checks. The following two tags can be used to filter
the Stackdriver Logging:

All the operations with *break-glass* behavior:
```shell
imagepolicywebhook.image-policy.k8s.io/break-glass: "true"
```

```shell
imagepolicywebhook.image-policy.k8s.io/overridden-verification-result: "${IMAGE_UNSIGNED}: Image ${IMAGE_UNSIGNED} denied by projects/${ATTESTOR_PROJ}/attestors/${ATTESTOR_ID}: ${reason}

# For container images not passing the check ${reason} will be:
"No attestations found that were valid and signed by a key trusted by the attestor"

# For using 'tag' (e.g. mllp-adapter:latest) instead of full name (with digest) to specify the
# container image, the ${reason} will be:
"Attestor cannot attest to an image deployed by tag"
```

For more details, please refer to the [official document](https://cloud.google.com/binary-authorization/docs/viewing-audit-logs#allowed_deployment_break-glass).


## Cleanup
Delete the cluster
```shell
gcloud --project=${DEPLOYER_PROJ} container clusters delete ${CLUSTER_NAME} --zone=${CLUSTER_ZONE}
```

Then delete all the projects.

## Combining GCP Projects
GCP projects can be combined at will. But the two IAM permission setting steps
(Setup Attestor step 7 and 8) need more attention.

If the `${ATTESTOR_PROJ}` and the `${DEPLOYER_PROJ}` is one project, there is **No Need** to add the
IAM binding which gives the deployer project’s binary authorization service account the role:
**roles/binaryauthorization.attestorsVerifier**.

If the `${ATTESTOR_PROJ}` and the `${NOTE_PROJ}` is one project. It is **Still Required** to grant
the role **roles/containeranalysis.notes.occurrences.viewer** to the binary authorization service
account.

## Compare with Normal Deployment
* Container Registry
  * Nothing changed
* Cluster
  * The owner project must enable Binary Authorization API
  * The Cluster must be created with --enable-binauthz
  * The owner project will store deployment policy
* New Players (Attestation flow)
  * Attestor
    * Uses the asymmetric key pairs to sign attested images
  * Attestation
  * Container Analysis Note
  * KMS (Optional)
    * Only needed if the user wants to use KMS to provide the key pairs

## Not Using KMS
If the user does not want Google KMS to manage the keys, he or she can set up the key pair locally,
sign the image locally, upload the public key to Attestor Project to be bound with an Attestor and
the signature file to the Attestation to be created in the Attestation Project.

Following steps will need to be changed:
* [Setup Attestor on Attestor Project](#setup-attestor-on-attestor-project)
  * Step 3
    SKIP: No need to enable KMS service, also no need to create KMS project.
  * Step 4
    CHANGE: Keys should be created locally,  Official Document.
  * Step 9
    SKIP: No need to enable KMS
  * Step 10
    CHANGE: Different command line arguments to add public key info
    * Please refer to: [Official Document: Create The Attestor](https://cloud.google.com/binary-authorization/docs/creating-attestors-cli#create_the_attestor).
* [Sign and Create Attestation on Attestation Project](#sign-and-create-attestation-on-attestion-project)
  * Step 2
    `sign-and-create` can not be used. The signing will happen locally.
    Please refer to: [Official Document: Create Attestation with PGP Signature](https://cloud.google.com/binary-authorization/docs/making-attestations#create_an_attestation_with_a_pgp_signature)

However, the general workflow stay unchanged.
  * Prepare signing keys
  * Let the attestor get the public key
  * Get Attestation payload.
  * Sign with the private key.
  * Upload the signature file to create Attestation.

Using Google KMS simplifies the steps of key pair preparation, signing and creating attestation.
