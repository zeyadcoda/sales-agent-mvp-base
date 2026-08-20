package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"salesagent.local/backend/internal/database"
	"salesagent.local/backend/internal/platform/audit"
	"salesagent.local/backend/internal/platform/auth"
	"salesagent.local/backend/internal/requestmeta"
)

const (
	platformAdminDatabaseTimeout = 10 * time.Second
	recoveryAuditListLimit       = 50
)

type recoveryAdministrator interface {
	Authorize(
		ctx context.Context,
		email string,
		reason string,
		operatorIdentifier string,
	) (auth.RecoveryAuthorization, error)
	Status(ctx context.Context, email string) (auth.RecoveryAuthorizationStatus, error)
	Revoke(
		ctx context.Context,
		email string,
		reason string,
		operatorIdentifier string,
	) (auth.RecoveryAuthorization, error)
}

type recoveryAuditReader interface {
	ListByResource(
		ctx context.Context,
		resourceType audit.ResourceType,
		resourceReference string,
		limit int,
	) ([]audit.Event, error)
}

type platformAdminServices struct {
	recovery recoveryAdministrator
	audit    recoveryAuditReader
	close    func()
}

type platformAdminServiceFactory func(ctx context.Context) (platformAdminServices, error)

func main() {
	if err := run(
		context.Background(),
		os.Args[1:],
		os.Stdin,
		os.Stdout,
		defaultPlatformAdminServices,
	); err != nil {
		fmt.Fprintf(os.Stderr, "platform-admin: %v\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	stdin io.Reader,
	output io.Writer,
	servicesFactory platformAdminServiceFactory,
) error {
	if len(args) < 2 || args[0] != "auth-recovery" {
		return errors.New("usage: platform-admin auth-recovery <authorize|status|revoke|audit> [options]")
	}
	if stdin == nil || output == nil || servicesFactory == nil {
		return errors.New("platform-admin command dependencies are required")
	}

	switch args[1] {
	case "authorize":
		return runRecoveryAuthorize(ctx, args[2:], stdin, output, servicesFactory)
	case "status":
		return runRecoveryStatus(ctx, args[2:], output, servicesFactory)
	case "revoke":
		return runRecoveryRevoke(ctx, args[2:], stdin, output, servicesFactory)
	case "audit":
		return runRecoveryAudit(ctx, args[2:], output, servicesFactory)
	default:
		return errors.New("auth-recovery command must be authorize, status, revoke, or audit")
	}
}

func runRecoveryAuthorize(
	ctx context.Context,
	args []string,
	stdin io.Reader,
	output io.Writer,
	servicesFactory platformAdminServiceFactory,
) error {
	flags := flag.NewFlagSet("auth-recovery authorize", flag.ContinueOnError)
	flags.SetOutput(output)
	emailFlag := flags.String("email", "", "exact Super Admin email")
	reasonFlag := flags.String("reason", "", "mandatory emergency recovery reason")
	operatorFlag := flags.String("operator", "", "deployment/operator identifier or label")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}

	email, err := normalizedRecoveryEmail(*emailFlag)
	if err != nil {
		return err
	}
	reason, err := auth.NormalizeRecoveryReason(*reasonFlag)
	if err != nil {
		return errors.New("--reason is required and must be a single line of at most 500 characters")
	}
	operator, err := auth.NormalizeRecoveryOperator(*operatorFlag)
	if err != nil {
		return errors.New("--operator is required and must be a single line of at most 200 characters")
	}

	fmt.Fprintln(output, "Emergency Super Admin recovery authorization")
	fmt.Fprintf(output, "Target Super Admin: %s\n", email)
	fmt.Fprintf(output, "Reason: %s\n", reason)
	fmt.Fprintf(output, "Deployment operator: %s\n", operator)
	fmt.Fprintln(output, "Recovery validity: 10 minutes")
	fmt.Fprintln(output, "Behavior: next successful password login bypasses email OTP once")
	fmt.Fprintln(output, "No recovery code, token, or browser secret will be created.")
	confirmed, err := confirmOperation(stdin, output)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(output, "Cancelled; no recovery authorization was created.")
		return nil
	}

	correlationID, err := newOperatorCorrelationID()
	if err != nil {
		return errors.New("could not create a safe operation correlation ID")
	}
	operationCtx := requestmeta.WithCorrelationID(ctx, correlationID)
	services, err := servicesFactory(operationCtx)
	if err != nil {
		return err
	}
	defer closePlatformAdminServices(services)
	if services.recovery == nil {
		return errors.New("recovery administration service is unavailable")
	}

	operationCtx, cancel := context.WithTimeout(operationCtx, platformAdminDatabaseTimeout)
	defer cancel()
	authorization, err := services.recovery.Authorize(operationCtx, email, reason, operator)
	if err != nil {
		return recoveryAuthorizeCommandError(err)
	}

	fmt.Fprintf(
		output,
		"Recovery authorization created for %s; it expires at %s.\n",
		authorization.SuperAdminEmail,
		authorization.ExpiresAt.UTC().Format(time.RFC3339),
	)
	fmt.Fprintf(output, "Audit correlation ID: %s\n", correlationID)
	return nil
}

