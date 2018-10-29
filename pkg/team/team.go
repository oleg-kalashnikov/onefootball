package team

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	sourceURL    = "https://vintagemonster.onefootball.com/api/teams/en/%d.json"
	workersCount = 10
)

type feed struct {
	Data struct {
		Team struct {
			Name    string        `json:"name"`
			Players []*feedPlayer `json:"players"`
		} `json:"team"`
	} `json:"data"`
}

type feedPlayer struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       string `json:"age"`
}

type Player struct {
	ID    string
	Name  string
	Age   string
	Teams []string
}

func (p *Player) String() {

}

type Service struct {
	m          *sync.Mutex
	httpClient *http.Client
	log        *log.Logger
	DoneChan   chan struct{}
	needs      map[string]struct{}
	Players    map[string]*Player
}

func NewService(c map[string]struct{}, l *log.Logger) *Service {
	return &Service{
		httpClient: &http.Client{},
		log:        l,
		DoneChan:   make(chan struct{}, 1),
		m:          &sync.Mutex{},
		needs:      c,
		Players:    map[string]*Player{},
	}
}

func (s *Service) Start() {
	urlsChan := s.URLs()
	s.startWorkers(urlsChan)
	s.log.Printf("Result:\n%s", s.Print())
}

func (s *Service) startWorkers(urlChan chan string) {
	wg := &sync.WaitGroup{}
	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urlChan {
				resp, err := s.httpClient.Get(u)
				if err != nil {
					s.log.Printf(err.Error())
					return
				}
				err = s.Process(resp.Body)
				if err != nil {
					s.log.Printf(err.Error())
					return
				}
			}
		}()
	}
	wg.Wait()
}

func (s *Service) URLs() chan string {
	urlChan := make(chan string, workersCount)
	go func() {
		for i := 1; true; i++ {
			select {
			case urlChan <- fmt.Sprintf(sourceURL, i):
			case <-s.DoneChan:
				close(urlChan)
				return
			}
		}
	}()
	return urlChan
}

func (s *Service) Process(rc io.ReadCloser) error {
	defer rc.Close()

	var f feed
	err := json.NewDecoder(rc).Decode(&f)
	if err != nil {
		return err
	}

	if _, ok := s.needs[f.Data.Team.Name]; ok {
		s.log.Printf("Found team: %s", f.Data.Team.Name)

		s.delete(f.Data.Team.Name)

		for _, player := range f.Data.Team.Players {
			s.addPlayer(player, f.Data.Team.Name)
		}
		if len(s.needs) == 0 {
			s.log.Printf("Jobe done!")
			s.DoneChan <- struct{}{}
		}
	}
	return nil
}

func (s *Service) delete(teamName string) {
	s.m.Lock()
	delete(s.needs, teamName)
	s.m.Unlock()
}

func (s *Service) addPlayer(player *feedPlayer, teamName string) {
	s.m.Lock()
	defer s.m.Unlock()
	if p, ok := s.Players[player.ID]; ok {
		p.Teams = append(p.Teams, teamName)
		return
	}
	fullName := strings.Join([]string{player.FirstName, player.LastName}, " ")
	s.Players[player.ID] = &Player{player.ID, fullName, player.Age, []string{teamName}}
}

func (s *Service) Print() string {
	players := make([]*Player, 0, len(s.Players))
	for _, p := range s.Players {
		players = append(players, p)
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].Name < players[j].Name
	})

	var buf bytes.Buffer
	for idx, p := range players {
		buf.WriteString(strconv.Itoa(idx + 1))
		buf.WriteString(". ")
		buf.WriteString(p.Name)
		buf.WriteString("; ")
		buf.WriteString(p.Age)
		buf.WriteString("; ")
		buf.WriteString(strings.Join(p.Teams, ", "))
		buf.WriteByte('\n')
	}
	return buf.String()
}
