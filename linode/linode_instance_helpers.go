package linode

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/linode/linodego"
	"golang.org/x/crypto/sha3"
)

var (
	boolFalse = false
	boolTrue  = true
)

func flattenInstanceSpecs(instance linodego.Instance) []map[string]int {
	return []map[string]int{{
		"vcpus":    instance.Specs.VCPUs,
		"disk":     instance.Specs.Disk,
		"memory":   instance.Specs.Memory,
		"transfer": instance.Specs.Transfer,
	}}
}

func flattenInstanceAlerts(instance linodego.Instance) []map[string]int {
	return []map[string]int{{
		"cpu":            instance.Alerts.CPU,
		"io":             instance.Alerts.IO,
		"network_in":     instance.Alerts.NetworkIn,
		"network_out":    instance.Alerts.NetworkOut,
		"transfer_quota": instance.Alerts.TransferQuota,
	}}
}

func flattenInstanceDisks(instanceDisks []*linodego.InstanceDisk) (disks []map[string]interface{}, swapSize int) {
	for _, disk := range instanceDisks {
		// Determine if swap exists and the size.  If it does not exist, swap_size=0
		if disk.Filesystem == "swap" {
			swapSize += disk.Size
		}
		disks = append(disks, map[string]interface{}{
			"size":       disk.Size,
			"label":      disk.Label,
			"filesystem": string(disk.Filesystem),
			// TODO(displague) these can not be retrieved after the initial send
			// "read_only":       disk.ReadOnly,
			// "image":           disk.Image,
			// "authorized_keys": disk.AuthorizedKeys,
			// "stackscript_id":  disk.StackScriptID,
		})
	}
	return
}

func flattenInstanceConfigs(instanceConfigs []*linodego.InstanceConfig) (configs []map[string]interface{}) {
	for _, config := range instanceConfigs {

		devices := []map[string]interface{}{{
			"sda": flattenInstanceConfigDevice(config.Devices.SDA),
			"sdb": flattenInstanceConfigDevice(config.Devices.SDB),
			"sdc": flattenInstanceConfigDevice(config.Devices.SDC),
			"sdd": flattenInstanceConfigDevice(config.Devices.SDD),
			"sde": flattenInstanceConfigDevice(config.Devices.SDE),
			"sdf": flattenInstanceConfigDevice(config.Devices.SDF),
			"sdg": flattenInstanceConfigDevice(config.Devices.SDG),
			"sdh": flattenInstanceConfigDevice(config.Devices.SDH),
		}}

		// Determine if swap exists and the size.  If it does not exist, swap_size=0
		configs = append(configs, map[string]interface{}{
			"kernel":       config.Kernel,
			"run_level":    string(config.RunLevel),
			"virt_mode":    string(config.VirtMode),
			"root_device":  config.RootDevice,
			"comments":     config.Comments,
			"memory_limit": config.MemoryLimit,
			"label":        config.Label,
			"helpers": []map[string]bool{{
				"updatedb_disabled":  config.Helpers.UpdateDBDisabled,
				"distro":             config.Helpers.Distro,
				"modules_dep":        config.Helpers.ModulesDep,
				"network":            config.Helpers.Network,
				"devtmpfs_automount": config.Helpers.DevTmpFsAutomount,
			}},
			// panic: interface conversion: interface {} is map[string]map[string]int, not *schema.Set
			"devices": devices,

			// TODO(displague) these can not be retrieved after the initial send
			// "read_only":       disk.ReadOnly,
			// "image":           disk.Image,
			// "authorized_keys": disk.AuthorizedKeys,
			// "stackscript_id":  disk.StackScriptID,
		})
	}
	return
}

func flattenInstanceConfigDevice(dev *linodego.InstanceConfigDevice) []map[string]interface{} {
	if dev == nil {
		return []map[string]interface{}{{
			"disk_id":   0,
			"volume_id": 0,
		}}
	}

	return []map[string]interface{}{{
		"disk_id":   dev.DiskID,
		"volume_id": dev.VolumeID,
	}}
}

