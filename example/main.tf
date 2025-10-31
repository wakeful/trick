# Copyright 2025 variHQ OÜ
# SPDX-License-Identifier: BSD-3-Clause

terraform {
  required_version = "1.13.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.18.0"
    }
  }
}

provider "aws" {
  region = "eu-west-1"
}

module "role_a" {
  source = "./modules/role"

  name        = "a"
  chain_start = true

  assume_from = "c"
  can_assume  = "b"
}

output "role_arn_a" {
  value = module.role_a.role_arn
}

module "role_b" {
  source = "./modules/role"

  name        = "b"
  assume_from = "a"
  can_assume  = "c"
}

output "role_arn_b" {
  value = module.role_b.role_arn
}

module "role_c" {
  source = "./modules/role"

  name        = "c"
  assume_from = "b"
  can_assume  = "a"
}

output "role_arn_c" {
  value = module.role_c.role_arn
}
