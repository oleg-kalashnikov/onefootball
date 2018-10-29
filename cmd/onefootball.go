package main

import (
	"github.com/oleg-kalashnikov/onefootball/pkg/service"
)

var conf = map[string]struct{}{
	"Germany":          {},
	"England":          {},
	"France":           {},
	"Spain":            {},
	"Manchester Utd":   {},
	"Arsenal":          {},
	"Chelsea":          {},
	"Barcelona":        {},
	"Real Madrid":      {},
	"FC Bayern Munich": {},
}

func main() {
	s, log, err := service.Setup(conf)
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}
