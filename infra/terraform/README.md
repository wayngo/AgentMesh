# Terraform AWS infrastructure

Phase 14 provisions the AWS foundation for AgentMesh: VPC networking, EKS, RDS PostgreSQL, ElastiCache Redis, and Amazon MSK Kafka.

## Safe workflow

```powershell
cd infra/terraform
Copy-Item terraform.tfvars.example terraform.tfvars
terraform init
terraform fmt -check
terraform validate
terraform plan -out agentmesh.tfplan
```

Review the plan before applying. Applying creates billable AWS resources and requires AWS credentials and approval:

```powershell
terraform apply agentmesh.tfplan
```

Destroying the environment is also billable/destructive and should be explicitly reviewed:

```powershell
terraform plan -destroy
```

The generated database password is stored in Terraform state. Use an encrypted remote backend and restrict state access before using this outside a sandbox. Feed Terraform outputs into the Helm values for the Kubernetes deployment.