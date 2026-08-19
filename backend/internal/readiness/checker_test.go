package readiness

import (
	"context"
	"errors"
	"testing"
)

type fakePinger struct {
	err   error
	calls int
}

func (p *fakePinger) Ping(_ context.Context) error {
	p.calls++
	return p.err
}

func TestCheckRequiresHealthyPostgresAndRedis(t *testing.T) {
	t.Parallel()

	postgresFailure := errors.New("simulated PostgreSQL failure")
	redisFailure := errors.New("simulated Redis failure")

	tests := []struct {
		name              string
		postgres          *fakePinger
		redis             *fakePinger
		wantErr           error
		wantPostgresCalls int
		wantRedisCalls    int
	}{
		{
			name:              "both dependencies healthy",
			postgres:          &fakePinger{},
			redis:             &fakePinger{},
			wantPostgresCalls: 1,
			wantRedisCalls:    1,
		},
		{
			name:              "PostgreSQL unavailable",
			postgres:          &fakePinger{err: postgresFailure},
			redis:             &fakePinger{},
			wantErr:           postgresFailure,
			wantPostgresCalls: 1,
		},
		{
			name:              "Redis unavailable",
			postgres:          &fakePinger{},
			redis:             &fakePinger{err: redisFailure},
			wantErr:           redisFailure,
			wantPostgresCalls: 1,
			wantRedisCalls:    1,
		},
		{
			name:    "PostgreSQL missing",
			redis:   &fakePinger{},
			wantErr: ErrDependencyMissing,
		},
		{
			name:     "Redis missing",
			postgres: &fakePinger{},
			wantErr:  ErrDependencyMissing,
		},
		{
			name:    "both dependencies missing",
			wantErr: ErrDependencyMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checker := New(tt.postgres, tt.redis)
			err := checker.Check(context.Background())

			if tt.wantErr == nil && err != nil {
				t.Fatalf("Check() error = %v, want nil", err)
			}

			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Check() error = %v, want error matching %v", err, tt.wantErr)
			}

			if tt.postgres != nil && tt.postgres.calls != tt.wantPostgresCalls {
				t.Fatalf(
					"PostgreSQL Ping calls = %d, want %d",
					tt.postgres.calls,
					tt.wantPostgresCalls,
				)
			}

			if tt.redis != nil && tt.redis.calls != tt.wantRedisCalls {
				t.Fatalf(
					"Redis Ping calls = %d, want %d",
					tt.redis.calls,
					tt.wantRedisCalls,
				)
			}
		})
	}
}

func TestNilCheckerFailsClosed(t *testing.T) {
	t.Parallel()

	var checker *Checker

	if err := checker.Check(context.Background()); !errors.Is(err, ErrDependencyMissing) {
		t.Fatalf("Check() error = %v, want %v", err, ErrDependencyMissing)
	}
}
