package linode

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/linode/linodego"
)

func TestAccLinodeInstance_basic(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceBasic(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "image", "linode/ubuntu18.04"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "swap_size", "256"),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstance_config(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceWithConfig(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					// resource.TestCheckResourceAttr(resName, "kernel", "linode/latest-64bit"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "swap_size", "0"),
					testAccCheckComputeInstanceConfigs(&instance, testConfig("config", testConfigKernel("linode/latest-64bit"))),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstance_multipleConfigs(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceWithMultipleConfigs(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					// resource.TestCheckResourceAttr(resName, "kernel", "linode/latest-64bit"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "swap_size", "0"),
					testAccCheckComputeInstanceConfigs(&instance, testConfig("configa", testConfigKernel("linode/latest-64bit"))),
					testAccCheckComputeInstanceConfigs(&instance, testConfig("configb", testConfigKernel("linode/latest-32bit"))),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstance_disk(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceWithDisk(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					// resource.TestCheckResourceAttr(resName, "kernel", "linode/latest-64bit"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "swap_size", "0"),
					testAccCheckComputeInstanceDisk(&instance, "disk", 3000),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstance_multipleDisks(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceMultipleDisks(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					// resource.TestCheckResourceAttr(resName, "kernel", "linode/latest-64bit"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "swap_size", "512"),
					testAccCheckComputeInstanceDisk(&instance, "diska", 3000),
					testAccCheckComputeInstanceDisk(&instance, "diskb", 512),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstance_diskAndConfig(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceWithDiskAndConfig(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					// resource.TestCheckResourceAttr(resName, "kernel", "linode/latest-64bit"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "swap_size", "0"),
					testAccCheckComputeInstanceConfigs(&instance, testConfig("config", testConfigKernel("linode/latest-64bit"))),
					testAccCheckComputeInstanceDisk(&instance, "disk", 3000),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstance_disksAndConfigs(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	var instance linodego.Instance
	var instanceDisk linodego.InstanceDisk

	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckLinodeInstanceDestroy,
			testAccCheckLinodeVolumeDestroy,
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceWithMultipleDiskAndConfig(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					// resource.TestCheckResourceAttr(resName, "kernel", "linode/latest-64bit"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "swap_size", "512"),
					testAccCheckLinodeInstanceDiskExists(&instance, "diska", &instanceDisk),
					// TODO(displague) create testAccCheckComputeInstanceDisks helper (like Configs)
					testAccCheckComputeInstanceDisk(&instance, "diska", 3000),
					testAccCheckComputeInstanceDisk(&instance, "diskb", 512),
					testAccCheckComputeInstanceConfigs(&instance, testConfig("configa", testConfigKernel("linode/latest-64bit"), testConfigSDADisk(instanceDisk))),
					testAccCheckComputeInstanceConfigs(&instance, testConfig("configb", testConfigKernel("linode/grub2"), testConfigComments("won't boot"), testConfigSDBDisk(instanceDisk))),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstance_volumeAndConfig(t *testing.T) {
	t.Parallel()

	resName := "linode_instance.foobar"
	volName := "linode_volume.foo"

	var instance linodego.Instance
	var instanceDisk linodego.InstanceDisk
	var volume linodego.Volume
	var instanceName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceWithVolumeAndConfig(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					testAccCheckLinodeVolumeExists(volName, &volume),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "region", "us-east"),
					// resource.TestCheckResourceAttr(resName, "kernel", "linode/latest-64bit"),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
					resource.TestCheckResourceAttr(resName, "boot_config_label", "config"),
					testAccCheckLinodeInstanceDiskExists(&instance, "disk", &instanceDisk),
					// TODO(displague) create testAccCheckComputeInstanceDisks helper (like Configs)
					testAccCheckComputeInstanceDisk(&instance, "disk", 3000),
					testAccCheckComputeInstanceConfigs(&instance, testConfig("config", testConfigKernel("linode/latest-64bit"), testConfigSDADisk(instanceDisk), testConfigSDBVolume(volume))),
				),
			},

			resource.TestStep{
				ResourceName: resName,
				ImportState:  true,
			},
		},
	})
}

func TestAccLinodeInstanceUpdate_simple(t *testing.T) {
	t.Parallel()
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	resName := "linode_instance.foobar"
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Error generating test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceBasic(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
				),
			},
			resource.TestStep{
				Config: testAccCheckLinodeInstanceSimpleUpdates(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", fmt.Sprintf("%s_r", instanceName)),
					resource.TestCheckResourceAttr(resName, "group", "tf_test_r"),
				),
			},
		},
	})
}

func TestAccLinodeInstanceUpdate_config(t *testing.T) {
	t.Parallel()
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	resName := "linode_instance.foobar"
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Error generating test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceWithConfig(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", instanceName),
					resource.TestCheckResourceAttr(resName, "group", "tf_test"),
				),
			},
			resource.TestStep{
				Config: testAccCheckLinodeInstanceConfigSimpleUpdates(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "label", fmt.Sprintf("%s_r", instanceName)),
					resource.TestCheckResourceAttr(resName, "group", "tf_test_r"),
					// changed kerel, not label
					resource.TestCheckResourceAttr(resName, "config.0.label", "config"),
					resource.TestCheckResourceAttr(resName, "config.0.kernel", "linode/latest-32bit"),
				),
			},
		},
	})
}

