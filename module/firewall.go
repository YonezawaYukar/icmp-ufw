package module

import (
	"fmt"
	"log"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"time"
)

func in_array(val interface{}, array interface{}) (exists bool) {
	exists = false
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				exists = true
				return
			}
		}
	}
	return
}

/**
 * 防火墙操作
 * @param ruleGroupName 规则组名称
 * @param allowPort 允许端口
 * @param firewallProgram 防火墙程序
 */
type Firewall struct {
	ruleGroupName   string              // 规则链名称
	allowPort       string              // 默认开放端口
	firewallProgram string              //iptables程序所在位置
	caches          map[string][]string //更新缓存
	cache_lock      sync.Mutex          //更新缓存锁
	time_out_table  []struct {
		Address string
		Ports   string
		Time    time.Time
	} //超时表
	time_out_lock sync.Mutex //超时锁
	time_out_stop bool
}

// 获取防火墙实例
func GetFirewall(ruleGroupName string, allowPort string, firewallProgram string) (firewall *Firewall) {
	firewall = &Firewall{allowPort: allowPort, firewallProgram: firewallProgram, ruleGroupName: ruleGroupName}
	firewall.cache_lock = sync.Mutex{}
	firewall.Start()
	//开启定时器
	//定时器每秒检查一次超时表
	//如果超时则删除缓存并且删除防火墙规则
	go func(firewall *Firewall) {
		for {
			if firewall.time_out_stop {
				break
			}
			firewall.time_out_lock.Lock()
			for index, item := range firewall.time_out_table {
				if time.Now().After(item.Time) {
					//超时操作
					firewall.Delete(item.Address, item.Ports)
					firewall.time_out_table = append(firewall.time_out_table[:index], firewall.time_out_table[index+1:]...)
				}
			}
			firewall.time_out_lock.Unlock()
			time.Sleep(time.Second * 3)
		}
	}(firewall)
	return firewall
}

// 更新防火墙
func (firewall *Firewall) Update(ruleGroupName string, allowPort string) {
	firewall.Stop(false)
	firewall.ruleGroupName = ruleGroupName
	firewall.allowPort = allowPort
	firewall.Start()
	firewall.cache_lock.Lock() //上锁 防止并发
	for address, allowPorts := range firewall.caches {
		for _, allowPort := range allowPorts {
			firewall.Allow(address, allowPort, 0)
		}
	}
	firewall.cache_lock.Unlock()
}

// 执行命令
func (firewall *Firewall) command(command string) {
	log.Printf("执行命令:%s %s\n", firewall.firewallProgram, command)
	err := exec.Command(firewall.firewallProgram, strings.Split(command, " ")...).Run()
	if err != nil {
		log.Print(err)
	}
}

// 启动防火墙
func (firewall *Firewall) Start() {
	//sudo iptables -t filter -N IN_WEB
	firewall.caches = make(map[string][]string)
	firewall.command("-t filter -N " + firewall.ruleGroupName)
	firewall.command("-t filter -F " + firewall.ruleGroupName)
	firewall.command("-t filter -I INPUT -p tcp --dport 1:65535 -j REJECT -I " + firewall.ruleGroupName)
	firewall.command("-t filter -p udp --dport 1:65535 -j REJECT -I " + firewall.ruleGroupName)
	firewall.command("-t filter -p icmp -j ACCEPT -I " + firewall.ruleGroupName)
	firewall.command(" -i lo -j ACCEPT -I " + firewall.ruleGroupName)
	for _, port := range strings.Split(firewall.allowPort, ",") {
		firewall.command(fmt.Sprintf("-t filter -I %s -p tcp --dport %s -j ACCEPT", firewall.ruleGroupName, port))
		firewall.command(fmt.Sprintf("-t filter -I %s -p udp --dport %s -j ACCEPT", firewall.ruleGroupName, port))
	}

}

// 停止防火墙
func (firewall *Firewall) Stop(stopTimeOut bool) {
	if stopTimeOut {
		firewall.time_out_stop = true
	}
	firewall.command("-t filter -F " + firewall.ruleGroupName)
	firewall.command("-t filter -X " + firewall.ruleGroupName)
}

// 允许访问
func (firewall *Firewall) Allow(address string, allowPorts string, timeOut int) {
	firewall.cache_lock.Lock() //上锁 防止并发
	for _, port := range strings.Split(allowPorts, ",") {
		if firewall.caches[address] == nil {
			firewall.caches[address] = make([]string, 0)
		}
		if !in_array(port, firewall.caches[address]) {
			//如果没有缓存则添加 并且开放端口
			firewall.caches[address] = append(firewall.caches[address], port)
			firewall.command(fmt.Sprintf("-t filter -I %s -s %s -p tcp --dport %s -j ACCEPT", firewall.ruleGroupName, address, port))
			firewall.command(fmt.Sprintf("-t filter -I %s -s %s -p udp --dport %s -j ACCEPT", firewall.ruleGroupName, address, port))
		}
	}
	firewall.cache_lock.Unlock()
	if timeOut > 0 {
		firewall.time_out_lock.Lock()
		firewall.time_out_table = append(firewall.time_out_table, struct {
			Address string
			Ports   string
			Time    time.Time
		}{Address: address, Ports: allowPorts, Time: time.Now().Add(time.Second * time.Duration(timeOut))})
		firewall.time_out_lock.Unlock()
	}
}

// 删除访问
func (firewall *Firewall) Delete(address string, allowPorts string) {
	if firewall.caches[address] == nil {
		return
	}
	firewall.cache_lock.Lock() //上锁 防止并发
	for _, port := range strings.Split(allowPorts, ",") {
		if in_array(port, firewall.caches[address]) {
			//如果缓存中有则删除 并且关闭端口
			delete(firewall.caches, address)
			firewall.command(fmt.Sprintf("-t filter -D %s -s %s -p tcp --dport %s -j ACCEPT", firewall.ruleGroupName, address, port))
			firewall.command(fmt.Sprintf("-t filter -D %s -s %s -p udp --dport %s -j ACCEPT", firewall.ruleGroupName, address, port))
		}
	}
	firewall.cache_lock.Unlock()
}
