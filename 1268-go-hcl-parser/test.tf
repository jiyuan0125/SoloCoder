provider "aws" {
  region = "us-east-1"
  profile = "default"
}

variable "instance_count" {
  type = number
  default = 2
}

resource "aws_instance" "web" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
  count         = var.instance_count

  tags = {
    Name = "web-server-${count.index}"
    Environment = "prod"
  }

  user_data = <<-EOF
#!/bin/bash
echo "Hello from ${var.instance_name}"
apt-get update -y
EOF
}

data "aws_ami" "latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amazon-linux-ami-*"]
  }
}

output "public_ip" {
  value = aws_instance.web[0].public_ip
}

locals {
  base_tags = {
    Project = "my-project"
    Owner   = "infra-team"
  }
}

module "vpc" {
  source = "./modules/vpc"

  cidr_block = var.vpc_cidr
  private_subnets = [
    "10.0.1.0/24",
    "10.0.2.0/24",
  ]
}
