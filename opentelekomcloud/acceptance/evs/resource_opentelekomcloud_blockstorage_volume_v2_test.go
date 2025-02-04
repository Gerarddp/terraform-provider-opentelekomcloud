package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/blockstorage/v2/volumes"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/evs/v2/tags"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common/quotas"
	ecs "github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/ecs"

	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/common"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/acceptance/env"
	"github.com/opentelekomcloud/terraform-provider-opentelekomcloud/opentelekomcloud/common/cfg"
)

const resourceVolumeV2Name = "opentelekomcloud_blockstorage_volume_v2.volume_1"

func TestAccBlockStorageV2Volume_basic(t *testing.T) {
	var volume volumes.Volume
	t.Parallel()
	quotas.BookMany(t, volumeQuotas(1))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumeBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageV2VolumeExists(resourceVolumeV2Name, &volume),
					testAccCheckBlockStorageV2VolumeMetadata(&volume, "foo", "bar"),
					resource.TestCheckResourceAttr(resourceVolumeV2Name, "name", "volume_1"),
				),
			},
			{
				Config: testAccBlockStorageV2VolumeUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageV2VolumeExists(resourceVolumeV2Name, &volume),
					testAccCheckBlockStorageV2VolumeMetadata(&volume, "foo", "bar"),
					resource.TestCheckResourceAttr(resourceVolumeV2Name, "name", "volume_1-updated"),
				),
			},
		},
	})
}

func TestAccBlockStorageV2Volume_importBasic(t *testing.T) {
	t.Parallel()
	quotas.BookMany(t, volumeQuotas(1))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumeBasic,
			},
			{
				ResourceName:      resourceVolumeV2Name,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cascade",
				},
			},
		},
	})
}

func TestAccBlockStorageV2Volume_upscaleDownScale(t *testing.T) {
	var volume volumes.Volume
	t.Parallel()
	quotas.BookMany(t, volumeQuotas(2))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumeBasic,
				Check:  testAccCheckBlockStorageV2VolumeExists(resourceVolumeV2Name, &volume),
			},
			{
				Config: testAccBlockStorageV2VolumeBigger,
				Check:  testAccCheckBlockStorageV2VolumeSame(resourceVolumeV2Name, &volume),
			},
			{
				Config: testAccBlockStorageV2VolumeBasic,
				Check:  testAccCheckBlockStorageV2VolumeNew(resourceVolumeV2Name, &volume),
			},
		},
	})
}
func TestAccBlockStorageV2Volume_upscaleDownScaleAssigned(t *testing.T) {
	var volume volumes.Volume
	t.Parallel()
	qts := volumeQuotas(12)
	qts = append(qts, ecs.QuotasForFlavor(env.OsFlavorID)...)
	qts = append(qts, &quotas.ExpectedQuota{Q: quotas.Server, Count: 1})
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumeAssigned(10),
				Check:  testAccCheckBlockStorageV2VolumeExists(resourceVolumeV2Name, &volume),
			},
			{
				Config: testAccBlockStorageV2VolumeAssigned(12),
				Check:  testAccCheckBlockStorageV2VolumeSame(resourceVolumeV2Name, &volume),
			},
			{
				Config: testAccBlockStorageV2VolumeAssigned(10),
				Check:  testAccCheckBlockStorageV2VolumeNew(resourceVolumeV2Name, &volume),
			},
		},
	})
}

func TestAccBlockStorageV2Volume_policy(t *testing.T) {
	t.Parallel()
	quotas.BookMany(t, volumeQuotas(40))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			common.TestAccPreCheck(t)
			testPolicyPreCheck(t)
		},
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumePolicy,
			},
		},
	})
}

func testPolicyPreCheck(t *testing.T) {
	if os.Getenv("OS_KMS_NAME") == "" {
		t.Skipf("OS_KMS_NAME should be set for this test to existing KMS key alias")
	}
}

func TestAccBlockStorageV2Volume_tags(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageV2VolumeTags(resourceVolumeV2Name, "foo", "bar"),
					testAccCheckBlockStorageV2VolumeTags(resourceVolumeV2Name, "key", "value"),
				),
			},
			{
				Config: testAccBlockStorageV2VolumeTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageV2VolumeTags(resourceVolumeV2Name, "foo2", "bar2"),
					testAccCheckBlockStorageV2VolumeTags(resourceVolumeV2Name, "key2", "value2"),
				),
			},
		},
	})
}

func TestAccBlockStorageV2Volume_image(t *testing.T) {
	var volume volumes.Volume
	t.Parallel()
	qts := []*quotas.ExpectedQuota{{Q: quotas.Volume, Count: 1}, {Q: quotas.VolumeSize, Count: 12}}
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumeImage,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageV2VolumeExists(resourceVolumeV2Name, &volume),
					resource.TestCheckResourceAttr(
						resourceVolumeV2Name, "name", "volume_1"),
				),
			},
		},
	})
}

func TestAccBlockStorageV2Volume_timeout(t *testing.T) {
	var volume volumes.Volume
	t.Parallel()
	qts := []*quotas.ExpectedQuota{{Q: quotas.Volume, Count: 1}, {Q: quotas.VolumeSize, Count: 1}}
	quotas.BookMany(t, qts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { common.TestAccPreCheck(t) },
		ProviderFactories: common.TestAccProviderFactories,
		CheckDestroy:      testAccCheckBlockStorageV2VolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageV2VolumeTimeout,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageV2VolumeExists(resourceVolumeV2Name, &volume),
				),
			},
		},
	})
}

