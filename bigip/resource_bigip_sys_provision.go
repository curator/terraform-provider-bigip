package bigip

import (
	"github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"fmt"
)

func resourceBigipSysProvision() *schema.Resource {
	return &schema.Resource{
		Create: resourceBigipSysProvisionCreate,
		Update: resourceBigipSysProvisionUpdate,
		Read:   resourceBigipSysProvisionRead,
		Delete: resourceBigipSysProvisionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceBigipSysProvisionImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the module to be provisioned",
				//ValidateFunc: validateF5Name,
			},

			"full_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "path",
			},

			"cpu_ratio": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "cpu Ratio",
			},

			"disk_ratio": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "disk Ratio",
			},

			"level": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "what level nominal or dedicated",
			},

			"memory_ratio": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "memory Ratio",
			},
		},
	}

}

func resourceBigipSysProvisionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Get("name").(string)
	FullPath := d.Get("full_path").(string)
	cpuRatio := d.Get("cpu_ratio").(int)
	diskRatio := d.Get("disk_ratio").(int)
	level := d.Get("level").(string)
	memoryRatio := d.Get("memory_ratio").(int)

	log.Println("[INFO] Provisioning  ")

	err := client.CreateProvision(
		name,
		FullPath,
		cpuRatio,
		diskRatio,
		level,
		memoryRatio,
	)

	if err != nil {
		return err
	}
	d.SetId(name)
	return resourceBigipSysProvisionRead(d, meta)
}

func resourceBigipSysProvisionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	log.Println("[INFO] Updating Provsioning " + name)

	r := &bigip.Provision{
		Name:        name,
		FullPath:    d.Get("full_path").(string),
		CpuRatio:    d.Get("cpu_ratio").(int),
		DiskRatio:   d.Get("disk_ratio").(int),
		Level:       d.Get("level").(string),
		MemoryRatio: d.Get("memory_ratio").(int),
	}

	err := client.ModifyProvision(r)
	if err != nil {
		return err
	}
	return resourceBigipSysProvisionRead(d, meta)
}

func resourceBigipSysProvisionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	log.Println("[INFO] Reading Provisions " + name)

	p, err := client.Provisions(name)
	if err != nil {
		return err
	}
	if p == nil {
		log.Printf("[WARN] Provision (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	d.Set("full_path", p.FullPath)
	if err := d.Set("full_path", p.FullPath); err != nil {
		return fmt.Errorf("[DEBUG] Error saving FullPath to state for Provision  (%s): %s", d.Id(), err)
	}

	p.Name = name

	return nil
}

func resourceBigipSysProvisionDelete(d *schema.ResourceData, meta interface{}) error {
	//API is not supported for Deleting
	return nil
}

func resourceBigipSysProvisionImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
