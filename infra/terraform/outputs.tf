output "ecr_repository_url" {
  value = aws_ecr_repository.app.repository_url
}

output "rds_endpoint" {
  value = aws_db_instance.master.endpoint
}

output "rds_replica_endpoint" {
  value = var.enable_read_replica ? aws_db_instance.replica[0].endpoint : "disabled"
}

output "redis_endpoint" {
  value = var.enable_redis ? aws_elasticache_cluster.redis[0].cache_nodes[0].address : "disabled"
}

output "alb_dns" {
  value = var.enable_alb ? aws_lb.main[0].dns_name : "disabled"
}

output "k6_public_ip" {
  value = aws_instance.k6.public_ip
}

output "k6_ssh" {
  value = "ssh ubuntu@${aws_instance.k6.public_ip}"
}
