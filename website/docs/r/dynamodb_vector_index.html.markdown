---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_vector_index"
description: |-
  Provides a DynamoDB Vector Index resource
---

# Resource: aws_dynamodb_vector_index

Manages a vector index on a DynamoDB table, enabling similarity search over vector embeddings stored in table items via the `SearchVectors` API.

A vector index cannot be modified in place. A change to any argument except `timeouts` replaces the index. Terraform deletes the index before it creates the new one, so `SearchVectors` requests against the index fail while the new index builds and for a short time after it becomes active. Deleting a vector index does not delete the underlying table or its items.

~> **Note:** The table must use the `PAY_PER_REQUEST` billing mode. DynamoDB does not support vector indexes on tables that use provisioned capacity.

~> **Note:** DynamoDB creates or deletes one vector index per table at a time, and creating a global secondary index on the same table counts toward that limit. While an index operation runs on a table, DynamoDB rejects the next one, and the provider retries the rejected request with exponential backoff up to the `max_retries` provider setting. If the retries run out, the apply fails with `LimitExceededException`. The `create` and `delete` timeouts do not extend these retries. A vector index build can take more than 15 minutes, even on an empty table. To add several indexes to one table without depending on these retries, use `depends_on` so that each index is created after the previous one is active.

~> **Note:** Do not add a search schema attribute to the `attribute` blocks of [`aws_dynamodb_table`](dynamodb_table.html) unless it is also a key of the table or of one of its secondary indexes. The vector index adds its search schema attributes to the table's attribute definitions, and `aws_dynamodb_table` rejects an `attribute` block that no key uses.

-> **Note:** On a global table, create the vector index once, in the Region of the `aws_dynamodb_table` resource. DynamoDB copies the index definition to every replica, including replicas added later.

## Example Usage

```terraform
resource "aws_dynamodb_vector_index" "example" {
  table_name = aws_dynamodb_table.example.name
  index_name = "DescriptionEmbeddingIndex"

  dimensions        = 1536
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "description_embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = "TenantId"
    attribute_type = "S"
    type           = "HASH"
  }
}

resource "aws_dynamodb_table" "example" {
  name         = "example"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TenantId"

  attribute {
    name = "TenantId"
    type = "S"
  }
}
```

## Argument Reference

The following arguments are required:

* `dimensions` - (Required) Number of dimensions in each vector. Valid value is between `1` and `4096`. Must match the size of the vector embeddings written to `vector_attribute.attribute_name`. Changing this value will re-create the resource.
* `distance_function` - (Required) Distance function used to calculate similarity between vectors. Valid values are `COSINE`, `DOT_PRODUCT`, or `EUCLIDEAN`. Changing this value will re-create the resource.
* `index_name` - (Required) Name of the vector index. Must be unique within the table. Changing this value will re-create the resource.
* `projection` - (Required) Attributes from the table to copy into the vector index. Changing this value will re-create the resource. See [`projection` below](#projection).
* `table_name` - (Required) Name of the table this vector index belongs to. Changing this value will re-create the resource.
* `vector_attribute` - (Required) Attribute on the table that holds the vector embedding. Changing this value will re-create the resource. See [`vector_attribute` below](#vector_attribute).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `search_schema` - (Optional) Partition key and inline filter attributes for the vector index. At most one element can have `type` set to `HASH`. Changing this value will re-create the resource. See [`search_schema` below](#search_schema).

### `projection`

* `non_key_attributes` - (Optional) Additional attributes to project into the index. Only valid when `projection_type` is `INCLUDE`.
* `projection_type` - (Required) Set of attributes represented in the index. One of `ALL`, `INCLUDE`, or `KEYS_ONLY`.

### `search_schema`

* `attribute_name` - (Required) Name of the attribute.
* `attribute_type` - (Required) Type of the attribute. Valid values are `S` (string), `N` (number), or `B` (binary). If the table already defines the attribute, for example as a key, this value must match that definition.
* `type` - (Required) Role of the attribute in the search schema.
  Valid values are `HASH` (a partition key that partitions the vector index for independent scaling; its value must be supplied in `SearchConditionExpression` when searching) or `INLINE_FILTER` (an attribute projected into the vector index for filtering at the storage layer during search; optional in `SearchConditionExpression`).

### `vector_attribute`

* `attribute_name` - (Required) Name of the table attribute that contains the vector embedding, stored as a `List` of `Number`s.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the vector index.

## Timeouts

~> **Note:** There are a variety of default timeouts set internally. If you set a shorter custom timeout than one of the defaults, the custom timeout will not be respected as the longer of the custom or internal default will be used.

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_dynamodb_vector_index.example
  identity = {
    "table_name" = "example-table"
    "index_name" = "DescriptionEmbeddingIndex"
  }
}

resource "aws_dynamodb_vector_index" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `index_name` (String) Name of the vector index.
* `table_name` (String) Name of the table this vector index belongs to.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB vector indexes using the `table_name` and `index_name`, separated by a comma. For example:

```terraform
import {
  to = aws_dynamodb_vector_index.example
  id = "example-table,DescriptionEmbeddingIndex"
}
```

Using `terraform import`, import DynamoDB vector indexes using the `table_name` and `index_name`, separated by a comma. For example:

```console
% terraform import aws_dynamodb_vector_index.example 'example-table,DescriptionEmbeddingIndex'
```
