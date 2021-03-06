package linode

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testFirewallResName = "linode_firewall.test"

func init() {
	resource.AddTestSweepers("linode_firewall", &resource.Sweeper{
		Name: "linode_firewall",
		F:    testSweepLinodeFirewall,
	})
}

func testSweepLinodeFirewall(prefix string) error {
	client, err := getClientForSweepers()
	if err != nil {
		return fmt.Errorf("failed to get client: %s", err)
	}

	firewalls, err := client.ListLKEClusters(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to get firewalls: %s", err)
	}
	for _, firewall := range firewalls {
		if !shouldSweepAcceptanceTestResource(prefix, firewall.Label) {
			continue
		}
		if err := client.DeleteFirewall(context.Background(), firewall.ID); err != nil {
			return fmt.Errorf("failed to destroy firewall %d during sweep: %s", firewall.ID, err)
		}
	}

	return nil
}

func TestAccLinodeFirewall_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf_test")
	devicePrefix := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: accTestWithProvider(testAccCheckLinodeFirewallBasic(name, devicePrefix), map[string]interface{}{
					providerKeySkipInstanceReadyPoll: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallResName, "label", name),
					resource.TestCheckResourceAttr(testFirewallResName, "disabled", "false"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound_policy", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.action", "ACCEPT"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.0", "::/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound_policy", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv6.0", "2001:db8::/32"),
					resource.TestCheckResourceAttr(testFirewallResName, "devices.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "devices.0.type", "linode"),
					resource.TestCheckResourceAttr(testFirewallResName, "linodes.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.0", "test"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.url"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.id"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.entity_id"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.label"),
				),
			},
			{
				ResourceName:      testFirewallResName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLinodeFirewall_minimum(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: accTestWithProvider(testAccCheckLinodeFirewallMinimum(name), map[string]interface{}{
					providerKeySkipInstanceReadyPoll: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallResName, "label", name),
					resource.TestCheckResourceAttr(testFirewallResName, "disabled", "false"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ports", ""),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "devices.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "linodes.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.0", "test"),
				),
			},
			{
				ResourceName:      testFirewallResName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLinodeFirewall_multipleRules(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf_test")
	devicePrefix := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: accTestWithProvider(testAccCheckLinodeFirewallMultipleRules(name, devicePrefix), map[string]interface{}{
					providerKeySkipInstanceReadyPoll: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallResName, "label", name),
					resource.TestCheckResourceAttr(testFirewallResName, "disabled", "false"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound_policy", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.#", "2"),

					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.action", "ACCEPT"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.0", "::/0"),

					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.action", "ACCEPT"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ports", "443"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv6.0", "::/0"),

					resource.TestCheckResourceAttr(testFirewallResName, "outbound_policy", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.#", "2"),

					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv6.0", "2001:db8::/32"),

					resource.TestCheckResourceAttr(testFirewallResName, "outbound.1.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.1.ports", "443"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.1.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.1.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.1.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.1.ipv6.0", "2001:db8::/32"),

					resource.TestCheckResourceAttr(testFirewallResName, "devices.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "devices.0.type", "linode"),
					resource.TestCheckResourceAttr(testFirewallResName, "linodes.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.0", "test"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.url"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.id"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.entity_id"),
					resource.TestCheckResourceAttrSet(testFirewallResName, "devices.0.label"),
				),
			},
			{
				ResourceName:      testFirewallResName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLinodeFirewall_no_device(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeFirewallNoDevice(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallResName, "label", name),
					resource.TestCheckResourceAttr(testFirewallResName, "disabled", "false"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.0", "::/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv6.0", "::/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "devices.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "linodes.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.0", "test"),
				),
			},
			{
				ResourceName:      testFirewallResName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLinodeFirewall_updates(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf_test")
	newName := acctest.RandomWithPrefix("tf_test")
	devicePrefix := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: accTestWithProvider(testAccCheckLinodeFirewallBasic(name, devicePrefix), map[string]interface{}{
					providerKeySkipInstanceReadyPoll: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallResName, "label", name),
					resource.TestCheckResourceAttr(testFirewallResName, "disabled", "false"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound_policy", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.action", "ACCEPT"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.0", "::/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound_policy", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.action", "ACCEPT"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv6.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.0.ipv6.0", "2001:db8::/32"),
					resource.TestCheckResourceAttr(testFirewallResName, "devices.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "devices.0.type", "linode"),
					resource.TestCheckResourceAttr(testFirewallResName, "linodes.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.0", "test"),
				),
			},
			{
				Config: accTestWithProvider(testAccCheckLinodeFirewallUpdates(newName, devicePrefix), map[string]interface{}{
					providerKeySkipInstanceReadyPoll: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testFirewallResName, "label", newName),
					resource.TestCheckResourceAttr(testFirewallResName, "disabled", "true"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound_policy", "ACCEPT"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.#", "3"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.action", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ports", "80"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.#", "2"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.0", "::/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.0.ipv6.1", "ff00::/8"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.action", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ports", "443"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv4.#", "2"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv4.1", "127.0.0.1/32"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.1.ipv6.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.2.action", "DROP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.2.protocol", "TCP"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.2.ports", "22"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.2.ipv4.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.2.ipv4.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(testFirewallResName, "inbound.2.ipv6.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound_policy", "ACCEPT"),
					resource.TestCheckResourceAttr(testFirewallResName, "outbound.#", "0"),
					resource.TestCheckResourceAttr(testFirewallResName, "linodes.#", "1"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.0", "test"),
					resource.TestCheckResourceAttr(testFirewallResName, "tags.1", "test2"),
				),
			},
		},
	})
}

func testAccCheckLinodeFirewallInstance(prefix, identifier string) string {
	return fmt.Sprintf(`
resource "linode_instance" "%[1]s" {
	label = "%.15[2]s-%[1]s"
	group = "tf_test"
	type = "g6-nanode-1"
	region = "ca-central"
	disk {
		label = "disk"
		image = "linode/alpine3.11"
		root_pass = "b4d_p4s5"
		authorized_keys = ["%[3]s"]
		size = 3000
	}
}`, identifier, prefix, publicKeyMaterial)
}

func testAccCheckLinodeFirewallBasic(name, devicePrefix string) string {
	return testAccCheckLinodeFirewallInstance(devicePrefix, "one") + fmt.Sprintf(`
resource "linode_firewall" "test" {
	label = "%s"
	tags  = ["test"]

	inbound {
		label    = "tf-test-in"
		action = "ACCEPT"
		protocol  = "TCP"
		ports     = "80"
		ipv4 = ["0.0.0.0/0"]
		ipv6 = ["::/0"]
	}
	inbound_policy = "DROP"

	outbound {
		label    = "tf-test-out"
		action = "ACCEPT"
		protocol  = "TCP"
		ports     = "80"
		ipv4 = ["0.0.0.0/0"]
		ipv6 = ["2001:db8::/32"]
	}
	outbound_policy = "DROP"

	linodes = [linode_instance.one.id]
}`, name)
}

func testAccCheckLinodeFirewallMinimum(name string) string {
	return fmt.Sprintf(`
resource "linode_firewall" "test" {
	label = "%s"
	tags  = ["test"]

	inbound {
		label    = "tf-test-in"
		action = "ACCEPT"
		protocol = "tcp"
		ipv4 = ["0.0.0.0/0"]
	}
	inbound_policy = "DROP"
	outbound_policy = "DROP"
}`, name)
}

func testAccCheckLinodeFirewallMultipleRules(name, devicePrefix string) string {
	return testAccCheckLinodeFirewallInstance(devicePrefix, "one") + fmt.Sprintf(`
resource "linode_firewall" "test" {
	label = "%s"
	tags  = ["test"]

	inbound {
		label    = "tf-test-in"
		action = "ACCEPT"
		protocol  = "TCP"
		ports     = "80"
		ipv4 = ["0.0.0.0/0"]
		ipv6 = ["::/0"]
	}

	inbound {
		label    = "tf-test-in-1"
		action = "ACCEPT"
		protocol  = "TCP"
		ports     = "443"
		ipv4 = ["0.0.0.0/0"]
		ipv6 = ["::/0"]
	}
	inbound_policy = "DROP"

	outbound {
		label    = "tf-test-out"
		action = "ACCEPT"
		protocol  = "TCP"
		ports     = "80"
		ipv4 = ["0.0.0.0/0"]
		ipv6 = ["2001:db8::/32"]
	}

	outbound {
		label    = "tf-test-out-1"
		action = "ACCEPT"
		protocol  = "TCP"
		ports     = "443"
		ipv4 = ["0.0.0.0/0"]
		ipv6 = ["2001:db8::/32"]
	}
	outbound_policy = "DROP"

	linodes = [linode_instance.one.id]
}`, name)
}

func testAccCheckLinodeFirewallNoDevice(name string) string {
	return fmt.Sprintf(`
resource "linode_firewall" "test" {
	label = "%s"
	tags  = ["test"]

	inbound {
		label    = "tf-test-in"
		action   = "ACCEPT"
		protocol = "TCP"
		ports    = "80"
		ipv6     = ["::/0"]
	}

	inbound_policy = "DROP"
	outbound {
		label    = "tf-test-out"
		action   = "ACCEPT"
		protocol = "TCP"
		ports    = "80"
		ipv6     = ["::/0"]
	}
	outbound_policy = "DROP"

	linodes = []
}`, name)
}

func testAccCheckLinodeFirewallUpdates(name, devicePrefix string) string {
	return testAccCheckLinodeFirewallInstance(devicePrefix, "one") +
		testAccCheckLinodeFirewallInstance(devicePrefix, "two") +
		fmt.Sprintf(`
resource "linode_firewall" "test" {
	label    = "%s"
	tags     = ["test", "test2"]
    disabled = true

	inbound {
		label    = "tf-test-in"
		action   = "DROP"
		protocol = "TCP"
		ports    = "80"
		ipv4     = ["0.0.0.0/0"]
		ipv6     = ["::/0", "ff00::/8"]
	}

	inbound {
		label    = "tf-test-in2"
		action   = "DROP"
		protocol = "TCP"
		ports    = "443"
		ipv4     = ["0.0.0.0/0", "127.0.0.1/32"]
	}

	inbound {
		label    = "tf-test-in3"
		action   = "DROP"
		protocol = "TCP"
		ports    = "22"
		ipv4     = ["0.0.0.0/0"]
	}
	inbound_policy = "ACCEPT"
	outbound_policy = "ACCEPT"

	linodes = [
		linode_instance.two.id,
	]
}`, name)
}
