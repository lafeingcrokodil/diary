# infra

Terraform configuration for the app's GCP resources, in the project
`lafeingcrokodil-diary` and the region `europe-west9` (Paris). Only the bucket
for Terraform's state is in `europe-west10` (Berlin), where the app was
originally hosted: a bucket's location can't be changed, and moving the state
to a new bucket isn't worth it.

Requires Terraform 1.17 or later and the [gcloud CLI](https://cloud.google.com/sdk/docs/install).

## One-off setup

These steps create what Terraform needs before it can run. They only need to
be done once.

1. Create the project and link it to your billing account (`gcloud billing
   accounts list` shows the account ID):

   ```bash
   gcloud projects create lafeingcrokodil-diary --name="Diary"
   ```

   ```bash
   gcloud billing projects link lafeingcrokodil-diary --billing-account=<billing-account-id>
   ```

2. Create the bucket for Terraform's state, with versioning so that earlier
   states can be recovered:

   ```bash
   gcloud storage buckets create gs://lafeingcrokodil-diary-tfstate \
     --project=lafeingcrokodil-diary \
     --location=europe-west10 \
     --uniform-bucket-level-access \
     --public-access-prevention
   ```

   ```bash
   gcloud storage buckets update gs://lafeingcrokodil-diary-tfstate --versioning
   ```

3. Let Terraform use your credentials, with the project as the quota project:

   ```bash
   gcloud auth application-default login
   ```

   ```bash
   gcloud auth application-default set-quota-project lafeingcrokodil-diary
   ```

4. Tell Terraform which billing account pays for the project, by creating
   `infra/terraform.tfvars` (it's ignored by Git):

   ```hcl
   billing_account = "<billing-account-id>"
   ```

   Creating the budget alert requires permission to manage budgets on the
   billing account, which the account's administrators have.

## Applying changes

```bash
cd infra
terraform init
terraform plan
terraform apply
```

After applying, `terraform output service_url` prints the app's URL. Until the
app is deployed, the service runs Google's placeholder "hello" image.

CI checks the formatting and validity of the configuration on every pull
request, but doesn't apply it.

## Deployment

Every push to `main` deploys the app once CI passes: the `deploy` job in
`.github/workflows/ci.yml` builds the image, pushes it to Artifact Registry
and updates the Cloud Run service. It authenticates through workload identity
federation, which only accepts workflow runs for `main` in this repository, so
no keys are stored in GitHub.

The job needs two repository variables (not secrets, since neither value is
sensitive). After the first `terraform apply`, set them from the Terraform
outputs, either in the repository's settings (Settings → Secrets and variables
→ Actions → Variables) or with the [GitHub CLI](https://cli.github.com/):

```bash
gh variable set GCP_WORKLOAD_IDENTITY_PROVIDER --body "$(terraform output -raw workload_identity_provider)"
```

```bash
gh variable set GCP_DEPLOYER_SERVICE_ACCOUNT --body "$(terraform output -raw deployer_service_account)"
```

## Budget

When the project's spending reaches 50%, 90% and 100% of the monthly budget
(€5 or $5 by default, in the billing account's currency), or is forecast to
exceed it, the billing account's administrators and users get an email. To
change the amount, set `monthly_budget` in `terraform.tfvars`.
