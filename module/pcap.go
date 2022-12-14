package module

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"icmp-ufw/config"
	"log"
	"sync"
	"time"
)

// Pcap
// @Description: Pcap接口
// firewall: 防火墙实例
type Pcap struct {
	icmpufw  *config.IcmpUfw
	handle   []*pcap.Handle
	devices  []pcap.Interface
	firewall *Firewall
}

// GetPcap
//
//	@Description: 获取Pcap实例
//	@param icmpufw 防火墙规则
//	@return p Pcap实例
//	@return err 错误
func GetPcap(icmpufw *config.IcmpUfw) (p *Pcap, err error) {
	p = &Pcap{icmpufw: icmpufw}
	p.devices, err = pcap.FindAllDevs()
	if err != nil {
		log.Printf("FindAllDevs error: %s", err)
		return
	}
	for _, listen := range icmpufw.GetListenInterface() {
		for _, device := range p.devices {
			if listen == device.Name || listen == "0.0.0.0" {
				handle, err := pcap.OpenLive(device.Name, int32(65535), false, time.Second)
				if err != nil {
					log.Printf("OpenLive error: %s", err)
					continue
				}
				log.Printf("Listen %s", device.Name)
				p.handle = append(p.handle, handle)
			}
		}
	}
	return
}

// StartPcap
//
//	@Description: 开始监听
//	@receiver p Pcap实例
func (p *Pcap) StartPcap() {
	wg := sync.WaitGroup{}
	var webhook *Webhook
	if p.icmpufw.GetWebhookUrl() != "" {
		webhook = NewWebhook(p.icmpufw)
	}
	p.firewall = GetFirewall(p.icmpufw.GetFirewallRuleName(), p.icmpufw.GetOpenPorts(), p.icmpufw.GetFireWallProgram(), webhook)
	for _, handle := range p.handle {
		// 为每一个接口单独开启协程
		wg.Add(1)
		go func(handle *pcap.Handle) {
			defer func() {
				wg.Done()
				handle.Close()
				log.Printf("handle close")
			}()
			// 开启icmp包的监听
			err := handle.SetBPFFilter("icmp")
			if err != nil {
				log.Printf("SetBPFFilter error: %s", err)
			}
			source := gopacket.NewPacketSource(handle, handle.LinkType())
			for {
				select {
				case packet := <-source.Packets():
					networkLayer := packet.NetworkLayer()
					if networkLayer == nil {
						break
					}
					size := len(packet.Layer(gopacket.LayerTypePayload).LayerContents())
					data := packet.Data()[size+8 : 32+size-8][0]           // 获取icmp填充数据
					rule := p.icmpufw.GetRule(size, data)                  // 根据size获取规则
					ipaddress := networkLayer.NetworkFlow().Src().String() // 获取源ip
					// 如果匹配到规则
					if rule != nil {
						timeout := rule.GetTimeOut()
						if timeout == 0 {
							timeout = p.icmpufw.GetTimeOut()
						}
						p.firewall.Allow(ipaddress, rule.GetAllowPorts(), timeout)
					}
				case <-p.icmpufw.GetStop():
					return
				}
			}
		}(handle)
	}
	wg.Wait()
}

// StopPcap
//
//	@Description: 停止监听
//	@receiver p Pcap实例
func (p *Pcap) StopPcap() {
	p.firewall.Stop(true)
	log.Printf("Stop!")
}
