data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }
}

resource "aws_key_pair" "k6" {
  key_name   = "${var.project}-k6-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "aws_instance" "k6" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.xlarge"
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.k6.id, aws_security_group.app.id]
  key_name               = aws_key_pair.k6.key_name
  iam_instance_profile   = aws_iam_instance_profile.k6.name

  root_block_device {
    volume_size = 20
  }

  user_data = <<-EOF
    #!/bin/bash
    set -e

    apt-get update -y
    apt-get install -y gnupg software-properties-common docker.io

    gpg -k
    gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
    echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | tee /etc/apt/sources.list.d/k6.list
    apt-get update -y
    apt-get install -y k6

    systemctl enable docker
    systemctl start docker
    usermod -aG docker ubuntu
  EOF

  tags = { Name = "${var.project}-k6" }
}
