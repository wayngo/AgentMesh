variable "aws_region" { type = string; default = "us-west-2" }
variable "project_name" { type = string; default = "agentmesh" }
variable "environment" { type = string; default = "dev" }
variable "vpc_cidr" { type = string; default = "10.40.0.0/16" }
variable "availability_zones" { type = list(string); default = ["us-west-2a", "us-west-2b"] }
variable "eks_version" { type = string; default = "1.31" }
variable "db_instance_class" { type = string; default = "db.t4g.micro" }
variable "redis_node_type" { type = string; default = "cache.t4g.small" }
variable "kafka_instance_type" { type = string; default = "kafka.t3.small" }
variable "tags" { type = map(string); default = {} }