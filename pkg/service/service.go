package service

import (
	"log"
	"os"

	"github.com/oleg-kalashnikov/onefootball/pkg/team"
)

func Setup(c map[string]struct{}) (*team.Service, *log.Logger, error) {
	log := log.New(os.Stdout, "ONEFOOTBALL> ", log.LstdFlags)
	s := team.NewService(c, log)
	return s, log, nil
}
