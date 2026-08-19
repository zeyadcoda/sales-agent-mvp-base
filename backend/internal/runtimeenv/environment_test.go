package runtimeenv

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		input string
		want  ExecutionEnvironment
	}{
		{"TEST", Test},
		{"test", Test},
		{"PRODUCTION", Production},
	} {
		got, err := Parse(tc.input)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("Parse(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}

	if _, err := Parse("staging"); err == nil {
		t.Fatal("expected invalid execution environment to fail")
	}
}

func TestTestEnvironmentBlocksProductionAdapter(t *testing.T) {
	t.Parallel()

	err := AssertAdapterAllowed(Test, true)
	if !errors.Is(err, ErrProductionAdapterInTest) {
		t.Fatalf("got %v, want %v", err, ErrProductionAdapterInTest)
	}
}

func TestAllowedAdapterCombinations(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		env       ExecutionEnvironment
		prod      bool
		shouldErr bool
	}{
		{Test, false, false},
		{Production, false, false},
		{Production, true, false},
	} {
		if err := AssertAdapterAllowed(tc.env, tc.prod); (err != nil) != tc.shouldErr {
			t.Fatalf("AssertAdapterAllowed(%q, %v) = %v", tc.env, tc.prod, err)
		}
	}
}
