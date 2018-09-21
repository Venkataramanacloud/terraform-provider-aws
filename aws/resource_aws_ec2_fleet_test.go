package aws

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEc2Fleet_basic(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "termination"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_configs.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_template_configs.0.launch_template_specification.0.launch_template_id"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_template_configs.0.launch_template_specification.0.version"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.0.allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr(resourceName, "replace_unhealthy_instances", "false"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_interruption_behavior", "terminate"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_pools_to_use_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "spot"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.total_target_capacity", "0"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances", "false"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_with_expiration", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "maintain"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
		},
	})
}

func TestAccAWSEc2Fleet_disappears(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					testAccCheckAWSEc2FleetDisappears(&fleet1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2Fleet_ExcessCapacityTerminationPolicy(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_ExcessCapacityTerminationPolicy(rName, "no-termination"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "no-termination"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_ExcessCapacityTerminationPolicy(rName, "termination"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "termination"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_OnDemandOptions_AllocationStrategy(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_OnDemandOptions_AllocationStrategy(rName, "prioritized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.0.allocation_strategy", "prioritized"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_OnDemandOptions_AllocationStrategy(rName, "lowestPrice"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_demand_options.0.allocation_strategy", "lowestPrice"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_ReplaceUnhealthyInstances(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_ReplaceUnhealthyInstances(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "replace_unhealthy_instances", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_ReplaceUnhealthyInstances(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "replace_unhealthy_instances", "false"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_SpotOptions_AllocationStrategy(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_SpotOptions_AllocationStrategy(rName, "diversified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.allocation_strategy", "diversified"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_SpotOptions_AllocationStrategy(rName, "lowestPrice"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.allocation_strategy", "lowestPrice"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_SpotOptions_InstanceInterruptionBehavior(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_SpotOptions_InstanceInterruptionBehavior(rName, "stop"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_interruption_behavior", "stop"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_SpotOptions_InstanceInterruptionBehavior(rName, "terminate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_interruption_behavior", "terminate"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_SpotOptions_InstancePoolsToUseCount(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_SpotOptions_InstancePoolsToUseCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_pools_to_use_count", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_SpotOptions_InstancePoolsToUseCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "spot_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_options.0.instance_pools_to_use_count", "3"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_Tags(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_Tags(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_Tags(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_TargetCapacitySpecification_DefaultTargetCapacityType(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "on-demand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "on-demand"),
				),
			},
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "spot"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_TargetCapacitySpecification_DefaultTargetCapacityType_OnDemand(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "on-demand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "on-demand"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
		},
	})
}

func TestAccAWSEc2Fleet_TargetCapacitySpecification_DefaultTargetCapacityType_Spot(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, "spot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.default_target_capacity_type", "spot"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
		},
	})
}

func TestAccAWSEc2Fleet_TargetCapacitySpecification_TotalTargetCapacity(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_TotalTargetCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.total_target_capacity", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_TargetCapacitySpecification_TotalTargetCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetNotRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_specification.0.total_target_capacity", "2"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_TerminateInstancesWithExpiration(t *testing.T) {
	var fleet1, fleet2 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_TerminateInstancesWithExpiration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_with_expiration", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			{
				Config: testAccAWSEc2FleetConfig_TerminateInstancesWithExpiration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
					testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_with_expiration", "false"),
				),
			},
		},
	})
}

func TestAccAWSEc2Fleet_Type(t *testing.T) {
	var fleet1 ec2.FleetData
	resourceName := "aws_ec2_fleet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2FleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2FleetConfig_Type(rName, "maintain"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2FleetExists(resourceName, &fleet1),
					resource.TestCheckResourceAttr(resourceName, "type", "maintain"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances"},
			},
			// This configuration will fulfill immediately, skip until ValidFrom is implemented
			// {
			// 	Config: testAccAWSEc2FleetConfig_Type(rName, "request"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckAWSEc2FleetExists(resourceName, &fleet2),
			// 		testAccCheckAWSEc2FleetRecreated(&fleet1, &fleet2),
			// 		resource.TestCheckResourceAttr(resourceName, "type", "request"),
			// 	),
			// },
		},
	})
}

func testAccCheckAWSEc2FleetExists(resourceName string, fleet *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Fleet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeFleetsInput{
			FleetIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeFleets(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("EC2 Fleet not found")
		}

		for _, fleetData := range output.Fleets {
			if fleetData == nil {
				continue
			}
			if aws.StringValue(fleetData.FleetId) != rs.Primary.ID {
				continue
			}
			*fleet = *fleetData
			break
		}

		if fleet == nil {
			return fmt.Errorf("EC2 Fleet not found")
		}

		return nil
	}
}

