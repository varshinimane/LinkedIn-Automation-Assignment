package cmd

import (
	"context"
	"fmt"
	"time"

	"linkedin-automation/auth"
	"linkedin-automation/config"
	"linkedin-automation/connect"
	"linkedin-automation/logger"
	"linkedin-automation/messaging"
	"linkedin-automation/search"
	"linkedin-automation/stealth"
	"linkedin-automation/storage"
)

// Run executes the CLI command, orchestrating the full automation workflow:
// authentication, people search, connection requests, and follow-up messaging.
// Returns an error if any critical step fails.
func Run(cfgPath, command string) error {
	if command != "start" {
		return fmt.Errorf("unknown command: %s", command)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	log := logger.New()
	log.SetDebug(cfg.Debug)

	db, err := storage.Init(cfg.Paths.DBPath)
	if err != nil {
		return err
	}
	store := storage.NewStore(db)

	ctx := context.Background()
	authSvc := auth.Auth{Cfg: cfg, Log: log}
	browser, page, err := authSvc.StartSession(ctx)
	if err != nil {
		return err
	}
	defer browser.Close()

	profiles, err := search.PeopleSearch(ctx, page, cfg, log)
	if err != nil {
		return err
	}

	if err := connect.SendConnections(ctx, page, profiles, cfg, store, log); err != nil {
		return err
	}

	if err := messaging.SendFollowups(ctx, page, cfg, store, log); err != nil {
		return err
	}

	// End-of-cycle anti-detection cooldown.
	stealth.Cooldown(5)
	log.Info("cycle complete", "time", time.Now())
	return nil
}



