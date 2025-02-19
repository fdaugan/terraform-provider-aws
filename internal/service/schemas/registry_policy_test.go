package schemas_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/schemas"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSchemasRegistryPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_registry_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, schemas.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &schemas.GetResourcePolicyOutput{}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchemasRegistryPolicy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_registry_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, schemas.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &schemas.GetResourcePolicyOutput{}),
					acctest.CheckResourceDisappears(acctest.Provider, tfschemas.ResourceRegistry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchemasRegistryPolicy_disappears_Registry(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parentResourceName := "aws_schemas_registry.test"
	resourceName := "aws_schemas_registry_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, schemas.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &schemas.GetResourcePolicyOutput{}),
					acctest.CheckResourceDisappears(acctest.Provider, tfschemas.ResourceRegistry(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchemasRegistryPolicy_Policy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_registry_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(schemas.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, schemas.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryPolicyConfig_policy(rName, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &schemas.GetResourcePolicyOutput{}),
					testAccCheckRegistryPolicy(resourceName, "test1"),
				),
			},
			{
				Config: testAccRegistryPolicyConfig_policy(rName, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryPolicyExists(resourceName, &schemas.GetResourcePolicyOutput{}),
					testAccCheckRegistryPolicy(resourceName, "test2"),
				),
			},
		},
	})
}

func testAccCheckRegistryPolicyExists(name string, v *schemas.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EventBridge Schemas Registry ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasConn
		output, err := tfschemas.FindRegistryPolicyByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output
		return nil
	}
}

func testAccCheckRegistryPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_schemas_registry_policy" {
			continue
		}

		_, err := tfschemas.FindRegistryPolicyByName(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EventBridge Schemas Registry Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRegistryPolicy(name string, expectedSid string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		partition := acctest.Partition()
		region := acctest.Region()
		account_id := acctest.AccountID()

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EventBridge Schemas Registry ID is set")
		}

		expectedPolicyText := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Sid": %[1]q,
					"Effect": "Allow",
					"Action": [
						"schemas:*"
					],
					"Principal": {
						"AWS": %[4]q
					},
					"Resource": [
						"arn:%[2]s:schemas:%[3]s:%[4]s:registry/%[5]s"
					]
				}
			]
		}`, expectedSid, partition, region, account_id, rs.Primary.ID)

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasConn
		policy, err := tfschemas.FindRegistryPolicyByName(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		actualPolicyText, _ := structure.FlattenJsonToString(policy.Policy)

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccRegistryPolicyConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_schemas_registry" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRegistryPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRegistryPolicyConfigBase(rName),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    sid    = "test"
    effect = "Allow"
    principals {
      identifiers = [data.aws_caller_identity.test.account_id]
      type        = "AWS"
    }
    actions   = ["schemas:*"]
    resources = [aws_schemas_registry.test.arn]
  }
}

resource "aws_schemas_registry_policy" "test" {
  registry_name = %[1]q
  policy        = data.aws_iam_policy_document.test.json
}
`, rName),
	)
}

func testAccRegistryPolicyConfig_policy(rName string, sid string) string {
	return acctest.ConfigCompose(
		testAccRegistryPolicyConfigBase(rName),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    sid    = %[2]q
    effect = "Allow"
    principals {
      identifiers = [data.aws_caller_identity.test.account_id]
      type        = "AWS"
    }
    actions   = ["schemas:*"]
    resources = [aws_schemas_registry.test.arn]
  }
}

resource "aws_schemas_registry_policy" "test" {
  registry_name = %[1]q
  policy        = data.aws_iam_policy_document.test.json
}
`, rName, sid),
	)
}