func TestAccLinodeInstanceResize(t *testing.T) {
	t.Parallel()
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	resName := "linode_instance.foobar"
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Error generating test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			// Start off with a Linode 1024
			resource.TestStep{
				Config: testAccCheckLinodeInstanceConfigUpsizeSmall(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "plan_storage_utilized", "25600"),
					resource.TestCheckResourceAttr(resName, "storage_utilized", "25600"),
					resource.TestCheckResourceAttr(resName, "storage", "25600"),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
				),
			},
			// Bump it to a 2048, but don't expand the disk
			resource.TestStep{
				Config: testAccCheckLinodeInstanceConfigUpsizeBigger(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "type", "g6-standard-1"),
					resource.TestCheckResourceAttr(resName, "plan_storage_utilized", "25600"),
					resource.TestCheckResourceAttr(resName, "storage_utilized", "25600"),
					resource.TestCheckResourceAttr(resName, "storage", "25600"),
				),
			},
			// Go back down to a 1024
			resource.TestStep{
				Config: testAccCheckLinodeInstanceConfigDownsize(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
				),
			},
		},
	})
}

func TestAccLinodeInstanceExpandDisk(t *testing.T) {
	t.Parallel()
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	resName := "linode_instance.foobar"
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Error generating test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			// Start off with a Linode 1024
			resource.TestStep{
				Config: testAccCheckLinodeInstanceConfigUpsizeSmall(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "type", "g6-nanode-1"),
					resource.TestCheckResourceAttr(resName, "plan_storage_utilized", "25600"),
				),
			},
			// Bump it to a 2048, and expand the disk
			resource.TestStep{
				Config: testAccCheckLinodeInstanceConfigUpsizeExpandDisk(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					resource.TestCheckResourceAttr(resName, "type", "g6-standard-1"),
					resource.TestCheckResourceAttr(resName, "plan_storage_utilized", "25600"),
				),
			},
		},
	})
}

func TestAccLinodeInstancePrivateNetworking(t *testing.T) {
	t.Parallel()
	var instance linodego.Instance
	var instanceName = acctest.RandomWithPrefix("tf_test")
	resName := "linode_instance.foobar"
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Error generating test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckLinodeInstanceConfigPrivateNetworking(instanceName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists(resName, &instance),
					testAccCheckLinodeInstanceAttributesPrivateNetworking("linode_instance.foobar"),
					resource.TestCheckResourceAttr(resName, "private_networking", "true"),
				),
			},
		},
	})
}

func testAccCheckLinodeInstanceExists(name string, instance *linodego.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(linodego.Client)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		id, err := strconv.Atoi(rs.Primary.ID)

		found, err := client.GetInstance(context.Background(), id)
		if err != nil {
			return fmt.Errorf("Error retrieving state of Instance %s: %s", rs.Primary.Attributes["label"], err)
		}

		*instance = *found

		return nil
	}
}

