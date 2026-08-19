package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"salesagent.local/backend/internal/database"
	"salesagent.local/backend/internal/platform/auth"
)

const databaseOperationTimeout = 10 * time.Second

type passwordHasher interface {
	Hash(password string) (string, error)
}

type bootstrapStore interface {
	auth.ProvisioningStore
	FindSuperAdminByEmail(ctx context.Context, normalizedEmail string) (auth.SuperAdmin, error)
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdin, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap Super Admin: %v\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	stdin *os.File,
	stderr io.Writer,
) error {
	flags := flag.NewFlagSet("bootstrap-super-admin", flag.ContinueOnError)
	flags.SetOutput(stderr)

	emailFlag := flags.String("email", "", "email address for the Super Admin")
	nameFlag := flags.String("name", "Super Admin", "display name for the Super Admin")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}

	normalizedEmail, err := auth.NormalizeEmail(*emailFlag)
	if err != nil {
		return errors.New("--email must be a valid email address")
	}
	displayName, err := auth.ValidateDisplayName(*nameFlag)
	if err != nil {
		return errors.New("--name must be between 1 and 100 characters")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	openCtx, cancelOpen := databaseOperationContext(ctx)
	db, err := database.Open(openCtx, databaseURL)
	cancelOpen()
	if err != nil {
		return errors.New("could not initialize PostgreSQL")
	}
	defer db.Close()

	store, err := auth.NewPostgresStore(db)
	if err != nil {
		return errors.New("could not initialize the Super Admin repository")
	}

	lookupCtx, cancelLookup := databaseOperationContext(ctx)
	_, lookupErr := store.FindSuperAdminByEmail(lookupCtx, normalizedEmail)
	cancelLookup()
	if lookupErr == nil {
		return fmt.Errorf("Super Admin %q already exists; no changes were made", normalizedEmail)
	} else if !errors.Is(lookupErr, auth.ErrSuperAdminNotFound) {
		return errors.New("could not check whether the Super Admin already exists")
	}

	// Refuse redirected/non-terminal input: accepting a password through a
	// pipe makes it much easier for secrets to leak through scripts or process
	// tooling. The password is never accepted as a command-line flag.
	if !term.IsTerminal(int(stdin.Fd())) {
		return errors.New("a secure interactive terminal is required for password entry")
	}

	password, err := readSecret(stdin, stderr, "Password: ")
	if err != nil {
		return err
	}
	defer zero(password)

	confirmation, err := readSecret(stdin, stderr, "Confirm password: ")
	if err != nil {
		return err
	}
	defer zero(confirmation)

	hasher := auth.NewPasswordHasher()
	created, err := provision(
		ctx,
		store,
		hasher,
		normalizedEmail,
		displayName,
		string(password),
		string(confirmation),
	)
	if errors.Is(err, auth.ErrSuperAdminExists) {
		return fmt.Errorf("Super Admin %q already exists; no changes were made", normalizedEmail)
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(stderr, "Super Admin %q (%s) created successfully.\n", created.DisplayName, created.Email)
	return nil
}

func readSecret(stdin *os.File, output io.Writer, prompt string) ([]byte, error) {
	fmt.Fprint(output, prompt)
	secret, err := term.ReadPassword(int(stdin.Fd()))
	fmt.Fprintln(output)
	if err != nil {
		return nil, errors.New("could not read password securely")
	}

	return secret, nil
}

func provision(
	ctx context.Context,
	store auth.ProvisioningStore,
	hasher passwordHasher,
	email string,
	displayName string,
	password string,
	confirmation string,
) (auth.SuperAdmin, error) {
	if password != confirmation {
		return auth.SuperAdmin{}, errors.New("password confirmation does not match")
	}
	if err := auth.ValidateBootstrapPassword(password); err != nil {
		return auth.SuperAdmin{}, errors.New("password must contain between 12 and 128 characters")
	}

	passwordHash, err := hasher.Hash(password)
	if err != nil {
		return auth.SuperAdmin{}, errors.New("could not hash the password securely")
	}

	// Password entry and hashing are deliberately outside this deadline. Each
	// database operation receives a fresh timeout immediately before I/O, so a
	// human taking time at the prompts cannot expire the eventual insert.
	createCtx, cancelCreate := databaseOperationContext(ctx)
	defer cancelCreate()
	created, err := store.CreateSuperAdmin(createCtx, auth.NewSuperAdmin{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		IsActive:     true,
	})
	if err != nil {
		if errors.Is(err, auth.ErrSuperAdminExists) {
			return auth.SuperAdmin{}, auth.ErrSuperAdminExists
		}

		return auth.SuperAdmin{}, errors.New("could not create the Super Admin")
	}

	return created, nil
}

func databaseOperationContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, databaseOperationTimeout)
}

func zero(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
