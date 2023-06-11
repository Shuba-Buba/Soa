package main

import (
	"context"
	"fmt"
	"log"
	my_game "mafia/proto"
	"mafia/server/constants"
	"math/rand"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type Server struct {
	my_game.UnimplementedMafiaServer
	mutex          sync.Mutex
	sessionId      int
	cntParticipant int
	members        chan constants.Info
	IdsToChans     map[int32]chan string
}

func InitGame(sessionId int, members <-chan constants.Info, ch chan<- constants.Post, votes <-chan string) {
	winMafia := false
	res := constants.WinCivials

	mafia_cnt := constants.CountOfGamers / 3
	commissar_cnt := 1
	civilian_cnt := constants.CountOfGamers - mafia_cnt

	roles := []my_game.Role{}
	for i := 0; i < mafia_cnt; i++ {
		roles = append(roles, my_game.Role_MAFIA)
	}
	roles = append(roles, my_game.Role_COMISSAR)
	for i := 0; i < civilian_cnt-1; i++ {
		roles = append(roles, my_game.Role_CIVILIAN)
	}
	rand.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})

	membersInfo := make([]constants.Info, 0, constants.CountOfGamers)

	died := []bool{}
	userNameToId := make(map[string]int)

	//Выдача ролей
	for i := 0; i < constants.CountOfGamers; i++ {
		cur_people := <-members
		membersInfo = append(membersInfo, constants.Info{
			Username: cur_people.Username,
			Stream:   cur_people.Stream,
			Role:     roles[i],
		})

		died = append(died, false)
		userNameToId[membersInfo[i].Username] = i

	}
	for i, val := range membersInfo {
		ch <- constants.Post{&my_game.Event{
			EventVariant: &my_game.Event_NewGame{&my_game.User{Username: membersInfo[i].Username,
				Role:      membersInfo[i].Role,
				SessionId: int32(sessionId)}},
		},
			val.Stream}
	}

	for {
		if mafia_cnt >= civilian_cnt {
			winMafia = true
			break
		}
		if mafia_cnt == 0 {
			break
		}

		// Логика игры

		civilian_usernames := getLivingPeopleWithoutMafias(died, membersInfo)
		all_civilian := getLivingPeople(died, membersInfo)

		comissarId := -1

		for idx, val := range membersInfo {
			if died[idx] == false {
				switch val.Role {
				case my_game.Role_MAFIA:
					ch <- constants.Post{
						&my_game.Event{
							EventVariant: &my_game.Event_NightIsComming{
								&my_game.Night{
									ListOfPeople: civilian_usernames, Action: my_game.Actions_KILL,
								},
							},
						},
						val.Stream,
					}
				case my_game.Role_COMISSAR:
					comissarId = idx
				default:
					ch <- constants.Post{
						&my_game.Event{
							EventVariant: &my_game.Event_NightIsComming{
								&my_game.Night{
									ListOfPeople: "", Action: my_game.Actions_NOTHING,
								},
							},
						},
						val.Stream,
					}
				}
			}
		}

		// Убийства и проверки

		diedName := generatedNick(mafia_cnt, userNameToId, votes)
		isMafia := ""

		if comissarId != -1 {
			ch <- constants.Post{
				&my_game.Event{
					EventVariant: &my_game.Event_NightIsComming{
						&my_game.Night{
							ListOfPeople: all_civilian, Action: my_game.Actions_CHECK,
						},
					},
				},
				membersInfo[comissarId].Stream,
			}
			isMafia = generatedNick(commissar_cnt, userNameToId, votes)
		}

		annonce := "Был убит: " + diedName + "\n"
		if comissarId != -1 && checkOnMafia(isMafia, userNameToId, membersInfo) {
			annonce += "Комиссар нашел мафию:" + isMafia + "\n"
		} else {
			annonce += "Комиссар, не нашел мафию\n"
		}
		if mafia_cnt > 0 {
			makeKill(died, diedName, userNameToId)
		}
		listOfPeople := getLivingPeople(died, membersInfo) + "\n"

		mafia_cnt, commissar_cnt, civilian_cnt = reWrite(membersInfo, died)
		if mafia_cnt >= civilian_cnt {
			winMafia = true
			break
		}

		for idx, val := range membersInfo {
			ch <- constants.Post{
				&my_game.Event{
					EventVariant: &my_game.Event_MorningIsComing{
						&my_game.Morning{
							ListOfPeople: listOfPeople,
							Announce:     annonce,
							Alive:        !died[idx],
						},
					},
				},
				val.Stream,
			}
		}

		newDied := generatedNick(mafia_cnt+civilian_cnt, userNameToId, votes)
		makeKill(died, newDied, userNameToId)

		for _, val := range membersInfo {
			ch <- constants.Post{
				&my_game.Event{
					EventVariant: &my_game.Event_DiedName{
						DiedName: newDied,
					},
				},
				val.Stream,
			}
		}
		mafia_cnt, commissar_cnt, civilian_cnt = reWrite(membersInfo, died)
	}
	if winMafia {
		res = constants.WinMafia
	}

	for _, val := range membersInfo {
		ch <- constants.Post{
			&my_game.Event{
				EventVariant: &my_game.Event_Result{
					Result: res,
				},
			},
			val.Stream,
		}
	}
}

