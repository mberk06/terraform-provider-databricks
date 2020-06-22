package databricks

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/stretchr/testify/assert"

	"github.com/databrickslabs/databricks-terraform/client/model"
	"github.com/databrickslabs/databricks-terraform/client/service"
)

func TestAccDatabricksPermissionsResourceFullLifecycle(t *testing.T) {
	var permissions model.ObjectACL
	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// create a resource
				Config: testClusterPolicyPermissions(randomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("databricks_permissions.dummy_can_use",
						"object_type", "cluster-policy"),
					testAccIDCallback(t, "databricks_permissions.dummy_can_use",
						func(client *service.DBApiClient, id string) error {
							resp, err := client.Permissions().Read(id)
							if err != nil {
								return err
							}
							permissions = *resp
							assert.Len(t, permissions.AccessControlList, 3)
							return nil
						}),
				),
			},
			{
				Config: testClusterPolicyPermissionsSecondGroupAdded(randomName),
				Check: testAccIDCallback(t, "databricks_permissions.dummy_can_use",
					func(client *service.DBApiClient, id string) error {
						resp, err := client.Permissions().Read(id)
						if err != nil {
							return err
						}
						permissions = *resp
						assert.Len(t, permissions.AccessControlList, 3)
						return nil
					}),
			},
		},
	})
}

func testClusterPolicyPermissions(name string) string {
	return fmt.Sprintf(`
	resource "databricks_cluster_policy" "something_simple" {
		name = "Terraform Policy %[1]s"
		definition = jsonencode({
			"spark_conf.spark.hadoop.javax.jdo.option.ConnectionURL": {
				"type": "forbidden"
			}
		})
	}
	resource "databricks_scim_group" "dummy_group" {
		display_name = "Terraform Group %[1]s"
	}
	resource "databricks_permissions" "dummy_can_use" {
		cluster_policy_id = databricks_cluster_policy.something_simple.id
		access_control {
			group_name = databricks_scim_group.dummy_group.display_name
			permission_level = "CAN_USE"
		}
	}
	`, name)
}

func testClusterPolicyPermissionsSecondGroupAdded(name string) string {
	return fmt.Sprintf(`
	resource "databricks_cluster_policy" "something_simple" {
		name = "Terraform Policy %[1]s"
		definition = jsonencode({
			"spark_conf.spark.hadoop.javax.jdo.option.ConnectionURL": {
				"type": "forbidden"
			},
			"spark_conf.spark.secondkey": {
				"type": "forbidden"
			}
		})
	}
	resource "databricks_scim_group" "dummy_group" {
		display_name = "Terraform Group %[1]s"
	}
	resource "databricks_scim_group" "second_group" {
		display_name = "Terraform Second Group %[1]s"
	}
	resource "databricks_permissions" "dummy_can_use" {
		cluster_policy_id = databricks_cluster_policy.something_simple.id
		access_control {
			group_name = databricks_scim_group.dummy_group.display_name
			permission_level = "CAN_USE"
		}
		access_control {
			group_name = databricks_scim_group.second_group.display_name
			permission_level = "CAN_USE"
		}
	}
	`, name)
}

func TestAccNotebookPermissions(t *testing.T) {
	randomName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// create a resource
				Config: fmt.Sprintf(`
				resource "databricks_notebook" "dummy" {
					content = base64encode("# Databricks notebook source\nprint(1)")
					path = "/Beginning/Init"
					overwrite = true
					mkdirs = true
					language = "PYTHON"
					format = "SOURCE"
				}
				resource "databricks_scim_group" "dummy_group" {
					display_name = "Terraform Group %[1]s"
				}
				resource "databricks_permissions" "dummy_can_use" {
					directory_path = "/Beginning"
					access_control {
						group_name = databricks_scim_group.dummy_group.display_name
						permission_level = "CAN_MANAGE"
					}
				}
				`, randomName),
			},
		},
	})
}
