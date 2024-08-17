package service

import "fmt"

type SessionAccepted struct {
}

func (this *SessionAccepted) String() string {
	return fmt.Sprintf("%+v", *this)
}
