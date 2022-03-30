package service

import (
	"flag"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
	"time"
)

type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	if service.Interactive() {
		log.Info("Running in terminal.")
	} else {
		log.Info("Running under service manager.")
	}
	go p.run()
	p.exit = make(chan struct{})
	return nil
}
func (p *program) run() {
	log.Infof("I'm running %v.", service.Platform())
	initDbusService()
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case tm := <-ticker.C:
			log.Infof("Still running at %v...", tm)
		case <-p.exit:
			ticker.Stop()
			return
		}
	}
}

func (p *program) Stop(s service.Service) error {
	log.Info("I'm Stopping!")
	close(p.exit)
	return nil
}

// DaemonSetup Service setup.
//   Define service config.
//   Create the service.
//   Setup the logger.
//   Handle service controls (optional).
//   Run the service.
func DaemonSetup() {
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"
	svcConfig := &service.Config{
		Name:        "utMsgDaemon",
		DisplayName: "utcloud Service Test Daemon",
		Description: "This is a test Daemon service.  It is designed to run well.",
		Dependencies: []string{
			"Requires=dbus.service",
			"After=dbus.service"},
		Option: options,
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}
