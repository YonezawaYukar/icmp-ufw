package module

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"icmp-ufw/config"
	"log"
	"sync"
)

type Pcap struct {
	icmpufw  *config.IcmpUfw
	handle   []*pcap.Handle
	devices  []pcap.Interface
	firewall []*Firewall
}

func GetPcap(icmpufw *config.IcmpUfw) (p *Pcap, err error) {
	p = &Pcap{icmpufw: icmpufw}
	p.devices, err = pcap.FindAllDevs()
	if err != nil {
		return
	}
	for _, listen := range icmpufw.GetListenInterface() {
		for _, device := range p.devices {
			if listen == device.Name || listen == "0.0.0.0" {
				handle, err := pcap.OpenLive(device.Name, int32(65535), false, pcap.BlockForever)
				if err != nil {
					log.Fatal(err)
					continue
				}
				log.Printf("Listen %s", device.Name)
				p.handle = append(p.handle, handle)
				handle.SetBPFFilter("icmp")
			}
		}
	}
	return
}

func (p *Pcap) StartPcap() {
	wg := sync.WaitGroup{}
	for _, handle := range p.handle {
		wg.Add(1)
		go func(handle *pcap.Handle, icmpufw *config.IcmpUfw) {
			defer wg.Done()
			defer handle.Close()
			handle.SetBPFFilter("icmp")
			source := gopacket.NewPacketSource(handle, handle.LinkType())
			for {
				select {
				case packet := <-source.Packets():
					networkLayer := packet.NetworkLayer()
					if networkLayer == nil {
						break
					}
					size := len(packet.Data()) - 32
					rule := icmpufw.GetRule(size)
					ipaddress := networkLayer.NetworkFlow().Src().String()
					if rule != nil {
						ruleGroupName := fmt.Sprintf("icmp-ufw-%s-%s", ipaddress, RandStr(6))
						p.firewall = append(p.firewall, AllowAccept(ipaddress, rule.GetAllowPort(), icmpufw.GetFireWallProgram(), ruleGroupName))
						cache := fmt.Sprintf("%s %s %s", ipaddress, rule.GetAllowPort(), icmpufw.GetFireWallProgram())
						if icmpufw.GetFireWallProgram() == "iptables" {
							cache += " " + ruleGroupName
						}
						icmpufw.GetArgs().SyncWrite.Writer.Write([]byte(cache + "\n"))
					}
				case <-icmpufw.GetStop():
					return
				}
			}
		}(handle, p.icmpufw)
	}
	defer wg.Wait()
}

func (p *Pcap) StopPcap() {
	for _, firewall := range p.firewall {
		firewall.Stop()
	}
	log.Printf("Stop!")
}