func runRecoveryStatus(
	ctx context.Context,
	args []string,
	output io.Writer,
	servicesFactory platformAdminServiceFactory,
) error {
	flags := flag.NewFlagSet("auth-recovery status", flag.ContinueOnError)
	flags.SetOutput(output)
	emailFlag := flags.String("email", "", "exact Super Admin email")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	email, err := normalizedRecoveryEmail(*emailFlag)
	if err != nil {
		return err
	}

	services, err := servicesFactory(ctx)
	if err != nil {
		return err
	}
	defer closePlatformAdminServices(services)
	if services.recovery == nil {
		return errors.New("recovery administration service is unavailable")
	}

	operationCtx, cancel := context.WithTimeout(ctx, platformAdminDatabaseTimeout)
	defer cancel()
	status, err := services.recovery.Status(operationCtx, email)
	if err != nil {
		return recoveryStatusCommandError(err)
	}

	fmt.Fprintf(output, "Target Super Admin: %s\n", email)
	fmt.Fprintf(output, "Recovery status: %s\n", status.State)
	if status.State == auth.RecoveryAuthorizationStateNone {
		fmt.Fprintln(output, "No recovery authorization history exists for this Super Admin.")
		return nil
	}

	authorization := status.Authorization
	fmt.Fprintf(output, "Created at: %s\n", authorization.CreatedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(output, "Expires at: %s\n", authorization.ExpiresAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(output, "Authorized by: %s\n", authorization.OperatorIdentifier)
	fmt.Fprintf(output, "Reason: %s\n", authorization.Reason)
	return nil
}

func runRecoveryRevoke(
	ctx context.Context,
	args []string,
	stdin io.Reader,
	output io.Writer,
	servicesFactory platformAdminServiceFactory,
) error {
	flags := flag.NewFlagSet("auth-recovery revoke", flag.ContinueOnError)
	flags.SetOutput(output)
	emailFlag := flags.String("email", "", "exact Super Admin email")
	reasonFlag := flags.String("reason", "", "mandatory revocation reason")
	operatorFlag := flags.String("operator", "", "deployment/operator identifier or label")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}

	email, err := normalizedRecoveryEmail(*emailFlag)
	if err != nil {
		return err
	}
	reason, err := auth.NormalizeRecoveryReason(*reasonFlag)
	if err != nil {
		return errors.New("--reason is required and must be a single line of at most 500 characters")
	}
	operator, err := auth.NormalizeRecoveryOperator(*operatorFlag)
	if err != nil {
		return errors.New("--operator is required and must be a single line of at most 200 characters")
	}

	fmt.Fprintln(output, "Revoke Super Admin recovery authorization")
	fmt.Fprintf(output, "Target Super Admin: %s\n", email)
	fmt.Fprintf(output, "Reason: %s\n", reason)
	fmt.Fprintf(output, "Deployment operator: %s\n", operator)
	fmt.Fprintln(output, "Behavior: the current active authorization becomes immediately unusable")
	confirmed, err := confirmOperation(stdin, output)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(output, "Cancelled; no recovery authorization was revoked.")
		return nil
	}

	correlationID, err := newOperatorCorrelationID()
	if err != nil {
		return errors.New("could not create a safe operation correlation ID")
	}
	operationCtx := requestmeta.WithCorrelationID(ctx, correlationID)
	services, err := servicesFactory(operationCtx)
	if err != nil {
		return err
	}
	defer closePlatformAdminServices(services)
	if services.recovery == nil {
		return errors.New("recovery administration service is unavailable")
	}

	operationCtx, cancel := context.WithTimeout(operationCtx, platformAdminDatabaseTimeout)
	defer cancel()
	if _, err := services.recovery.Revoke(operationCtx, email, reason, operator); err != nil {
		return recoveryRevokeCommandError(err)
	}

	fmt.Fprintf(output, "Active recovery authorization for %s was revoked.\n", email)
	fmt.Fprintf(output, "Audit correlation ID: %s\n", correlationID)
	return nil
}