func testAccCheckLinodeInstanceDestroy(s *terraform.State) error {
	client, ok := testAccProvider.Meta().(linodego.Client)
	if !ok {
		return fmt.Errorf("Error getting Linode client")
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_instance" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing %v as int", rs.Primary.ID)
		}

		if id == 0 {
			return fmt.Errorf("should not have Linode ID 0")
		}

		_, err = client.GetInstance(context.Background(), id)

		if err == nil {
			return fmt.Errorf("should not find Linode ID %d existing after delete", id)
		}

		if apiErr, ok := err.(*linodego.Error); ok && apiErr.Code != 404 {
			return fmt.Errorf("Error getting Linode ID %d: %s", id, err)
		}
	}

	return nil
}

func testAccCheckLinodeInstanceAttributesPrivateNetworking(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("should have found linode_instance resource %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("should have a Linode ID")
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("should have an integer Linode ID: %s", err)
		}

		client, ok := testAccProvider.Meta().(linodego.Client)
		if !ok {
			return fmt.Errorf("should have a linodego.Client")
		}

		if err != nil {
			return err
		}

		instanceIPs, err := client.GetInstanceIPAddresses(context.Background(), id)
		if err != nil {
			return err
		}
		if len(instanceIPs.IPv4.Private) == 0 {
			return fmt.Errorf("should have a private ip on Linode ID %d", id)
		}
		return nil
	}
}

type testConfigFunc func(config *linodego.InstanceConfig) error
type testConfigsFunc func(config []*linodego.InstanceConfig) error

// testConfig verifies a labeled config exists and runs many tests against that config
func testConfig(label string, configTests ...testConfigFunc) testConfigsFunc {
	return func(configs []*linodego.InstanceConfig) error {
		for _, config := range configs {
			if config.Label == label {
				for _, test := range configTests {
					if err := test(config); err != nil {
						return err
					}
				}
				return nil
			}
		}
		return fmt.Errorf("should have found Instance config with label: %s", label)
	}
}

func testConfigLabel(label string) testConfigFunc {
	return func(config *linodego.InstanceConfig) error {
		if config.Label != label {
			return fmt.Errorf("should have matching labels: %s != %s", config.Label, label)
		}
		return nil
	}
}

func testConfigKernel(kernel string) testConfigFunc {
	return func(config *linodego.InstanceConfig) error {
		if config.Kernel != kernel {
			return fmt.Errorf("should have matching kernels: %s != %s", config.Kernel, kernel)
		}
		return nil
	}
}

func testConfigComments(comments string) testConfigFunc {
	return func(config *linodego.InstanceConfig) error {
		if config.Comments != comments {
			return fmt.Errorf("should have matching comments: %s != %s", config.Comments, comments)
		}
		return nil
	}
}

func testConfigSDADisk(disk linodego.InstanceDisk) testConfigFunc {
	return func(config *linodego.InstanceConfig) error {
		if config.Devices.SDA.DiskID == disk.ID {
			return fmt.Errorf("should have SDA with expected disk id")
		}
		return nil
	}
}

func testConfigSDBDisk(disk linodego.InstanceDisk) testConfigFunc {
	return func(config *linodego.InstanceConfig) error {
		if config.Devices.SDB.DiskID == disk.ID {
			return fmt.Errorf("should have SDB with expected disk id")
		}
		return nil
	}
}

func testConfigSDBVolume(volume linodego.Volume) testConfigFunc {
	return func(config *linodego.InstanceConfig) error {
		if config.Devices.SDB.VolumeID == volume.ID {
			return fmt.Errorf("should have SDB with expected volume id")
		}
		return nil
	}
}

