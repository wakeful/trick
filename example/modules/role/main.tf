# Copyright 2025 variHQ OÜ
# SPDX-License-Identifier: BSD-3-Clause

variable "name" {
  type        = string
  description = "Name of the IAM role"
}

variable "assume_from" {
  type        = string
  description = "ARN of the role that can assume this role"
}

variable "can_assume" {
  type        = string
  description = "ARN of the role that this role can assume"
}

variable "chain_start" {
  type        = bool
  default     = false
  description = "Allow role assumption from root account for testing role chaining"
}

data "aws_caller_identity" "this" {}

locals {
  account_id = data.aws_caller_identity.this.account_id
  prefix     = "trick-role"
  name       = "${local.prefix}-${var.name}"
  to_arn     = "arn:aws:iam::${local.account_id}:role/${local.prefix}-${var.can_assume}"
  from_arn   = "arn:aws:iam::${local.account_id}:role/${local.prefix}-${var.assume_from}"

  trust = var.chain_start ? [local.trust_first, local.trust_next] : [local.trust_next]

  trust_first = {
    Action = "sts:AssumeRole"
    Effect = "Allow"
    Principal = {
      AWS = "arn:aws:iam::${local.account_id}:root"
    }
    Condition = {}
  }

  trust_next = {
    Action = "sts:AssumeRole"
    Effect = "Allow"
    Principal = {
      AWS = "*"
    }
    Condition = {
      StringEquals = {
        "aws:PrincipalAccount" = local.account_id
      }
      StringLike = {
        "aws:PrincipalArn" = local.from_arn
      }
    }
  }
}

resource "aws_iam_role" "this" {
  name = local.name

  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = local.trust
  })
}

output "role_arn" {
  value = aws_iam_role.this.arn
}

resource "aws_iam_role_policy" "this" {
  name = local.name
  role = aws_iam_role.this.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sts:AssumeRole",
        ]
        Effect   = "Allow"
        Resource = "*"

        Condition = {
          StringEquals = {
            "aws:PrincipalAccount" = local.account_id
          }
          StringLike = {
            "aws:PrincipalArn" = local.to_arn
          }
        }
      },
    ]
  })
}
