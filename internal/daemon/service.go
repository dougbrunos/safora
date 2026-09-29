package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/kardianos/service"
	"safora/internal/api"
)

const listenAddr = "127.0.0.1:3434"

type program struct {
	server *api.Server
	runner *Runner
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *program) Start(s service.Service) error {
	p.ctx, p.cancel = context.WithCancel(context.Background())

	go func() {
		if err := p.server.Start(listenAddr); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}()
	go p.runner.Start(p.ctx)

	return nil
}

func (p *program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func ManageService(action string, db *sql.DB) error {
	svcConfig := &service.Config{
		Name:        "Safora",
		DisplayName: "Safora Backup Manager",
		Description: "Automated backup management and background sync service.",
	}

	prg := &program{
		server: api.NewServer(db),
		runner: NewRunner(db),
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		return err
	}

	switch action {
	case "install":
		return s.Install()
	case "uninstall":
		return s.Uninstall()
	case "start":
		return s.Start()
	case "stop":
		return s.Stop()
	case "status":
		status, err := s.Status()
		if err != nil {
			return err
		}
		switch status {
		case service.StatusRunning:
			fmt.Println("Service is running")
		case service.StatusStopped:
			fmt.Println("Service is stopped")
		default:
			fmt.Println("Service status unknown")
		}
		return nil
	case "run":
		return s.Run()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
