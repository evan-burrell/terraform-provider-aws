// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBVectorIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var index awstypes.VectorIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_vector_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "dynamodb", "table/{table_name}/index/{index_name}"),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rNameTable),
					resource.TestCheckResourceAttr(resourceName, "dimensions", "2"),
					resource.TestCheckResourceAttr(resourceName, "distance_function", "COSINE"),
					resource.TestCheckResourceAttr(resourceName, "vector_attribute.attribute_name", "embedding"),
					resource.TestCheckResourceAttr(resourceName, "projection.0.projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_name", rNameTable),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_type", "S"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.type", "HASH"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccVectorIndexImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.TableDescription
	var index awstypes.VectorIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_vector_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_basic(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdynamodb.ResourceVectorIndex, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_createWhileDeleting(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.TableDescription
	var index1, index2 awstypes.VectorIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName1 := "aws_dynamodb_vector_index.test1"
	resourceName2 := "aws_dynamodb_vector_index.test2"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.14.0",
			},
		},
		CheckDestroy: testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_createWhileDeleting1(rNameTable, rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName1, &index1),
				),
			},
			{
				// The delete of test1 and the create of the table item start together, once
				// the AWS provider is configured. time_sleep then delays the create of test2 by
				// 1.5 seconds. In three measurements on indexes created with their table, a
				// deleted vector index stayed DELETING for 2.9 to 5.4 seconds, so the first poll
				// of the create normally sees test1 DELETING.
				Config: testAccVectorIndexConfig_createWhileDeleting2(rNameTable, rName2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionDestroy),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName2, &index2),
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckTableVectorIndexCount(&conf, 1),
				),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_searchSchemaTypeMismatch(t *testing.T) {
	ctx := acctest.Context(t)

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccVectorIndexConfig_searchSchemaTypeMismatch(rNameTable, rName),
				ExpectError: regexache.MustCompile(`Invalid Search Schema Type Change(?s:.*)attribute\s+"pk"\s+was\s+previously\s+defined\s+as\s+"S"(?s:.*)Invalid Search Schema Type Change(?s:.*)attribute\s+"sk"\s+was\s+previously\s+defined\s+as\s+"N"`),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_dimensionsOutOfRange(t *testing.T) {
	ctx := acctest.Context(t)

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccVectorIndexConfig_dimensionsAndDistance(rNameTable, rName, 4097, "COSINE"),
				ExpectError: regexache.MustCompile(`Attribute\s+dimensions\s+value\s+must\s+be\s+between\s+1\s+and\s+4096,\s+got:\s+4097`),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_timeouts(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var index awstypes.VectorIndexDescription

	resourceName := "aws_dynamodb_vector_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_noSearchSchema(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
				),
			},
			{
				Config: testAccVectorIndexConfig_timeouts(rNameTable, rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "timeouts.create", "45m"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "15m"),
				),
			},
		},
	})
}

func testAccCheckVectorIndexDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_vector_index" {
				continue
			}

			tableName := rs.Primary.Attributes[names.AttrTableName]
			indexName := rs.Primary.Attributes["index_name"]

			_, err := tfdynamodb.FindVectorIndexByTwoPartKey(ctx, conn, tableName, indexName)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("DynamoDB Vector Index %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVectorIndexExists(ctx context.Context, t *testing.T, n string, v *awstypes.VectorIndexDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)
		tableName := rs.Primary.Attributes[names.AttrTableName]
		indexName := rs.Primary.Attributes["index_name"]

		output, err := tfdynamodb.FindVectorIndexByTwoPartKey(ctx, conn, tableName, indexName)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTableVectorIndexCount(table *awstypes.TableDescription, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got := len(table.VectorIndexes); got != want {
			return fmt.Errorf("table %s has %d vector indexes, expected %d", aws.ToString(table.TableName), got, want)
		}

		return nil
	}
}

func testAccVectorIndexImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		parts := []string{
			rs.Primary.Attributes[names.AttrTableName],
			rs.Primary.Attributes["index_name"],
		}

		return strings.Join(parts, intflex.ResourceIdSeparator), nil
	}
}

func testAccVectorIndexConfig_basic(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}

resource "aws_dynamodb_vector_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = %[1]q
    attribute_type = "S"
    type           = "HASH"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_noSearchSchema(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"

  attribute {
    name = "pk"
    type = "S"
  }
}

resource "aws_dynamodb_vector_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_dimensionsAndDistance(tableName, indexName string, dimensions int, distanceFunction string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"

  attribute {
    name = "pk"
    type = "S"
  }
}

resource "aws_dynamodb_vector_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = %[3]d
  distance_function = %[4]q

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }
}
`, tableName, indexName, dimensions, distanceFunction)
}

func testAccVectorIndexConfig_searchSchemaTypeMismatch(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"
  range_key    = "sk"

  attribute {
    name = "pk"
    type = "S"
  }

  attribute {
    name = "sk"
    type = "N"
  }
}

resource "aws_dynamodb_vector_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = "pk"
    attribute_type = "N"
    type           = "HASH"
  }

  search_schema {
    attribute_name = "sk"
    attribute_type = "S"
    type           = "INLINE_FILTER"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_timeouts(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"

  attribute {
    name = "pk"
    type = "S"
  }
}

resource "aws_dynamodb_vector_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  timeouts {
    create = "45m"
    delete = "15m"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_createWhileDeleting1(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"

  attribute {
    name = "pk"
    type = "S"
  }
}

resource "aws_dynamodb_vector_index" "test1" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_createWhileDeleting2(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = %[1]q
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"

  attribute {
    name = "pk"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key

  item = <<ITEM
{
  "pk": {"S": "test"}
}
ITEM
}

resource "time_sleep" "test" {
  create_duration = "1.5s"

  depends_on = [aws_dynamodb_table_item.test]
}

resource "aws_dynamodb_vector_index" "test2" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[2]q

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  depends_on = [time_sleep.test]
}
`, tableName, indexName)
}