func testAccCheckAWSEc2FleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_fleet" {
			continue
		}

		input := &ec2.DescribeFleetsInput{
			FleetIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeFleets(input)

		if isAWSErr(err, "InvalidFleetId.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if output == nil {
			continue
		}

		for _, fleetData := range output.Fleets {
			if fleetData == nil {
				continue
			}
			if aws.StringValue(fleetData.FleetId) != rs.Primary.ID {
				continue
			}
			if aws.StringValue(fleetData.FleetState) == ec2.FleetStateCodeDeleted {
				break
			}
			terminateInstances, err := strconv.ParseBool(rs.Primary.Attributes["terminate_instances"])
			if err != nil {
				return fmt.Errorf("error converting terminate_instances (%s) to bool: %s", rs.Primary.Attributes["terminate_instances"], err)
			}
			if !terminateInstances && aws.StringValue(fleetData.FleetState) == ec2.FleetStateCodeDeletedRunning {
				break
			}
			// AWS SDK constant is incorrect
			if !terminateInstances && aws.StringValue(fleetData.FleetState) == "deleted_running" {
				break
			}
			return fmt.Errorf("EC2 Fleet (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(fleetData.FleetState))
		}
	}

	return nil
}

func testAccCheckAWSEc2FleetDisappears(fleet *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteFleetsInput{
			FleetIds:           []*string{fleet.FleetId},
			TerminateInstances: aws.Bool(false),
		}

		_, err := conn.DeleteFleets(input)

		return err
	}
}

func testAccCheckAWSEc2FleetNotRecreated(i, j *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime) != aws.TimeValue(j.CreateTime) {
			return errors.New("EC2 Fleet was recreated")
		}

		return nil
	}
}

func testAccCheckAWSEc2FleetRecreated(i, j *ec2.FleetData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime) == aws.TimeValue(j.CreateTime) {
			return errors.New("EC2 Fleet was not recreated")
		}

		return nil
	}
}

func testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = "${data.aws_ami.test.id}"
  instance_type = "t3.micro"
  name          = %q
}
`, rName)
}

func testAccAWSEc2FleetConfig_ExcessCapacityTerminationPolicy(rName, excessCapacityTerminationPolicy string) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  excess_capacity_termination_policy = %q

  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, excessCapacityTerminationPolicy)
}

func testAccAWSEc2FleetConfig_OnDemandOptions_AllocationStrategy(rName, allocationStrategy string) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  on_demand_options {
    allocation_strategy = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "on-demand"
    total_target_capacity        = 0
  }
}
`, allocationStrategy)
}

func testAccAWSEc2FleetConfig_ReplaceUnhealthyInstances(rName string, replaceUnhealthyInstances bool) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  replace_unhealthy_instances = %t

  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, replaceUnhealthyInstances)
}

func testAccAWSEc2FleetConfig_SpotOptions_AllocationStrategy(rName, allocationStrategy string) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  spot_options {
    allocation_strategy = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, allocationStrategy)
}

func testAccAWSEc2FleetConfig_SpotOptions_InstanceInterruptionBehavior(rName, instanceInterruptionBehavior string) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  spot_options {
    instance_interruption_behavior = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, instanceInterruptionBehavior)
}

func testAccAWSEc2FleetConfig_SpotOptions_InstancePoolsToUseCount(rName string, instancePoolsToUseCount int) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  spot_options {
    instance_pools_to_use_count = %d
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, instancePoolsToUseCount)
}

func testAccAWSEc2FleetConfig_Tags(rName, key1, value1 string) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  tags {
    %q = %q
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, key1, value1)
}

func testAccAWSEc2FleetConfig_TargetCapacitySpecification_DefaultTargetCapacityType(rName, defaultTargetCapacityType string) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = %q
    total_target_capacity        = 0
  }
}
`, defaultTargetCapacityType)
}

func testAccAWSEc2FleetConfig_TargetCapacitySpecification_TotalTargetCapacity(rName string, totalTargetCapacity int) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  terminate_instances = true

  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = %d
  }
}
`, totalTargetCapacity)
}

func testAccAWSEc2FleetConfig_TerminateInstancesWithExpiration(rName string, terminateInstancesWithExpiration bool) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  terminate_instances_with_expiration = %t

  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, terminateInstancesWithExpiration)
}

func testAccAWSEc2FleetConfig_Type(rName, fleetType string) string {
	return testAccAWSEc2FleetConfig_BaseLaunchTemplate(rName) + fmt.Sprintf(`
resource "aws_ec2_fleet" "test" {
  type = %q

  launch_template_configs {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.test.id}"
      version            = "${aws_launch_template.test.latest_version}"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 0
  }
}
`, fleetType)
}
