package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	my_game "mafia/proto"
	"mafia/server/constants"

	"google.golang.org/grpc"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Проиграл")
	} else {
		fmt.Println("Ok!")
	}
}

func run() error {

	port := os.Args[1]

	conn, err := net.ListenPacket("udp", ":"+port)
	if err != nil {
		return fmt.Errorf("Error in ListenPacket")
	}
	defer conn.Close()

	// small talk
	buf := make([]byte, 4096)
	_, addr, err := conn.ReadFrom(buf)
	if err != nil {
		return fmt.Errorf("Bad read")
	}

	_, err = conn.WriteTo([]byte("Введите свой никнейм: "), addr)

	buf = make([]byte, 4096)
	n, addr, err := conn.ReadFrom(buf)
	startGame(string(buf[:n]), conn, addr)
	return nil
}

func startGame(nickName string, user_conn net.PacketConn, user_addr net.Addr) error {
	mySessiondId := int32(0)
	nickName = strings.Replace(nickName, "\n", "", -1)
	conn, err := grpc.Dial(
		"server:9000",
		grpc.WithInsecure(),
	)

	if err != nil {
		return fmt.Errorf("failed to establish a connection with server: %w", err)
	}
	defer conn.Close()
	fmt.Println("Your nickname", nickName)

	// Create GRPC client
	client := my_game.NewMafiaClient(conn)

	// Registration gamer
	events, err := client.Start(context.Background(), &my_game.UserInfo{
		Username: nickName,
	})
	if err != nil {
		return fmt.Errorf("failed to join the game: %w", err)
	}
	for {
		event, err := events.Recv()
		if err != nil {
			panic("I let this city down")
		}

		switch event := event.EventVariant.(type) {
		case *my_game.Event_Booked:
			_, err = user_conn.WriteTo([]byte("Ты успешно зарегистрирован(а), ожидай начало игры \n"), user_addr)
			if err != nil {
				fmt.Println("Bad Event Booked")
			}
		case *my_game.Event_NewGame:
			mySessiondId = event.NewGame.SessionId
			role := parseRole(event.NewGame.Role)
			_, err = user_conn.WriteTo([]byte(event.NewGame.Username+", Game has begun\nYour Role: "+role+"\n"), user_addr)
			if err != nil {
				fmt.Println("Bad New Game")
			}
		case *my_game.Event_NightIsComming:
			switch event.NightIsComming.Action {
			case my_game.Actions_KILL:
				_, err = user_conn.WriteTo([]byte(constants.MafiaPhrase+event.NightIsComming.ListOfPeople+"\n"), user_addr)
				victim, err := makeVictim(user_conn, mySessiondId)
				if err != nil {
					return err
				}
				_, err = client.VoteProcessing(context.Background(), victim)
				if err != nil {
					return err
				}
			case my_game.Actions_CHECK:
				_, err = user_conn.WriteTo([]byte(constants.ComissarPhrase+event.NightIsComming.ListOfPeople+"\n"), user_addr)
				victim, err := makeVictim(user_conn, mySessiondId)
				if err != nil {
					return err
				}
				_, err = client.VoteProcessing(context.Background(), victim)
				if err != nil {
					return err
				}
			default:
				_, err = user_conn.WriteTo([]byte("Наступила ночь, город засыпает\n"), user_addr)
			}
		case *my_game.Event_MorningIsComing:
			if !event.MorningIsComing.Alive {
				_, err = user_conn.WriteTo([]byte("Наступило утро,\n"+event.MorningIsComing.Announce), user_addr)

			} else {
				_, err = user_conn.WriteTo([]byte("Наступило утро,\n"+event.MorningIsComing.Announce+
					constants.Vouting+"\n"+event.MorningIsComing.ListOfPeople), user_addr)

				victim, err := makeVictim(user_conn, mySessiondId)
				if err != nil {
					return err
				}

				_, err = client.VoteProcessing(context.Background(), victim)
				if err != nil {
					return err
				}
			}

		case *my_game.Event_DiedName:
			_, err = user_conn.WriteTo([]byte("Игру заканчивает:"+event.DiedName+"\n"), user_addr)
		case *my_game.Event_Result:
			_, err = user_conn.WriteTo([]byte("Игра завершилась, "+event.Result+"\n"), user_addr)
		default:
			fmt.Println("Not implemented")

		}
	}
}

func makeVictim(user_conn net.PacketConn, sessionId int32) (*my_game.Victim, error) {
	buf := make([]byte, 4096)
	n, _, err := user_conn.ReadFrom(buf)
	if err != nil {
		return nil, err
	}
	str := strings.Replace(string(buf[:n]), "\n", "", -1)
	return &my_game.Victim{Username: str, SessionId: sessionId}, nil

}

func parseRole(role my_game.Role) string {
	switch role {
	case my_game.Role_CIVILIAN:
		return "Civilian"
	case my_game.Role_COMISSAR:
		return "Comissar"
	default:
		return "Mafia"
	}
}
