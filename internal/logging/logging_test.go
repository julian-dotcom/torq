package logging

import (
	"github.com/rs/zerolog"
	"testing"
)

func TestInitDevLog(t *testing.T) {
	//zerolog.SetGlobalLevel(zerolog.InfoLevel)
	got := zerolog.GlobalLevel()
	want := zerolog.TraceLevel
	if got == want {
		t.Logf("Passed")
	} else {
		t.Fatalf("Failed")
	}
}

func TestInitLogProd(t *testing.T) {
	//zerolog.SetGlobalLevel(zerolog.Disabled)
	got := zerolog.GlobalLevel()
	want := zerolog.TraceLevel
	if got != want {
		t.Errorf("Failed")
	}
}
