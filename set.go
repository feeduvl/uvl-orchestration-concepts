package main

import (
	"fmt"
	"strings"
)

/*
 * set implementation
 */
var exists = struct{}{}

type set struct {
	m map[string]ObservableTwitter
}

// NewSet is a custom implementation for imitating a set in golang
func NewSet() *set {
	s := &set{}
	s.m = make(map[string]ObservableTwitter)
	return s
}

func (s *set) Add(key string, observable ObservableTwitter) {
	s.m[key] = observable
}

func (s *set) Remove(key string) {
	delete(s.m, key)
}

func (s *set) Contains(key string) bool {
	_, c := s.m[key]
	return c
}

func (s *set) String() string {
	var keys []string
	for key := range s.m {
		log := fmt.Sprintf("%s ", s.m[key])
		keys = append(keys, log)
	}

	return strings.Join(keys, " ")
}
