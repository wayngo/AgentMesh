output "vpc_id" { value=aws_vpc.main.id }
output "eks_cluster_name" { value=aws_eks_cluster.main.name }
output "eks_cluster_endpoint" { value=aws_eks_cluster.main.endpoint }
output "postgres_endpoint" { value=aws_db_instance.postgres.endpoint }
output "redis_primary_endpoint" { value=aws_elasticache_replication_group.redis.primary_endpoint_address }
output "kafka_bootstrap_brokers_tls" { value=aws_msk_cluster.kafka.bootstrap_brokers_tls }
output "database_url_template" { value="postgres://agentmesh:<password>@${aws_db_instance.postgres.endpoint}/agentmesh?sslmode=require"; sensitive=true }