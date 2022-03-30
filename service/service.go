package service

import (
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *program) run() {
	initDbusService()
}

func (p *program) Stop(s service.Service) error {
	stopped <- struct{}{}
	return nil
}

func InitDaemon() {
	svcConfig := &service.Config{
		Name:        "github.com/jibenliu/utMsgDaemon",
		DisplayName: "utcloud Service Test Daemon",
		Description: "This is a test Daemon service.  It is designed to run well.",
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		log.Error(err)
	}
}
