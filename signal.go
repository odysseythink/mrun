package mrun

import (
	"context"
	"errors"
	"log"
)

type Signal struct {
	name string
}

func (s *Signal) Init(args ...any) error {
	return nil
}
func (s *Signal) Destroy() {

}
func (s *Signal) RunOnce(ctx context.Context) error {
	return nil
}
func (s *Signal) UserData() any {
	return nil
}

func (s *Signal) Emit() any {
	return nil
}

func Connect(sender any) (*Signal, error) {
	if name == "" {
		log.Printf("[E]missing name\n")
		return nil, errors.New("missing name")
	}

	return nil, nil
}
