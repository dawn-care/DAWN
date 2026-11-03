# EC2 deployment foundation

This creates one facility's AWS infrastructure for the existing Docker Compose
stack. It does not deploy the application, copy RIT data, or require ECS/ECR.

## Resources

- One VPC, public subnet, internet gateway and route table; no NAT gateway or load balancer.
- One Ubuntu 24.04 x86_64 EC2 instance, default `t3.medium` (2 vCPU, 4 GiB RAM).
- One encrypted 50 GiB gp3 root disk. Docker's default data directory, including
  PostgreSQL and MinIO named volumes, lives on this disk. EBS manages encryption
  through the account's default EBS KMS key; no application key service is created.
- One Elastic IP. SSH is restricted to your administrator IP; only ports 80 and
  443 are publicly accessible. PostgreSQL, MinIO and port 8080 are not exposed
  by the security group.
- An SSH public key registration, Docker/Compose installation, and `/opt/dawn`.
- Optionally, one DNS A record in an existing shared Route 53 public hosted zone.

T3 uses **standard** CPU credits to avoid surplus-credit charges. Sustained CPU
load can throttle it; confirm sizing with a realistic tablet workload.

## Before the AWS account exists

Install Terraform (1.6 or later, below 2.0). From this directory:

```sh
terraform fmt -check
terraform init -backend=false
terraform validate
```

Initialization downloads the AWS provider. These commands do not need an AWS
account and do not create resources. Commit the generated `.terraform.lock.hcl`
after initialization so the team uses the same provider version.

The example values are deliberately invalid placeholders. Do not invent AWS
credentials or disable provider credential checks to try to apply them.

## When the account is available

1. Authenticate using the sponsor's approved AWS CLI profile/SSO workflow; never
   put AWS credentials in these files. Select that profile with `AWS_PROFILE`.
2. Copy `terraform.tfvars.example` to `terraform.tfvars` and replace the AMI ID,
   administrator IPv4 `/32`, and SSH **public** key. Verify the AMI is Canonical's
   Ubuntu Server 24.04 LTS x86_64 image in your region. Keep its ID pinned;
   changing the AMI later would replace the instance.
3. If the domain is registered and its nameservers delegate to Route 53, supply
   the existing shared hosted zone ID and facility hostname. This configuration
   does not register a domain, create the shared zone, or change registrar settings.
4. Run:

```sh
terraform init
terraform plan -out=facility.tfplan
terraform apply facility.tfplan
terraform output ssh_command
```

Applying creates billable resources. After connecting as `ubuntu`, wait for setup:

```sh
sudo cloud-init status --wait
sudo docker compose version
```

If installation fails, inspect `/var/log/cloud-init-output.log`. The instance can
take a few minutes to obtain its Elastic IP and finish package installation.
Use `sudo docker`; the bootstrap does not grant the login user Docker root access.

Keep Terraform state private, backed up and out of Git. Initially state is local;
use one designated operator and preserve that directory. Before multiple people
manage production, configure an encrypted remote state backend with locking.
For another facility, use a separate working directory/state and different facility
inputs; changing `facility` in an existing deployment does not create an independent
deployment safely. VPC CIDRs repeat intentionally for isolated deployments; allocate
different ranges before connecting facility networks together.

## Application deployment is the next step

Copy/check out the approved application release into `/opt/dawn`. Do **not** start
the current production Compose unchanged: it contains the RIT API URL, development
database/MinIO credentials, and its backend's certificate allowlist is hardcoded to
`startup-da.gccis.rit.edu`. Before deploying, configure the facility hostname,
Google OAuth origins, production secrets and persistent certificate storage.

Finish HTTPS/WSS, PostgreSQL TLS and MinIO HTTPS with certificate verification;
their current internal connections are plaintext. Configure encrypted backups and
test restores before importing resident data. Restricting ports does not encrypt
connections. The Android tablet app is distributed separately from this server.

Encrypted EBS is not a claim of FIPS 140-3 compliance. Verify the applicable AWS
storage/key-management validations and each deployed TLS cryptographic module,
version, operating environment and configuration before making that claim.
Backups, retention and TLS/FIPS configuration are not implemented by these files.

## Protecting stored data

Terraform blocks destruction/replacement of the EC2 instance with `prevent_destroy`.
Changing its AMI or bootstrap script therefore requires a deliberate migration plan.
The root disk also has `delete_on_termination = false`; if an instance is terminated
outside Terraform, the disk remains billable and must be recovered/reattached
explicitly. Terraform does not automatically attach a retained disk to a replacement.
Do not remove these protections or delete state to bypass a replacement error.
These protections are not backups, and removing the entire resource configuration
also removes its Terraform lifecycle protection.

References: [EC2 resource](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance),
[offline validation](https://developer.hashicorp.com/terraform/cli/commands/validate),
[EBS encryption](https://docs.aws.amazon.com/ebs/latest/userguide/ebs-encryption.html).
