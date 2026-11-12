# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dynamodb_vector_index" "test" {
  count = var.resource_count

  table_name = aws_dynamodb_table.test.name
  index_name = "${var.rName}-${count.index}"

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = var.rName
    attribute_type = "S"
    type           = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  name         = var.rName
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = var.rName

  attribute {
    name = var.rName
    type = "S"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}
