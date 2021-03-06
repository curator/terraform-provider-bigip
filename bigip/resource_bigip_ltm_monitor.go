package bigip

import (
	"fmt"
	"log"

	"strings"

	"github.com/f5devcentral/go-bigip"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceBigipLtmMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceBigipLtmMonitorCreate,
		Read:   resourceBigipLtmMonitorRead,
		Update: resourceBigipLtmMonitorUpdate,
		Delete: resourceBigipLtmMonitorDelete,
		Exists: resourceBigipLtmMonitorExists,
		Importer: &schema.ResourceImporter{
			State: resourceBigipLtmMonitorImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the monitor",
				ForceNew:     true,
				ValidateFunc: validateF5Name,
			},

			"parent": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateParent,
				ForceNew:     true,
				Description:  "Existing monitor to inherit from. Must be one of /Common/http, /Common/https, /Common/icmp or /Common/gateway-icmp.",
			},

			"interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Check interval in seconds",
				Default:     3,
			},

			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Timeout in seconds",
				Default:     16,
			},

			"send": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "GET /\\r\\n",
				Description: "Request string to send.",
				StateFunc: func(s interface{}) string {
					return strings.Replace(s.(string), "\r\n", "\\r\\n", -1)
				},
			},

			"receive": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expected response string.",
			},

			"receive_disable": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Expected response string.",
			},

			"reverse": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"transparent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"manual_resume": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ip_dscp": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"time_until_up": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Time in seconds",
			},

			"destination": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Alias for the destination",
			},
		},
	}
}

func resourceBigipLtmMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)
	name := d.Get("name").(string)

	log.Println("[INFO] Creating monitor " + name + " :: " + monitorParent(d.Get("parent").(string)))

	client.CreateMonitor(
		name,
		monitorParent(d.Get("parent").(string)),
		d.Get("interval").(int),
		d.Get("timeout").(int),
		d.Get("send").(string),
		d.Get("receive").(string),
	)

	d.SetId(name)

	resourceBigipLtmMonitorUpdate(d, meta)
	return resourceBigipLtmMonitorRead(d, meta)
}

func resourceBigipLtmMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	monitors, err := client.Monitors()
	if err != nil {
		return err
	}
	if monitors == nil {
		log.Printf("[WARN] Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	for _, m := range monitors {
		if m.FullPath == name {
			d.Set("interval", m.Interval)
			d.Set("timeout", m.Timeout)
			d.Set("send", m.SendString)
			d.Set("receive", m.ReceiveString)
			d.Set("receive_disable", m.ReceiveDisable)
			d.Set("reverse", m.Reverse)
			d.Set("transparent", m.Transparent)
			d.Set("ip_dscp", m.IPDSCP)
			d.Set("time_until_up", m.TimeUntilUp)
			d.Set("manual_resume", m.ManualResume)
			d.Set("parent", m.ParentMonitor)
			d.Set("destination", m.Destination)
			d.Set("name", name)
			return nil
		}
	}
	log.Println("I am printing whole thing   ", monitors)
	return fmt.Errorf("Couldn't find monitor %s", name)

}

func resourceBigipLtmMonitorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*bigip.BigIP)

	name := d.Id()
	log.Println("[INFO] Fetching monitor " + name)

	monitors, err := client.Monitors()
	if err != nil {
		return false, err
	}

	for _, m := range monitors {
		if m.FullPath == name {
			return true, nil
		}
	}

	return false, nil
}

func resourceBigipLtmMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)

	name := d.Id()

	m := &bigip.Monitor{
		Interval:       d.Get("interval").(int),
		Timeout:        d.Get("timeout").(int),
		SendString:     d.Get("send").(string),
		ReceiveString:  d.Get("receive").(string),
		ReceiveDisable: d.Get("receive_disable").(string),
		Reverse:        d.Get("reverse").(bool),
		Transparent:    d.Get("transparent").(bool),
		IPDSCP:         d.Get("ip_dscp").(int),
		TimeUntilUp:    d.Get("time_until_up").(int),
		ManualResume:   d.Get("manual_resume").(bool),
		Destination:    d.Get("destination").(string),
	}

	err := client.ModifyMonitor(name, monitorParent(d.Get("parent").(string)), m)
	if err != nil {
		return err
	}
	return resourceBigipLtmMonitorRead(d, meta)
}

func resourceBigipLtmMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*bigip.BigIP)
	name := d.Id()
	parent := monitorParent(d.Get("parent").(string))
	log.Println("[Info] Deleting monitor " + name + "::" + parent)
	err := client.DeleteMonitor(name, parent)
	if err != nil {
		return err
	}
	if err == nil {
		log.Printf("[WARN] Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	return nil
}

func validateParent(v interface{}, k string) ([]string, []error) {
	p := v.(string)
	if p == "/Common/http" || p == "/Common/https" || p == "/Common/icmp" || p == "/Common/gateway-icmp" || p == "/Common/tcp" {
		return nil, nil
	}

	return nil, []error{fmt.Errorf("parent must be one of /Common/http, /Common/https, /Common/icmp, /Common/gateway-icmp, or /Common/tcp")}
}

func monitorParent(s string) string {
	return strings.TrimPrefix(s, "/Common/")
}

func resourceBigipLtmMonitorImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
