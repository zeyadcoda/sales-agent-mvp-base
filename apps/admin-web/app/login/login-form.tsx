"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { getLoginErrorMessage, login } from "../../lib/auth-api";

interface FieldErrors {
  email?: string;
  password?: string;
}

export function LoginForm() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const normalizedEmail = email.trim();
    const errors = validateLogin(normalizedEmail, password);
    setFieldErrors(errors);
    setFormError(null);

    if (Object.keys(errors).length > 0) {
      return;
    }

    setIsSubmitting(true);

    try {
      await login({ email: normalizedEmail, password });
      router.replace("/dashboard");
    } catch (error) {
      setFormError(getLoginErrorMessage(error));
      setIsSubmitting(false);
    }
  }

  return (
    <form className="auth-form" onSubmit={handleSubmit} noValidate aria-busy={isSubmitting}>
      {formError ? (
        <div className="alert alert-error" role="alert">
          {formError}
        </div>
      ) : null}

      <div className="field-group">
        <label htmlFor="email">Email</label>
        <input
          id="email"
          name="email"
          type="email"
          inputMode="email"
          autoComplete="username"
          maxLength={254}
          value={email}
          disabled={isSubmitting}
          aria-invalid={fieldErrors.email ? true : undefined}
          aria-describedby={fieldErrors.email ? "email-error" : undefined}
          onChange={(event) => {
            setEmail(event.target.value);
            if (fieldErrors.email) {
              setFieldErrors((current) => ({ ...current, email: undefined }));
            }
          }}
        />
        {fieldErrors.email ? (
          <p className="field-error" id="email-error">
            {fieldErrors.email}
          </p>
        ) : null}
      </div>

      <div className="field-group">
        <label htmlFor="password">Password</label>
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          maxLength={1024}
          value={password}
          disabled={isSubmitting}
          aria-invalid={fieldErrors.password ? true : undefined}
          aria-describedby={fieldErrors.password ? "password-error" : undefined}
          onChange={(event) => {
            setPassword(event.target.value);
            if (fieldErrors.password) {
              setFieldErrors((current) => ({ ...current, password: undefined }));
            }
          }}
        />
        {fieldErrors.password ? (
          <p className="field-error" id="password-error">
            {fieldErrors.password}
          </p>
        ) : null}
      </div>

      <button className="button button-primary button-full" type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Signing in…" : "Sign in"}
      </button>
    </form>
  );
}

function validateLogin(email: string, password: string): FieldErrors {
  const errors: FieldErrors = {};

  if (email === "") {
    errors.email = "Enter your email address.";
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    errors.email = "Enter a valid email address.";
  }

  if (password === "") {
    errors.password = "Enter your password.";
  }

  return errors;
}
