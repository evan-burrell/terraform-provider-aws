// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBVectorIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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

func TestAccDynamoDBVectorIndex_disappears_table(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
					acctest.CheckSDKResourceDisappears(ctx, t, tfdynamodb.ResourceTable(), resourceNameTable),
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

func TestAccDynamoDBVectorIndex_searchSchemaAttributesNotOnTable(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
				Config: testAccVectorIndexConfig_searchSchemaHashAndInlineFilter(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					testAccCheckTableHasAttributeDefinition(&conf, "category", awstypes.ScalarAttributeTypeS),
					testAccCheckTableHasAttributeDefinition(&conf, "year", awstypes.ScalarAttributeTypeN),
					resource.TestCheckResourceAttr(resourceName, "search_schema.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_name", "category"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_type", "S"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.type", "HASH"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.attribute_name", "year"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.attribute_type", "N"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.type", "INLINE_FILTER"),
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

func TestAccDynamoDBVectorIndex_searchSchemaInlineFilterFirst(t *testing.T) {
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
				Config: testAccVectorIndexConfig_searchSchemaInlineFilterFirst(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "search_schema.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_name", "year"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.type", "INLINE_FILTER"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.attribute_name", "category"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.type", "HASH"),
				),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_searchSchemaInlineFiltersOnly(t *testing.T) {
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
				Config: testAccVectorIndexConfig_searchSchemaInlineFiltersOnly(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "distance_function", "DOT_PRODUCT"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_name", "year"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.attribute_type", "N"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.0.type", "INLINE_FILTER"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.attribute_name", "brand"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.attribute_type", "S"),
					resource.TestCheckResourceAttr(resourceName, "search_schema.1.type", "INLINE_FILTER"),
				),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_noSearchSchema(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "search_schema.#", "0"),
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

func TestAccDynamoDBVectorIndex_projectionInclude(t *testing.T) {
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
				Config: testAccVectorIndexConfig_projectionInclude(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "projection.0.projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr(resourceName, "projection.0.non_key_attributes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "projection.0.non_key_attributes.*", "title"),
					resource.TestCheckTypeSetElemAttr(resourceName, "projection.0.non_key_attributes.*", "price"),
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

func TestAccDynamoDBVectorIndex_projectionKeysOnly(t *testing.T) {
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
				Config: testAccVectorIndexConfig_projectionKeysOnly(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "projection.0.projection_type", "KEYS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "projection.0.non_key_attributes.#", "0"),
				),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_multiple(t *testing.T) {
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
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_multiple(rNameTable, rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName1, &index1),
					testAccCheckVectorIndexExists(ctx, t, resourceName2, &index2),
					resource.TestCheckResourceAttr(resourceName1, "dimensions", "2"),
					resource.TestCheckResourceAttr(resourceName1, "distance_function", "COSINE"),
					resource.TestCheckResourceAttr(resourceName1, "vector_attribute.attribute_name", "embedding"),
					resource.TestCheckResourceAttr(resourceName1, "projection.0.projection_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName2, "dimensions", "3"),
					resource.TestCheckResourceAttr(resourceName2, "distance_function", "EUCLIDEAN"),
					resource.TestCheckResourceAttr(resourceName2, "vector_attribute.attribute_name", "image_embedding"),
					resource.TestCheckResourceAttr(resourceName2, "projection.0.projection_type", "KEYS_ONLY"),
				),
			},
			{
				// Both indexes are deleted in one apply and the table stays.
				Config: testAccVectorIndexConfig_tableOnly(rNameTable),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resourceName1, plancheck.ResourceActionDestroy),
						plancheck.ExpectResourceAction(resourceName2, plancheck.ResourceActionDestroy),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTableExists(ctx, t, resourceNameTable, &conf),
					testAccCheckTableVectorIndexCount(&conf, 0),
				),
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

func TestAccDynamoDBVectorIndex_withGlobalSecondaryIndex(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var index awstypes.VectorIndexDescription
	var gsi awstypes.GlobalSecondaryIndexDescription

	resourceName := "aws_dynamodb_vector_index.test"
	resourceNameGSI := "aws_dynamodb_global_secondary_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameGSI := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVectorIndexDestroy(ctx, t),
			testAccCheckGlobalSecondaryIndexDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_withGlobalSecondaryIndex(rNameTable, rName, rNameGSI),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					testAccCheckGlobalSecondaryIndexExists(ctx, t, resourceNameGSI, &gsi),
				),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_replace(t *testing.T) {
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
				Config: testAccVectorIndexConfig_dimensionsAndDistance(rNameTable, rName, 2, "COSINE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "dimensions", "2"),
					resource.TestCheckResourceAttr(resourceName, "distance_function", "COSINE"),
				),
			},
			{
				Config: testAccVectorIndexConfig_dimensionsAndDistance(rNameTable, rName, 3, "COSINE"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "dimensions", "3"),
				),
			},
			{
				Config: testAccVectorIndexConfig_dimensionsAndDistance(rNameTable, rName, 3, "EUCLIDEAN"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "distance_function", "EUCLIDEAN"),
				),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_provisionedTable(t *testing.T) {
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
				Config:      testAccVectorIndexConfig_provisionedTable(rNameTable, rName),
				ExpectError: regexache.MustCompile(`Vector\s+indexes\s+are\s+only\s+supported\s+for\s+PAY_PER_REQUEST\s+tables`),
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

func TestAccDynamoDBVectorIndex_searchSchemaMultipleHash(t *testing.T) {
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
				Config:      testAccVectorIndexConfig_searchSchemaMultipleHash(rNameTable, rName),
				ExpectError: regexache.MustCompile(`Member\s+must\s+have\s+HASH\s+count\s+less\s+than\s+or\s+equal\s+to\s+1`),
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

func TestAccDynamoDBVectorIndex_globalTable(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var index, replica awstypes.VectorIndexDescription

	resourceName := "aws_dynamodb_vector_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_globalTable(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					testAccCheckVectorIndexExistsInRegion(ctx, t, resourceName, acctest.AlternateRegion(), &replica),
					testAccCheckVectorIndexReplicaMatches(&index, &replica),
				),
			},
		},
	})
}

func TestAccDynamoDBVectorIndex_globalTable_addReplica(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var index, replica awstypes.VectorIndexDescription

	resourceNameTable := "aws_dynamodb_table.test"
	resourceName := "aws_dynamodb_vector_index.test"

	rNameTable := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckVectorIndexDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVectorIndexConfig_globalTableNoReplica(rNameTable, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceNameTable, "replica.#", "0"),
				),
			},
			{
				Config: testAccVectorIndexConfig_globalTable(rNameTable, rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceNameTable, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVectorIndexExists(ctx, t, resourceName, &index),
					resource.TestCheckResourceAttr(resourceNameTable, "replica.#", "1"),
					testAccCheckVectorIndexExistsInRegion(ctx, t, resourceName, acctest.AlternateRegion(), &replica),
					testAccCheckVectorIndexReplicaMatches(&index, &replica),
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

func testAccCheckVectorIndexExistsInRegion(ctx context.Context, t *testing.T, n, region string, v *awstypes.VectorIndexDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DynamoDBClient(ctx)
		tableName := rs.Primary.Attributes[names.AttrTableName]
		indexName := rs.Primary.Attributes["index_name"]

		// DynamoDB copies the vector index definition to each replica asynchronously.
		output, err := tfresource.RetryWhenNotFound(ctx, 10*time.Minute, func(ctx context.Context) (*awstypes.VectorIndexDescription, error) {
			table, err := tfdynamodb.FindTableByName(ctx, conn, tableName, func(o *dynamodb.Options) {
				o.Region = region
			})
			if err != nil {
				return nil, err
			}

			for _, v := range table.VectorIndexes {
				if aws.ToString(v.IndexName) == indexName {
					return &v, nil
				}
			}

			return nil, &retry.NotFoundError{}
		})
		if err != nil {
			return fmt.Errorf("DynamoDB Vector Index %s not found on table %s in %s: %w", indexName, tableName, region, err)
		}

		*v = *output

		return nil
	}
}

func testAccCheckVectorIndexReplicaMatches(primary, replica *awstypes.VectorIndexDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got, want := aws.ToInt64(replica.Dimensions), aws.ToInt64(primary.Dimensions); got != want {
			return fmt.Errorf("replica vector index has %d dimensions, expected %d", got, want)
		}

		if got, want := replica.DistanceFunction, primary.DistanceFunction; got != want {
			return fmt.Errorf("replica vector index has distance function %s, expected %s", got, want)
		}

		return nil
	}
}

func testAccCheckTableHasAttributeDefinition(table *awstypes.TableDescription, name string, attributeType awstypes.ScalarAttributeType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, v := range table.AttributeDefinitions {
			if aws.ToString(v.AttributeName) != name {
				continue
			}

			if v.AttributeType != attributeType {
				return fmt.Errorf("attribute definition %s has type %s, expected %s", name, v.AttributeType, attributeType)
			}

			return nil
		}

		return fmt.Errorf("table has no attribute definition %s", name)
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

func testAccVectorIndexConfig_searchSchemaHashAndInlineFilter(tableName, indexName string) string {
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

  search_schema {
    attribute_name = "category"
    attribute_type = "S"
    type           = "HASH"
  }

  search_schema {
    attribute_name = "year"
    attribute_type = "N"
    type           = "INLINE_FILTER"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_searchSchemaInlineFilterFirst(tableName, indexName string) string {
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

  search_schema {
    attribute_name = "year"
    attribute_type = "N"
    type           = "INLINE_FILTER"
  }

  search_schema {
    attribute_name = "category"
    attribute_type = "S"
    type           = "HASH"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_searchSchemaInlineFiltersOnly(tableName, indexName string) string {
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
  distance_function = "DOT_PRODUCT"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = "year"
    attribute_type = "N"
    type           = "INLINE_FILTER"
  }

  search_schema {
    attribute_name = "brand"
    attribute_type = "S"
    type           = "INLINE_FILTER"
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

func testAccVectorIndexConfig_projectionInclude(tableName, indexName string) string {
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
    projection_type    = "INCLUDE"
    non_key_attributes = ["title", "price"]
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_projectionKeysOnly(tableName, indexName string) string {
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
    projection_type = "KEYS_ONLY"
  }
}
`, tableName, indexName)
}

func testAccVectorIndexConfig_multiple(tableName, indexName1, indexName2 string) string {
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

  search_schema {
    attribute_name = "category"
    attribute_type = "S"
    type           = "HASH"
  }
}

resource "aws_dynamodb_vector_index" "test2" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[3]q

  dimensions        = 3
  distance_function = "EUCLIDEAN"

  vector_attribute = {
    attribute_name = "image_embedding"
  }

  projection {
    projection_type = "KEYS_ONLY"
  }

  search_schema {
    attribute_name = "category"
    attribute_type = "S"
    type           = "HASH"
  }
}
`, tableName, indexName1, indexName2)
}

func testAccVectorIndexConfig_withGlobalSecondaryIndex(tableName, indexName, gsiName string) string {
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

resource "aws_dynamodb_global_secondary_index" "test" {
  table_name = aws_dynamodb_table.test.name
  index_name = %[3]q

  projection {
    projection_type = "ALL"
  }

  key_schema {
    attribute_name = "category"
    attribute_type = "S"
    key_type       = "HASH"
  }
}
`, tableName, indexName, gsiName)
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

func testAccVectorIndexConfig_provisionedTable(tableName, indexName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "pk"

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

func testAccVectorIndexConfig_globalTableNoReplica(tableName, indexName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  billing_mode     = "PAY_PER_REQUEST"
  hash_key         = "pk"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

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

  search_schema {
    attribute_name = "category"
    attribute_type = "S"
    type           = "HASH"
  }
}
`, tableName, indexName))
}

func testAccVectorIndexConfig_globalTable(tableName, indexName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  billing_mode     = "PAY_PER_REQUEST"
  hash_key         = "pk"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "pk"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.region
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
    attribute_name = "category"
    attribute_type = "S"
    type           = "HASH"
  }
}
`, tableName, indexName))
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

func testAccVectorIndexConfig_tableOnly(tableName string) string {
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
`, tableName)
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

func testAccVectorIndexConfig_searchSchemaMultipleHash(tableName, indexName string) string {
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

  search_schema {
    attribute_name = "category"
    attribute_type = "S"
    type           = "HASH"
  }

  search_schema {
    attribute_name = "brand"
    attribute_type = "S"
    type           = "HASH"
  }
}
`, tableName, indexName)
}
