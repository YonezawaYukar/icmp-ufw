package main

import (
	"flag"
	"icmp-ufw/config"
	"icmp-ufw/module"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	args *config.Args_struct
	pcap *module.Pcap
)

func flagInit() {
	args = &config.Args_struct{}
	flag.StringVar(&args.ConfigFile, "c", "/usr/local/share/icmpfw/config.yaml", "config file")
	flag.BoolVar(&args.Help, "h", false, "help")
	flag.StringVar(&args.HotUpdate, "hotUpdate", "false", "hotUpdate")
	flag.BoolVar(&args.AutoReload, "autoReload", false, "autoReload")
	flag.IntVar(&args.TimeOut, "timeOut", 3600, "timeOut")
	flag.Parse()
}
func main() {
	flagInit()
	if args.Help != false {
		flag.Usage()
		return
	}
	icmpufw, err := config.GetConfig(args)
	if err != nil {
		log.Fatal(err)
		return
	}
	pcap, err = module.GetPcap(icmpufw)
	if err != nil {
		log.Fatal(err)
		return
	}
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
				icmpufw.SetStop()
			}
		}
	}()
	log.Printf("Start!")
	defer pcap.StopPcap()
	pcap.StartPcap()
}
