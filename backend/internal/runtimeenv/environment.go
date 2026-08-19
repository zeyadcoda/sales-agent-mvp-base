package runtimeenv

import (
	"errors"
	"fmt"
	"strings"
)

type ExecutionEnvironment string

const (
	Test       ExecutionEnvironment = "TEST"
	Production ExecutionEnvironment = "PRODUCTION"
)

var ErrProductionAdapterInTest = errors.New("production side-effect adapter is forbidden in TEST")

func Parse(value string) (ExecutionEnvironment, error) {
	switch ExecutionEnvironment(strings.ToUpper(strings.TrimSpace(value))) {
	case Test:
		return Test, nil
	case Production:
		return Production, nil
	default:
		return "", fmt.Errorf("invalid execution environment %q", value)
	}
}

// AssertAdapterAllowed is a hard runtime safety guard. A TEST execution plane
// must never resolve a production-capable side-effect adapter.
func AssertAdapterAllowed(env ExecutionEnvironment, adapterProductionCapable bool) error {
	if env == Test && adapterProductionCapable {
		return ErrProductionAdapterInTest
	}
	return nil
}