// testAccCheckComputeInstanceConfigs verifies any configs exist and runs config specific tests against a target instance
func testAccCheckComputeInstanceConfigs(instance *linodego.Instance, configsTests ...testConfigsFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(linodego.Client)

		if instance == nil || instance.ID == 0 {
			return fmt.Errorf("Error fetching configs: invalid Instance argument")
		}

		instanceConfigs, err := client.ListInstanceConfigs(context.Background(), instance.ID, nil)

		if err != nil {
			return fmt.Errorf("Error fetching configs: %s", err)
		}

		if len(instanceConfigs) == 0 {
			return fmt.Errorf("No configs")
		}

		for _, tests := range configsTests {
			if err := tests(instanceConfigs); err != nil {
				return err
			}
		}

		return nil
	}
}
func testAccCheckLinodeInstanceDiskExists(instance *linodego.Instance, label string, instanceDisk *linodego.InstanceDisk) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(linodego.Client)

		if instance == nil || instance.ID == 0 {
			return fmt.Errorf("Error fetching disks: invalid Instance argument")
		}

		instanceDisks, err := client.ListInstanceDisks(context.Background(), instance.ID, nil)

		if err != nil {
			return fmt.Errorf("Error fetching disks: %s", err)
		}

		if len(instanceDisks) == 0 {
			return fmt.Errorf("No disks")
		}

		for _, disk := range instanceDisks {
			if disk.Label == label {
				*instanceDisk = *disk
				return nil
			}
		}

		return fmt.Errorf("Disk not found: %s", label)
	}
}

func testAccCheckComputeInstanceDisk(instance *linodego.Instance, label string, size int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(linodego.Client)

		if instance == nil || instance.ID == 0 {
			return fmt.Errorf("Error fetching disks: invalid Instance argument")
		}

		instanceDisks, err := client.ListInstanceDisks(context.Background(), instance.ID, nil)

		if err != nil {
			return fmt.Errorf("Error fetching disks: %s", err)
		}

		if len(instanceDisks) == 0 {
			return fmt.Errorf("No disks")
		}

		for _, disk := range instanceDisks {
			if disk.Label == label && disk.Size == size {
				return nil
			}
		}

		return fmt.Errorf("Disk not found: %s", label)
	}
}

func testAccCheckLinodeInstanceBasic(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	group = "tf_test"
	type = "g6-nanode-1"
	image = "linode/ubuntu18.04"
	region = "us-east"
	root_pass = "terraform-test"
	swap_size = 256
	authorized_keys = "%s"
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceWithConfig(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	group = "tf_test"
	type = "g6-nanode-1"
	region = "us-east"
	config {
		label = "config"
		kernel = "linode/latest-64bit"
	}
}`, instance)
}

func testAccCheckLinodeInstanceWithMultipleConfigs(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	group = "tf_test"
	type = "g6-nanode-1"
	region = "us-east"
	config {
		label = "configa"
		kernel = "linode/latest-64bit"
	}
	config {
		label = "configb"
		kernel = "linode/latest-32bit"
	}
}`, instance)
}

