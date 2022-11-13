package config

import (
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type SyncWriter struct {
	M      sync.Mutex
	Writer io.Writer
}

func (w *SyncWriter) Write(b []byte) (n int, err error) {
	w.M.Lock()
	defer w.M.Unlock()
	return w.Writer.Write(b)
}

type Args_struct struct {
	ConfigFile string
	HotUpdate  string
	AutoReload bool
	Help       bool
	TimeOut    int
	SyncWrite  *SyncWriter
}

type IcmpUfw struct {
	// The name of the interface to listen on
	ListenInterface      []string         `yaml:"listen_interface"`
	FireWallProgram      string           `yaml:"firewall_program"`
	Icmp_ufw_rules       []*icmp_ufw_rule `yaml:"icmp_ufw_rules"`
	TimeOut              int              `yaml:"time_out"`
	Webhook_url          string           `yaml:"webhook_url"`
	Webhook_method       string           `yaml:"webhook_method"`
	Webhook_data         string           `yaml:"webhook_data"`
	Webhook_headers      []string         `yaml:"webhook_headers"`
	HotUpdate            string           `yaml:"hot_update"`
	AutoReload           bool             `yaml:"auto_reload"`
	args                 *Args_struct
	stop                 chan bool
	icmp_ufw_rule_caches map[int]*icmp_ufw_rule
}

type icmp_ufw_rule struct {
	Size      int    `yaml:"size"`
	TimeOut   int    `yaml:"time_out"`
	AllowPort string `yaml:"allow_port"`
}

func GetHotUpdate(hotUpdate_url string) (body []byte, err error) {
	resp, err := http.Get(hotUpdate_url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func GetConfig(args *Args_struct) (_icmp_ufw *IcmpUfw, err error) {
	yamlConetent, err := os.ReadFile(args.ConfigFile)
	if err != nil {
		return
	}
	_icmp_ufw = &IcmpUfw{args: args, icmp_ufw_rule_caches: make(map[int]*icmp_ufw_rule), stop: make(chan bool, 100)}
	if _icmp_ufw.GetHotUpdate() != "" {
		go func(_icmp_ufw *IcmpUfw) {
			for {
				yamlConetent, _ := GetHotUpdate(_icmp_ufw.GetHotUpdate())
				err := yaml.Unmarshal(yamlConetent, _icmp_ufw)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("HotUpdate: %s", _icmp_ufw.GetHotUpdate())
				time.Sleep(time.Second * time.Duration(10))
			}
		}(_icmp_ufw)
		return
	}
	err = yaml.Unmarshal(yamlConetent, &_icmp_ufw)
	if _icmp_ufw.GetAutoReload() {
		go func(_icmp_ufw *IcmpUfw) {
			for {
				yamlConetent, _ := os.ReadFile(_icmp_ufw.args.ConfigFile)
				err := yaml.Unmarshal(yamlConetent, _icmp_ufw)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("AutoReload: %s", _icmp_ufw.args.ConfigFile)
				time.Sleep(time.Second * time.Duration(10))
			}
		}(_icmp_ufw)
	}
	if err != nil {
		return
	}
	return
}

func (c *IcmpUfw) GetListenInterface() []string {
	return c.ListenInterface
}

func (c *IcmpUfw) GetIcmpUfwRules() []*icmp_ufw_rule {
	return c.Icmp_ufw_rules
}

func (c *IcmpUfw) GetTimeOut() int {
	if c.args.TimeOut != 0 {
		return c.args.TimeOut
	}
	return c.TimeOut
}

func (c *IcmpUfw) GetWebhookUrl() string {
	return c.Webhook_url
}

func (c *IcmpUfw) GetWebhookMethod() string {
	return c.Webhook_method
}

func (c *IcmpUfw) GetWebhookData() string {
	return c.Webhook_data
}

func (c *IcmpUfw) GetWebhookHeaders() []string {
	return c.Webhook_headers
}

func (c *IcmpUfw) GetHotUpdate() string {
	if c.args.HotUpdate != "" {
		return c.HotUpdate
	}
	return c.HotUpdate
}

func (c *IcmpUfw) GetAutoReload() bool {
	if c.args.AutoReload != false && c.GetHotUpdate() == "" {
		return true
	}
	return c.AutoReload
}

func (c *IcmpUfw) GetArgs() *Args_struct {
	return c.args
}

func (c *IcmpUfw) GetFireWallProgram() string {
	return c.FireWallProgram
}

func (c *icmp_ufw_rule) GetSize() int {
	return c.Size
}

func (c *icmp_ufw_rule) GetTimeOut() int {
	return c.TimeOut
}
func (c *icmp_ufw_rule) GetAllowPort() string {
	return c.AllowPort
}

func (c *IcmpUfw) GetRule(size int) *icmp_ufw_rule {
	if c.icmp_ufw_rule_caches[size] != nil {
		return c.icmp_ufw_rule_caches[size]
	}
	for _, rule := range c.GetIcmpUfwRules() {
		if rule.GetSize() == size {
			c.icmp_ufw_rule_caches[size] = rule
			return rule
		}
	}
	return nil
}

func (c *IcmpUfw) GetStop() chan bool {
	return c.stop
}

func (c *IcmpUfw) SetStop() {
	c.stop <- true
}
