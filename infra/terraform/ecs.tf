resource "aws_ecs_cluster" "main" {
  name = "${var.project}-cluster"
  tags = { Name = "${var.project}-cluster" }
}

resource "aws_iam_role" "ecs_task_execution" {
  name = "${var.project}-ecs-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.project}"
  retention_in_days = 1
}

resource "aws_ecs_task_definition" "app" {
  family                   = var.project
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "1024"
  memory                   = "2048"
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn

  container_definitions = jsonencode([{
    name  = var.project
    image = var.app_image
    portMappings = [{ containerPort = 8080, protocol = "tcp" }]
    environment = [
      { name = "SERVER_PORT", value = ":8080" },
      { name = "BASE_URL", value = var.enable_alb ? "http://${aws_lb.main[0].dns_name}" : "http://localhost:8080" },
      { name = "ALLOWED_ORIGINS", value = "*" },
      { name = "DATABASE_URL", value = "postgres://postgres:${var.db_password}@${aws_db_instance.master.endpoint}/encurtador?sslmode=require" },
      { name = "DATABASE_READ_URL", value = var.enable_read_replica ? "postgres://postgres:${var.db_password}@${aws_db_instance.replica[0].endpoint}/encurtador?sslmode=require" : "postgres://postgres:${var.db_password}@${aws_db_instance.master.endpoint}/encurtador?sslmode=require" },
      { name = "REDIS_URL", value = var.enable_redis ? "redis://${aws_elasticache_cluster.redis[0].cache_nodes[0].address}:6379" : "" },
      { name = "REDIS_START_OFFSET", value = "14000000" },
      { name = "HASH_SALT", value = var.hash_salt },
      { name = "HASH_PEPPER", value = var.hash_pepper },
      { name = "HASH_MIN_LENGTH", value = "6" },
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.app.name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

resource "aws_ecs_service" "app" {
  name            = var.project
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = var.ecs_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private[*].id
    security_groups  = [aws_security_group.app.id]
    assign_public_ip = false
  }

  dynamic "load_balancer" {
    for_each = var.enable_alb ? [1] : []
    content {
      target_group_arn = aws_lb_target_group.app[0].arn
      container_name   = var.project
      container_port   = 8080
    }
  }

  depends_on = [aws_db_instance.master]
}

resource "aws_appautoscaling_target" "ecs" {
  count              = var.enable_autoscaling ? 1 : 0
  max_capacity       = var.ecs_max_count
  min_capacity       = var.ecs_desired_count
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.app.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "cpu" {
  count              = var.enable_autoscaling ? 1 : 0
  name               = "${var.project}-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 50.0
    scale_in_cooldown  = 60
    scale_out_cooldown = 30
  }
}