func testAccCheckLinodeInstanceWithDisk(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	group = "tf_test"
	type = "g6-nanode-1"
	region = "us-east"
	disk {
		label = "disk"
		image = "linode/ubuntu18.04"
		root_pass = "b4d_p4s5"
		authorized_keys = "%s"
		size = 3000
	}
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceMultipleDisks(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	group = "tf_test"
	type = "g6-nanode-1"
	region = "us-east"
	disk {
		label = "diska"
		image = "linode/ubuntu18.04"
		root_pass = "b4d_p4s5"
		authorized_keys = "%s"
		size = 3000
	}
	disk {
		label = "diskb"
		filesystem = "swap"
		size = 512
	}
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceWithDiskAndConfig(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	type = "g6-nanode-1"
	region = "us-east"
	group = "tf_test"

	disk {
		label = "disk"
		image = "linode/ubuntu18.04"
		root_pass = "b4d_p4s5"
		authorized_keys = "%s"
		size = 3000
	}

	config {
		label = "config"
		kernel = "linode/latest-64bit"
		devices = { sda = { disk_label = "disk" } }
	}
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceWithMultipleDiskAndConfig(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	type = "g6-nanode-1"
	region = "us-east"
	group = "tf_test"

	disk {
		label = "diska"
		image = "linode/ubuntu18.04"
		root_pass = "b4d_p4s5"
		authorized_keys = "%s"
		size = 3000
	}

	disk {
		label = "diskb"
		filesystem = "swap"
		size = 512
	}

	config {
		label = "configa"
		kernel = "linode/latest-64bit"
		devices = { sda = { disk_label = "diska" }, sdb = { disk_label = "diskb" } }
	}

	config {
		label = "configb"
		comments = "won't boot"
		kernel = "linode/grub2"
		devices = { sda = { disk_label = "diskb" }, sdb = { disk_label = "diska" } }
	}

	boot_config_label = "configa"
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceWithVolumeAndConfig(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_volume" "foo" {
	label = "%s"
	size = "10"
	region = "us-east"
}

resource "linode_instance" "foobar" {
	label = "%s"
	type = "g6-nanode-1"
	region = "us-east"
	group = "tf_test"

	disk {
		label = "disk"
		image = "linode/ubuntu18.04"
		root_pass = "b4d_p4s5"
		authorized_keys = "%s"
		size = 3000
	}

	config {
		label = "config"
		kernel = "linode/latest-64bit"
		devices = {
			sda = { disk_label = "disk" },
			sdb = { volume_id = "${linode_volume.foo.id}" }
		}
	}
}`, instance, instance, pubkey)
}

// testAccCheckLinodeInstanceSimpleUpdates is testAccCheckLinodeInstanceWithConfig with an instance and group rename
func testAccCheckLinodeInstanceSimpleUpdates(instance string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s_r"
	type = "g6-nanode-1"
	region = "us-east"
	group = "tf_test_r"

	config {
		label = "config"
		kernel = "linode/latest-64bit"
	}
}`, instance)
}

// testAccCheckLinodeInstanceConfigSimpleUpdates is testAccCheckLinodeInstanceWithConfig with an instance and group rename and a different kernel
func testAccCheckLinodeInstanceConfigSimpleUpdates(instance string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s_r"
	type = "g6-nanode-1"
	region = "us-east"
	group = "tf_test_r"

	config {
		label = "config"
		kernel = "linode/latest-32bit"
	}
}`, instance)
}

func testAccCheckLinodeInstanceConfigUpsizeSmall(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	type = "g6-nanode-1"
	image = "linode/ubuntu18.04"
	region = "us-east"
	kernel = "linode/latest-64bit"
	root_password = "terraform-test"
	swap_size = 256
	authorized_keys = "%s"
	group = "tf_test"
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceConfigUpsizeBigger(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s_upsized"
	type = "g6-standard-1"
	image = "linode/ubuntu18.04"
	region = "us-east"
	kernel = "linode/latest-64bit"
	root_password = "terraform-test"
	swap_size = 256
	authorized_keys = "%s"
	group = "tf_test"
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceConfigDownsize(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s_downsized"
	type = "g6-nanode-1"
	image = "linode/ubuntu18.04"
	region = "us-east"
	kernel = "linode/latest-64bit"
	root_password = "terraform-test"
	swap_size = 256
	authorized_keys = "%s"
	group = "tf_test"
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceConfigUpsizeExpandDisk(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s_expanded"
	type = "g6-standard-1"
	disk_expansion = true
	image = "linode/ubuntu18.04"
	region = "us-east"
	kernel = "linode/latest-64bit"
	root_password = "terraform-test"
	swap_size = 256
	authorized_keys = "%s"
	group = "tf_test"
}`, instance, pubkey)
}

func testAccCheckLinodeInstanceConfigPrivateNetworking(instance string, pubkey string) string {
	return fmt.Sprintf(`
resource "linode_instance" "foobar" {
	label = "%s"
	type = "g6-nanode-1"
	image = "linode/ubuntu18.04"
	region = "us-east"
	kernel = "linode/latest-64bit"
	root_password = "terraform-test"
	swap_size = 256
	private_networking = true
	authorized_keys = "%s"
	group = "tf_test"
}`, instance, pubkey)
}