func testAccCheckBlockStorageV2VolumeDestroy(s *terraform.State) error {
	config := common.TestAccProvider.Meta().(*cfg.Config)
	blockStorageClient, err := config.BlockStorageV2Client(env.OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("error creating OpenTelekomCloud block storage client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_blockstorage_volume_v2" {
			continue
		}

		_, err := volumes.Get(blockStorageClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("volume still exists")
		}
	}

	return nil
}

func testAccCheckBlockStorageV2VolumeExists(n string, volume *volumes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		blockStorageClient, err := config.BlockStorageV2Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud block storage client: %s", err)
		}

		found, err := volumes.Get(blockStorageClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("volume not found")
		}

		*volume = *found

		return nil
	}
}

func testAccCheckBlockStorageV2VolumeMetadata(
	volume *volumes.Volume, k string, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if volume.Metadata == nil {
			return fmt.Errorf("no metadata")
		}

		for key, value := range volume.Metadata {
			if k != key {
				continue
			}

			if v == value {
				return nil
			}

			return fmt.Errorf("bad value for %s: %s", k, value)
		}

		return fmt.Errorf("metadata not found: %s", k)
	}
}

func testAccCheckBlockStorageV2VolumeTags(n string, k string, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := common.TestAccProvider.Meta().(*cfg.Config)
		blockStorageClient, err := config.BlockStorageV2Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud block storage client: %s", err)
		}

		found, err := volumes.Get(blockStorageClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("volume not found")
		}

		client, err := config.BlockStorageV2Client(env.OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("error creating OpenTelekomCloud block storage client: %s", err)
		}
		taglist, err := tags.Get(client, "volumes", found.ID).Extract()
		if err != nil {
			return fmt.Errorf("error creating tags for the volume: %w", err)
		}
		for key, value := range taglist.Tags {
			if k != key {
				continue
			}

			if v == value {
				return nil
			}

			return fmt.Errorf("bad value for %s: %s", k, value)
		}

		return fmt.Errorf("tag not found: %s", k)
	}
}

func testAccCheckBlockStorageV2VolumeSame(n string, volume *volumes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != volume.ID {
			return fmt.Errorf("volume ID changed")
		}
		return nil
	}
}

func testAccCheckBlockStorageV2VolumeNew(n string, volume *volumes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := testAccCheckBlockStorageV2VolumeSame(n, volume)(s)
		if err == nil {
			return fmt.Errorf("volume ID not changed")
		}
		return nil
	}
}

func testAccBlockStorageV2VolumeAssigned(size int) string {
	return fmt.Sprintf(`
%s

%s

resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name     = "volume_1"
  size     = %d
  image_id = data.opentelekomcloud_images_image_v2.latest_image.id
}

resource "opentelekomcloud_compute_instance_v2" "instance_1" {
  name            = "instance_1"
  security_groups = ["default"]
  network {
    uuid = data.opentelekomcloud_vpc_subnet_v1.shared_subnet.network_id
  }
  block_device {
    uuid                  = opentelekomcloud_blockstorage_volume_v2.volume_1.id
    source_type           = "volume"
    boot_index            = 0
    destination_type      = "volume"
    delete_on_termination = true
  }
}
`, common.DataSourceImage, common.DataSourceSubnet, size)
}

const (
	testAccBlockStorageV2VolumeUpdate = `
resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name        = "volume_1-updated"
  description = "first test volume"
  metadata = {
    foo = "bar"
  }
  size = 1
}
`
	testAccBlockStorageV2VolumeTags = `
resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name        = "volume_1"
  description = "first test volume"
  metadata = {
    foo = "bar"
  }
  tags = {
    foo = "bar"
	key = "value"
  }
  size = 1
}
`

	testAccBlockStorageV2VolumeTagsUpdate = `
resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name        = "volume_1-updated"
  description = "first test volume"
  metadata = {
    foo = "bar"
  }
  tags = {
    foo2 = "bar2"
	key2 = "value2"
  }
  size = 1
}
`
	testAccBlockStorageV2VolumeTimeout = `
resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name        = "volume_1"
  description = "first test volume"
  size = 1
  device_type = "SCSI"

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`
	testAccBlockStorageV2VolumeBasic = `
resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name        = "volume_1"
  description = "first test volume"
  metadata = {
    foo = "bar"
  }
  size = 1
}
`

	testAccBlockStorageV2VolumeBigger = `
resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name        = "volume_1"
  description = "first test volume"
  metadata = {
    foo = "bar"
  }
  size = 2
}
`
)

var testAccBlockStorageV2VolumeImage = fmt.Sprintf(`
%s
resource "opentelekomcloud_blockstorage_volume_v2" "volume_1" {
  name     = "volume_1"
  size     = 12
  image_id = data.opentelekomcloud_images_image_v2.latest_image.id
}
`, common.DataSourceImage)

var testAccBlockStorageV2VolumePolicy = fmt.Sprintf(`
%s

data opentelekomcloud_compute_availability_zones_v2 available {}

locals {
  count = 1
}

resource "opentelekomcloud_blockstorage_volume_v2" "volume" {
  count             = local.count
  availability_zone = data.opentelekomcloud_compute_availability_zones_v2.available.names[count.index]
  name              = "test-vol0${count.index + 1}-datadisk"
  size              = 40
  tags = {
    generator = "terraform"
  }
  metadata = {
    __system__encrypted = "1"
    __system__cmkid     = data.opentelekomcloud_kms_key_v1.default_key.id
    attached_mode       = "rw"
    readonly            = "False"
  }
}

resource "opentelekomcloud_vbs_backup_policy_v2" "vbs_policy1" {
  name   = "policy_001"
  status = "ON"

  start_time          = "12:00"
  retain_first_backup = "N"
  rentention_num      = 7
  frequency           = 1

  resources = opentelekomcloud_blockstorage_volume_v2.volume[*].id

}
`, common.DataSourceKMSKey)