func runRecoveryAudit(
	ctx context.Context,
	args []string,
	output io.Writer,
	servicesFactory platformAdminServiceFactory,
) error {
	flags := flag.NewFlagSet("auth-recovery audit", flag.ContinueOnError)
	flags.SetOutput(output)
	emailFlag := flags.String("email", "", "exact Super Admin email")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	email, err := normalizedRecoveryEmail(*emailFlag)
	if err != nil {
		return err
	}

	services, err := servicesFactory(ctx)
	if err != nil {
		return err
	}
	defer closePlatformAdminServices(services)
	if services.audit == nil {
		return errors.New("platform audit reader is unavailable")
	}

	operationCtx, cancel := context.WithTimeout(ctx, platformAdminDatabaseTimeout)
	defer cancel()
	events, err := services.audit.ListByResource(
		operationCtx,
		audit.ResourceTypeSuperAdminAccount,
		email,
		recoveryAuditListLimit,
	)
	if err != nil {
		return errors.New("could not read recovery Audit Events")
	}

	fmt.Fprintf(output, "Recovery Audit Events for %s (newest first):\n", email)
	if len(events) == 0 {
		fmt.Fprintln(output, "No recovery Audit Events found.")
		return nil
	}
	for _, event := range events {
		reason := "-"
		if event.Reason != nil {
			reason = *event.Reason
		}
		fmt.Fprintf(
			output,
			"%s | %s | %s | actor=%s:%s | reason=%s | correlation=%s\n",
			event.OccurredAt.UTC().Format(time.RFC3339),
			event.Action,
			event.Result,
			event.ActorType,
			event.ActorIdentifier,
			reason,
			event.CorrelationID,
		)
	}

	return nil
}

func defaultPlatformAdminServices(ctx context.Context) (platformAdminServices, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return platformAdminServices{}, errors.New("DATABASE_URL is required")
	}

	openCtx, cancel := context.WithTimeout(ctx, platformAdminDatabaseTimeout)
	defer cancel()
	db, err := database.Open(openCtx, databaseURL)
	if err != nil {
		return platformAdminServices{}, errors.New("could not initialize PostgreSQL")
	}

	authStore, err := auth.NewPostgresStore(db)
	if err != nil {
		db.Close()
		return platformAdminServices{}, errors.New("could not initialize recovery persistence")
	}
	recoveryService, err := auth.NewRecoveryService(authStore, auth.RecoveryServiceOptions{})
	if err != nil {
		db.Close()
		return platformAdminServices{}, errors.New("could not initialize recovery administration")
	}
	auditStore, err := audit.NewPostgresStore(db)
	if err != nil {
		db.Close()
		return platformAdminServices{}, errors.New("could not initialize Platform Audit")
	}

	return platformAdminServices{
		recovery: recoveryService,
		audit:    auditStore,
		close:    db.Close,
	}, nil
}

func confirmOperation(stdin io.Reader, output io.Writer) (bool, error) {
	fmt.Fprint(output, "Confirm? [y/N] ")
	answer, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, errors.New("could not read confirmation")
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func normalizedRecoveryEmail(value string) (string, error) {
	normalized, err := auth.NormalizeEmail(value)
	if err != nil {
		return "", errors.New("--email must be an exact valid email address")
	}
	return normalized, nil
}

func recoveryAuthorizeCommandError(err error) error {
	switch {
	case errors.Is(err, auth.ErrRecoveryTargetNotEligible):
		return errors.New("target Super Admin was not found or is not eligible; no authorization was created")
	case errors.Is(err, auth.ErrRecoveryAlreadyActive):
		return errors.New("an active recovery authorization already exists; inspect status or revoke it first")
	default:
		return errors.New("could not create the recovery authorization")
	}
}

func recoveryStatusCommandError(err error) error {
	if errors.Is(err, auth.ErrRecoveryTargetNotEligible) {
		return errors.New("target Super Admin was not found or is not eligible")
	}
	return errors.New("could not read the recovery authorization status")
}

func recoveryRevokeCommandError(err error) error {
	switch {
	case errors.Is(err, auth.ErrRecoveryTargetNotEligible):
		return errors.New("target Super Admin was not found or is not eligible; no authorization was revoked")
	case errors.Is(err, auth.ErrRecoveryNotActive):
		return errors.New("no active recovery authorization exists; nothing was revoked")
	default:
		return errors.New("could not revoke the recovery authorization")
	}
}

func newOperatorCorrelationID() (string, error) {
	material := make([]byte, 16)
	if _, err := rand.Read(material); err != nil {
		return "", err
	}
	return hex.EncodeToString(material), nil
}

func closePlatformAdminServices(services platformAdminServices) {
	if services.close != nil {
		services.close()
	}
}
