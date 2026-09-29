package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/kardianos/service"
	"safora/internal/api"
	"safora/internal/database"
)

const listenAddr = "127.0.0.1:3434"

type program struct {
	db     *sql.DB
	server *api.Server
	runner *Runner
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *program) Start(s service.Service) error {
	p.ctx, p.cancel = context.WithCancel(context.Background())

	// Before any Trigger can fire, so a fresh Run is never mistaken for an orphan.
	if n, err := database.RecoverOrphanedRuns(p.db); err != nil {
		return err
	} else if n > 0 {
		log.Printf("Marked %d interrupted run(s) as failed", n)
	}

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

// ManageService installs, controls or runs the Safora system service. The
// installed service starts "safora service run", keeping --portable if it was
// used at install time, so it opens the same data directory as the installer.
func ManageService(action string, db *sql.DB, tokenPath string, portable bool) error {
	args := []string{"service", "run"}
	if portable {
		args = append([]string{"--portable"}, args...)
	}
	svcConfig := &service.Config{
		Name:        "Safora",
		DisplayName: "Safora Backup Manager",
		Description: "Automated backup management and background sync service.",
		Arguments:   args,
	}

	server, err := api.NewServer(db, tokenPath)
	if err != nil {
		return err
	}

	runner := NewRunner(db, server.Engine())
	server.OnJobsChanged = runner.SyncSchedules

	prg := &program{db: db, server: server, runner: runner}

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
