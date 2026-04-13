variable "region" {
  default = "us-east-1"
}

variable "project" {
  default = "encurtador"
}

variable "db_password" {
  sensitive = true
}

variable "hash_salt" {
  sensitive = true
}

variable "hash_pepper" {
  sensitive = true
}

variable "vpc_cidr" {
  default = "10.0.0.0/16"
}

variable "app_image" {
  description = "ECR image URI"
  default     = ""
}

variable "ecs_desired_count" {
  default = 1
}

variable "ecs_max_count" {
  default = 10
}

variable "enable_redis" {
  default = false
}

variable "enable_alb" {
  default = false
}

variable "enable_read_replica" {
  default = false
}

variable "enable_autoscaling" {
  default = false
}
