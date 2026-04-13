resource "aws_db_subnet_group" "main" {
  name       = "${var.project}-db-subnet"
  subnet_ids = aws_subnet.private[*].id

  tags = { Name = "${var.project}-db-subnet" }
}

resource "aws_db_instance" "master" {
  identifier           = "${var.project}-master"
  engine               = "postgres"
  engine_version       = "16.4"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  db_name              = "encurtador"
  username             = "postgres"
  password             = var.db_password
  db_subnet_group_name = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.db.id]
  skip_final_snapshot    = true
  publicly_accessible    = false
  backup_retention_period = 1
  apply_immediately       = true

  tags = { Name = "${var.project}-master" }
}

resource "aws_db_instance" "replica" {
  count               = var.enable_read_replica ? 1 : 0
  identifier          = "${var.project}-replica"
  replicate_source_db = aws_db_instance.master.identifier
  instance_class      = "db.t3.micro"
  skip_final_snapshot = true
  publicly_accessible = false

  tags = { Name = "${var.project}-replica" }
}