func reWrite(membersInfo []constants.Info, died []bool) (int, int, int) {
	mafia_cnt := 0
	commissar_cnt := 0
	civilian_cnt := 0

	for idx, val := range membersInfo {
		if !died[idx] {
			if val.Role == my_game.Role_MAFIA {
				mafia_cnt += 1
			} else {
				civilian_cnt += 1
				if val.Role == my_game.Role_COMISSAR {
					commissar_cnt += 1
				}
			}
		}
	}
	return mafia_cnt, commissar_cnt, civilian_cnt
}

func checkOnMafia(nick string, userNameToId map[string]int, membersInfo []constants.Info) bool {
	id := userNameToId[nick]

	if membersInfo[id].Role == my_game.Role_MAFIA {
		return true
	}
	return false
}

func makeKill(died []bool, nick string, userNameToId map[string]int) {
	id := userNameToId[nick]
	died[id] = true
}

func generatedNick(n int, userNameToId map[string]int, votes <-chan string) string {

	mp := make(map[string]int)

	for i := 0; i < n; i++ {
		killNick := <-votes
		mp[killNick] += 1
	}

	mx := 0
	name := ""
	for key, val := range mp {
		_, ok := userNameToId[key]
		if mx < val && ok {
			mx = val
			name = key
		}
	}
	return name
}

func getLivingPeopleWithoutMafias(died []bool, membersInfo []constants.Info) string {
	res := ""
	for idx, val := range membersInfo {
		if died[idx] == false && val.Role != my_game.Role_MAFIA {
			res += val.Username + " "
		}
	}
	return res
}

func getLivingPeople(died []bool, membersInfo []constants.Info) string {
	res := ""
	for idx, val := range membersInfo {
		if died[idx] == false {
			res += val.Username + " "
		}
	}
	return res
}

func eventWriter(ch <-chan constants.Post) {
	for newEvent := range ch {
		newEvent.Stream.Send(newEvent.Event)
	}
}

func (s *Server) Start(in *my_game.UserInfo, stream my_game.Mafia_StartServer) error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	s.mutex.Lock()

	s.cntParticipant += 1
	if s.cntParticipant == constants.CountOfGamers {
		s.cntParticipant = 0
		s.sessionId += 1

		ch := make(chan constants.Post, 10000)
		s.IdsToChans[int32(s.sessionId)] = make(chan string, 100)

		go func(sessionId int) {
			InitGame(sessionId, s.members, ch, s.IdsToChans[int32(s.sessionId)])

		}(s.sessionId)

		go eventWriter(ch)

	}
	s.mutex.Unlock()

	s.members <- constants.Info{Username: in.Username, Stream: stream}
	stream.Send(&my_game.Event{
		EventVariant: &my_game.Event_Booked{},
	})
	wg.Wait()
	return fmt.Errorf("Not implemented")
}

func (s *Server) VoteProcessing(ctx context.Context, in *my_game.Victim) (*my_game.Empty, error) {
	ch, ok := s.IdsToChans[in.SessionId]
	if !ok {
		return nil, fmt.Errorf("No exist")
	}
	ch <- in.Username
	return &my_game.Empty{}, nil
}

func main() {

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	my_game.RegisterMafiaServer(srv, &Server{members: make(chan constants.Info, 100), IdsToChans: make(map[int32]chan string)})

	log.Fatalln(srv.Serve(lis))
}
