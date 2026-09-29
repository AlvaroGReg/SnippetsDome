package main

import (
	"testing"

	"github.com/wailsapp/wails/v2/pkg/options"
)

func TestInitialWindowStartState(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want options.WindowStartState
	}{
		{name: "normal launch", args: nil, want: options.Normal},
		{name: "autostart launch", args: []string{startMinimizedArgument}, want: options.Minimised},
		{name: "autostart launch with other arguments", args: []string{"--other", startMinimizedArgument}, want: options.Minimised},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := initialWindowStartState(test.args); got != test.want {
				t.Errorf("initialWindowStartState(%v) = %v, want %v", test.args, got, test.want)
			}
		})
	}
}

func TestStartMinimizedAtLoginIntegration(t *testing.T) {
	// TODO: verify a real operating-system auto-start launch opens minimized on supported platforms.
	t.Skip("TODO: implement operating-system startup integration test")
}
