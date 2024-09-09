package helper

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	password := "password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}
	if hash == "" {
		t.Error("Hash is empty")
	}
	fmt.Println(hash)
}

func TestCoalesceString(t *testing.T) {
	s1 := ""
	s2 := "s2"
	s3 := "s3"
	s4 := ""
	s5 := CoalesceString(s1, s2, s3, s4)
	if s5 != s2 {
		t.Errorf("Expected %s, got %s", s1, s5)
	}
}
