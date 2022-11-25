package config

import (
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Args_struct struct {
	ConfigFile string
	HotUpdate  string
	AutoReload bool
	Help       bool
	TimeOut    int
}

// IcmpUfw
// @Description: icmp_ufw
type IcmpUfw struct {
	// The name of the interface to listen on
	ListenInterface      []string       `yaml:"listen_interface"`
	FireWallProgram      string         `yaml:"firewall_program"`
	Icmp_ufw_rules       []*IcmpUfwRule `yaml:"rules"`
	Firewall_rule_name   string         `yaml:"firewall_rule_name"`
	TimeOut              int            `yaml:"time_out"`
	Webhook_url          string         `yaml:"webhook_url"`
	Webhook_method       string         `yaml:"webhook_method"`
	Webhook_data         string         `yaml:"webhook_data"`
	Webhook_headers      []string       `yaml:"webhook_headers"`
	HotUpdate            string         `yaml:"hot_update"`
	AutoReload           bool           `yaml:"auto_reload"`
	AutoRelaodDelay      int            `yaml:"auto_reload_delay"`
	Open_ports           string         `yaml:"open_ports"`
	args                 *Args_struct
	stop                 chan bool
	icmp_ufw_rule_caches map[int]*IcmpUfwRule
}

type IcmpUfwRule struct {
	Size       int    `yaml:"size"`
	TimeOut    int    `yaml:"time_out"`
	AllowPorts string `yaml:"allow_ports"`
	Pattern    string `yaml:"pattern"`
}

// GetHotUpdate
//
//	@Description: GetHotUpdate
//	@param hotUpdate_url string 热更新地址
//	@return body []byte 热更新内容
//	@return err error 错误
func GetHotUpdate(hotUpdate_url string) (body []byte, err error) {
	resp, err := http.Get(hotUpdate_url)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("GetHotUpdate: %s", err)
		}
	}(resp.Body)
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

// GetConfig
//
//	@Description: GetConfig
//	@param args *Args_struct 命令行参数
//	@return _icmp_ufw *IcmpUfw 配置
//	@return err error 错误
func GetConfig(args *Args_struct) (_icmp_ufw *IcmpUfw, err error) {
	yamlConetent, err := os.ReadFile(args.ConfigFile)
	if err != nil {
		return
	}
	_icmp_ufw = &IcmpUfw{args: args, icmp_ufw_rule_caches: make(map[int]*IcmpUfwRule), stop: make(chan bool, 100)}
	if _icmp_ufw.GetHotUpdate() != "" {
		// 热更新协程
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
	// 自动重载协程
	if _icmp_ufw.GetAutoReload() {
		go func(_icmp_ufw *IcmpUfw) {
			for {
				yamlConetent, _ := os.ReadFile(_icmp_ufw.args.ConfigFile)
				err := yaml.Unmarshal(yamlConetent, _icmp_ufw)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("AutoReload: %s", _icmp_ufw.args.ConfigFile)
				time.Sleep(time.Second * time.Duration(_icmp_ufw.GetAutoReloadDelay()))
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

func (c *IcmpUfw) GetIcmpUfwRules() []*IcmpUfwRule {
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

func (c *IcmpUfw) GetAutoReloadDelay() int {
	if c.AutoRelaodDelay != 0 {
		return c.AutoRelaodDelay
	}
	return 10
}

func (c *IcmpUfw) GetArgs() *Args_struct {
	return c.args
}

func (c *IcmpUfw) GetFireWallProgram() string {
	return c.FireWallProgram
}

func (c *IcmpUfwRule) GetSize() int {
	return c.Size
}

func (c *IcmpUfwRule) GetPattern() byte {
	if c.Pattern == "" {
		return 32
	}
	parseInt, _ := strconv.ParseInt(c.Pattern, 0, 0)
	return byte(parseInt)
}

func (c *IcmpUfwRule) GetTimeOut() int {
	return c.TimeOut
}
func (c *IcmpUfwRule) GetAllowPorts() string {
	return c.AllowPorts
}

func (icmp_ufw *IcmpUfw) GetFirewallRuleName() string {
	return icmp_ufw.Firewall_rule_name
}

func (icmp_ufw *IcmpUfw) GetOpenPorts() string {
	return icmp_ufw.Open_ports
}

func (c *IcmpUfw) GetRule(size int, data byte) *IcmpUfwRule {
	if c.icmp_ufw_rule_caches[size] != nil {
		return c.icmp_ufw_rule_caches[size]
	}
	for _, rule := range c.GetIcmpUfwRules() {
		if rule.GetSize() == size && rule.GetPattern() == data {
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
