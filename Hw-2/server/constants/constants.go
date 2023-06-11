package constants

import (
	my_game "mafia/proto"
)

const (
	CountOfGamers  = 4
	MafiaPhrase    = "Наступила ночь, напиши имя того, кого хочешь убить, из списка доступных имен ниже\n"
	ComissarPhrase = "Наступила ночь, напиши имя того, кого хочешь проверить, из списка доступных имен ниже\n"
	Vouting        = "\nВыбирайте за кого голосуете из списка ниже"
	WinCivials     = "Победили мирные жители"
	WinMafia       = "Победила мафия"
)

type Info struct {
	Username string
	Stream   my_game.Mafia_StartServer
	Role     my_game.Role
}

type Post struct {
	Event  *my_game.Event
	Stream my_game.Mafia_StartServer
}