// TODO(displague) do we need a disk_label map?
func expandInstanceConfigDeviceMap(m map[string]interface{}, diskIDLabelMap map[string]int) (deviceMap *linodego.InstanceConfigDeviceMap, err error) {
	if len(m) > 0 {
		return nil, nil
	}
	for k, rdev := range m {
		devSlots := rdev.([]interface{})
		for _, rrdev := range devSlots {
			dev := rrdev.(map[string]interface{})
			if k == "sda" {
				deviceMap.SDA = &linodego.InstanceConfigDevice{}
				if err := assignConfigDevice(deviceMap.SDA, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
			if k == "sdb" {
				deviceMap.SDB = &linodego.InstanceConfigDevice{}
				if err := assignConfigDevice(deviceMap.SDB, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
			if k == "sdc" {
				deviceMap.SDC = &linodego.InstanceConfigDevice{}
				if err := assignConfigDevice(deviceMap.SDC, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
			if k == "sdd" {
				deviceMap.SDD = &linodego.InstanceConfigDevice{}
				if err := assignConfigDevice(deviceMap.SDD, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
			if k == "sde" {
				deviceMap.SDE = &linodego.InstanceConfigDevice{}

				if err := assignConfigDevice(deviceMap.SDE, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
			if k == "sdf" {
				deviceMap.SDF = &linodego.InstanceConfigDevice{}

				if err := assignConfigDevice(deviceMap.SDF, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
			if k == "sdg" {
				deviceMap.SDG = &linodego.InstanceConfigDevice{}
				if err := assignConfigDevice(deviceMap.SDG, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
			if k == "sdh" {
				deviceMap.SDH = &linodego.InstanceConfigDevice{}
				if err := assignConfigDevice(deviceMap.SDH, dev, diskIDLabelMap); err != nil {
					return nil, err
				}
			}
		}
	}
	return deviceMap, nil
}

func expandInstanceConfigDevice(m map[string]interface{}) *linodego.InstanceConfigDevice {
	var dev *linodego.InstanceConfigDevice
	// be careful of `disk_label string` in m
	if diskID, ok := m["disk_id"]; ok && diskID.(int) > 0 {
		dev = &linodego.InstanceConfigDevice{
			DiskID: diskID.(int),
		}
	} else if volumeID, ok := m["volume_id"]; ok && volumeID.(int) > 0 {
		dev = &linodego.InstanceConfigDevice{
			VolumeID: m["volume_id"].(int),
		}
	}

	return dev
}

func createDiskFromSet(client linodego.Client, instance linodego.Instance, v interface{}, d *schema.ResourceData) (*linodego.InstanceDisk, error) {
	disk, ok := v.(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("Error converting disk from Terraform Set to golang map")
	}

	diskOpts := linodego.InstanceDiskCreateOptions{
		Label:      disk["label"].(string),
		Filesystem: disk["filesystem"].(string),
		Size:       disk["size"].(int),
	}

	if image, ok := disk["image"]; ok {
		diskOpts.Image = image.(string)

		if rootPass, ok := disk["root_pass"]; ok {
			diskOpts.RootPass = rootPass.(string)
		}

		if authorizedKeys, ok := disk["authorized_keys"]; ok {
			for _, sshKey := range authorizedKeys.([]interface{}) {
				diskOpts.AuthorizedKeys = append(diskOpts.AuthorizedKeys, sshKey.(string))
			}
		}

		if stackscriptID, ok := disk["stackscript_id"]; ok {
			diskOpts.StackscriptID = stackscriptID.(int)
		}

		if stackscriptData, ok := disk["stackscript_data"]; ok {
			for name, value := range stackscriptData.(map[string]interface{}) {
				diskOpts.StackscriptData[name] = value.(string)
			}
		}

		/*
			if sshKeys, ok := d.GetOk("authorized_keys"); ok {
				if sshKeysArr, ok := sshKeys.([]interface{}); ok {
					diskOpts.AuthorizedKeys = make([]string, len(sshKeysArr))
					for k, v := range sshKeys.([]interface{}) {
						if val, ok := v.(string); ok {
							diskOpts.AuthorizedKeys[k] = val
						}
					}
				}
			}
		*/
	}

	instanceDisk, err := client.CreateInstanceDisk(context.Background(), instance.ID, diskOpts)

	if err != nil {
		return nil, fmt.Errorf("Error creating Linode instance %d disk: %s", instance.ID, err)
	}

	_, err = client.WaitForEventFinished(context.Background(), instance.ID, linodego.EntityLinode, linodego.ActionDiskCreate, instanceDisk.Created, int(d.Timeout(schema.TimeoutCreate).Seconds()))
	if err != nil {
		return nil, fmt.Errorf("Error waiting for Linode instance %d disk: %s", instanceDisk.ID, err)
	}

	return instanceDisk, err
}

// getTotalDiskSize returns the number of disks and their total size.
func getTotalDiskSize(client *linodego.Client, linodeID int) (totalDiskSize int, err error) {
	disks, err := client.ListInstanceDisks(context.Background(), linodeID, nil)
	if err != nil {
		return 0, err
	}

	for _, disk := range disks {
		totalDiskSize += disk.Size
	}

	return totalDiskSize, nil
}

// getBiggestDisk returns the ID and Size of the largest disk attached to the Linode
func getBiggestDisk(client *linodego.Client, linodeID int) (biggestDiskID int, biggestDiskSize int, err error) {
	diskFilter := "{\"+order_by\": \"size\", \"+order\": \"desc\"}"
	disks, err := client.ListInstanceDisks(context.Background(), linodeID, linodego.NewListOptions(1, diskFilter))
	if err != nil {
		return 0, 0, err
	}

	for _, disk := range disks {
		// Find Biggest Disk ID & Size
		if disk.Size > biggestDiskSize {
			biggestDiskID = disk.ID
			biggestDiskSize = disk.Size
		}
	}
	return biggestDiskID, biggestDiskSize, nil
}

// sshKeyState hashes a string passed in as an interface
func sshKeyState(val interface{}) string {
	return hashString(strings.Join(val.([]string), "\n"))
}

// rootPasswordState hashes a string passed in as an interface
func rootPasswordState(val interface{}) string {
	return hashString(val.(string))
}

// hashString hashes a string
func hashString(key string) string {
	hash := sha3.Sum512([]byte(key))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// changeInstanceType resizes the Linode Instance
func changeInstanceType(client *linodego.Client, instance *linodego.Instance, targetType string, d *schema.ResourceData) error {
	if err := client.ResizeInstance(context.Background(), instance.ID, targetType); err != nil {
		return fmt.Errorf("Error resizing instance %d: %s", instance.ID, err)
	}

	_, err := client.WaitForEventFinished(context.Background(), instance.ID, linodego.EntityLinode, linodego.ActionLinodeResize, *instance.Created, int(d.Timeout(schema.TimeoutUpdate).Seconds()))
	if err != nil {
		return fmt.Errorf("Error waiting for instance %d to finish resizing: %s", instance.ID, err)
	}

	return nil
}

func changeInstanceDiskSize(client *linodego.Client, instance *linodego.Instance, disk *linodego.InstanceDisk, targetSize int, d *schema.ResourceData) error {
	if instance.Specs.Disk > targetSize {
		client.ResizeInstanceDisk(context.Background(), instance.ID, disk.ID, targetSize)

		// Wait for the Disk Resize Operation to Complete
		// waitForEventComplete(client, instance.ID, "linode_resize", waitMinutes)
		_, err := client.WaitForEventFinished(context.Background(), instance.ID, linodego.EntityLinode, linodego.ActionDiskResize, disk.Updated, int(d.Timeout(schema.TimeoutUpdate).Seconds()))
		if err != nil {
			return fmt.Errorf("Error waiting for resize of Instance %d Disk %d: %s", instance.ID, disk.ID, err)
		}
	} else {
		return fmt.Errorf("Error resizing Disk %d: size exceeds disk size for Instance %d", disk.ID, instance.ID)
	}
	return nil
}

// privateIP determines if an IP is for private use (RFC1918)
// https://stackoverflow.com/a/41273687
func privateIP(ip net.IP) bool {
	private := false
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
	private = private24BitBlock.Contains(ip) || private20BitBlock.Contains(ip) || private16BitBlock.Contains(ip)
	return private
}

func labelHashcode(v interface{}) int {
	switch t := v.(type) {
	case linodego.InstanceConfig:
		return schema.HashString(t.Label)
	case linodego.InstanceDisk:
		return schema.HashString(t.Label)
	case map[string]interface{}:
		if label, ok := t["label"]; ok {
			return schema.HashString(label.(string))
		}
		panic(fmt.Sprintf("Error hashing label for unknown map: %#v", v))
	default:
		panic(fmt.Sprintf("Error hashing label for unknown interface: %#v", v))
	}
}

func assignConfigDevice(device *linodego.InstanceConfigDevice, dev map[string]interface{}, diskIDLabelMap map[string]int) error {
	if label, ok := dev["disk_label"].(string); ok && len(label) > 0 {
		if dev["disk_id"], ok = diskIDLabelMap[label]; !ok {
			return fmt.Errorf("Error mapping disk label %s to ID", dev["disk_label"])
		}
	}
	expanded := expandInstanceConfigDevice(dev)
	*device = *expanded
	return nil
}
