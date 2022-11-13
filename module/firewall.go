package module

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
)

type Firewall struct {
	ruleGroupName   string
	ipaddress       string
	allowPort       string
	fireWallProgram string
}

func Clear(cache string) {
	scanner := bufio.NewScanner(strings.NewReader(cache))
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		if len(line) < 3 {
			continue
		}
		firewall := Firewall{ipaddress: line[0], allowPort: line[1], fireWallProgram: line[2]}
		if len(line) == 4 {
			firewall.ruleGroupName = line[3]
		}
		firewall.Stop()
		log.Printf("delete %s %s %s", firewall.ipaddress, firewall.allowPort, firewall.fireWallProgram)
	}
}

func AllowAccept(ipaddress string, allowPort string, fireWallProgram string, ruleGroupName string) (firewall *Firewall) {
	firewall = &Firewall{ipaddress: ipaddress, allowPort: allowPort, fireWallProgram: fireWallProgram}
	if fireWallProgram == "ufw" {
		firewall.allowUfw()
	} else if fireWallProgram == "iptables" {
		firewall.allowIptables()
	}
	return firewall
}

func (firewall *Firewall) Stop() {
	if firewall.fireWallProgram == "ufw" {
		firewall.deleteUfw()
	} else if firewall.fireWallProgram == "iptables" {
		firewall.deleteIptables()
	}
}

func (firewall *Firewall) allowUfw() {
	// ufw allow from
	err := exec.Command("ufw", "allow", "proto", "tcp", "from", firewall.ipaddress, "to", "any", "port", firewall.allowPort).Run()
	if err != nil {
		log.Print(err)
	}
	err = exec.Command("ufw", "allow", "proto", "udp", "from", firewall.ipaddress, "to", "any", "port", firewall.allowPort).Run()
	if err != nil {
		log.Print(err)
	}
	log.Printf("%s open port %s", firewall.ipaddress, firewall.allowPort)
}

func (firewall *Firewall) allowIptables() {
	//iptables allow from
}
func (firewall *Firewall) deleteUfw() {
	// ufw delete
	err := exec.Command("ufw", "delete", "proto", "tcp", "from", firewall.ipaddress, "to", "any", "port", firewall.allowPort).Run()
	if err != nil {
		log.Print(err)
	}
	err = exec.Command("ufw", "delete", "proto", "udp", "from", firewall.ipaddress, "to", "any", "port", firewall.allowPort).Run()
	if err != nil {
		log.Print(err)
	}
	log.Printf("%s open port %s", firewall.ipaddress, firewall.allowPort)
}

func (firewall *Firewall) deleteIptables() {
	//iptables delete
}
