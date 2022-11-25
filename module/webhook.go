package module

import (
	"icmp-ufw/config"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Webhook struct {
	webhook_url     string
	webhook_method  string
	webhook_headers []string
	webhook_data    string
}

type FirewallLog struct {
	Address string
	Time    time.Time
	Ports   string
}

func NewWebhook(icmpufw *config.IcmpUfw) *Webhook {
	return &Webhook{
		webhook_url:     icmpufw.Webhook_url,
		webhook_method:  icmpufw.Webhook_method,
		webhook_headers: icmpufw.Webhook_headers,
		webhook_data:    icmpufw.Webhook_data,
	}
}

func (c *Webhook) SetWebhookUrl(webhook_url string) *Webhook {
	c.webhook_url = webhook_url
	return c
}

func (c *Webhook) SetWebhookMethod(webhook_method string) *Webhook {
	c.webhook_method = webhook_method
	return c
}

func (c *Webhook) replace(webhook_data string, firewallLog *FirewallLog) string {
	webhook_data = strings.ReplaceAll(webhook_data, "{address}", url.QueryEscape(firewallLog.Address))
	webhook_data = strings.ReplaceAll(webhook_data, "{time}", url.QueryEscape(firewallLog.Time.Format("2006-01-02 15:04:05")))
	webhook_data = strings.ReplaceAll(webhook_data, "{ports}", url.QueryEscape((firewallLog.Ports)))
	return webhook_data
}

func (c *Webhook) SendData(firewallLog *FirewallLog) *Webhook {
	webhook_url := c.webhook_url
	webhook_data := c.replace(c.webhook_data, firewallLog)
	if c.webhook_method == "GET" {
		webhook_data = ""
		webhook_url = c.replace(webhook_url, firewallLog)
	}
	req, err := http.NewRequest(c.webhook_method, c.webhook_url, strings.NewReader(webhook_data))
	if err != nil {
		log.Printf("Webhook: %s", err)
	}
	for _, i := range c.webhook_headers {
		header := strings.Split(i, ":")
		req.Header.Set(header[0], header[1])
	}
	if c.webhook_method != "GET" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	_, err = (&http.Client{}).Do(req)
	if err != nil {
		log.Printf("Webhook: %s", err)
	}
	return c
}

func (c *Webhook) SetWebhookHeaders(webhook_headers []string) *Webhook {
	c.webhook_headers = webhook_headers
	return c
}
