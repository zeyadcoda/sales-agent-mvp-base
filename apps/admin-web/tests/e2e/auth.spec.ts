import { APIRequestContext, expect, Page, test } from "@playwright/test";
import { execFile } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { promisify } from "node:util";

const configuredEmail = process.env.E2E_ADMIN_EMAIL?.trim();
const configuredPassword = process.env.E2E_ADMIN_PASSWORD;
const authenticationMode = process.env.E2E_AUTH_MODE?.trim().toLowerCase() || "otp";
const mailpitOrigin = process.env.E2E_MAILPIT_ORIGIN?.trim() || "http://127.0.0.1:8025";
const e2eDatabaseURL =
  process.env.E2E_DATABASE_URL?.trim() || process.env.TEST_DATABASE_URL?.trim();
const executeFile = promisify(execFile);
const backendDirectory = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../../../backend",
);

const approvedNavigation = [
  { name: "Dashboard", path: "/dashboard" },
  { name: "Organizations", path: "/organizations" },
  { name: "Applications", path: "/applications" },
  { name: "Packages", path: "/packages" },
  { name: "AI & Usage", path: "/ai-usage" },
  { name: "Integrations", path: "/integrations" },
  { name: "AI Agents", path: "/ai-agents" },
  { name: "System Health", path: "/system-health" },
  { name: "Logs", path: "/logs" },
  { name: "Audit", path: "/audit" },
] as const;

const dashboardSectionNames = [
  "Needs Attention",
  "AI Cost & Consumption",
  "Organizations",
  "System Health",
  "Recent Important Activity",
] as const;

test.describe("Super Admin authentication", () => {
  test.describe.configure({ mode: "serial" });

  test("redirects direct dashboard access when logged out", async ({ page }) => {
    await page.goto("/dashboard");

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole("heading", { name: "Super Admin sign in" })).toBeVisible();
  });

  test("redirects direct OTP-page access without a challenge", async ({ page }) => {
    await page.goto("/verify-otp");

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole("heading", { name: "Super Admin sign in" })).toBeVisible();
  });

  test("keeps the user on login and shows a generic error for a wrong password", async ({
    page,
  }) => {
    test.skip(!hasCredentials(), credentialsSkipReason);
    const { email, password } = credentials();
    const definitelyWrongPassword =
      password === "definitely-not-the-valid-password"
        ? "another-definitely-wrong-password"
        : "definitely-not-the-valid-password";

    await page.goto("/login");
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Password").fill(definitelyWrongPassword);
    await page.getByRole("button", { name: "Sign in" }).click();

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.locator("form").getByRole("alert")).toHaveText(
      "The email or password is incorrect.",
    );
    await expect(page.getByLabel("Password")).toHaveValue("");
  });

  test("uses a real Mailpit OTP, resends safely, and exercises the protected application shell", async ({
    page,
    request,
  }) => {
    test.setTimeout(180_000);
    test.skip(authenticationMode !== "otp", "Set E2E_AUTH_MODE=otp for the real OTP flow.");
    test.skip(!hasCredentials(), credentialsSkipReason);
    const { email, password } = credentials();
    const loginStartedAt = new Date();

    await beginSignIn(page, email, password);
    await expect(page).toHaveURL(/\/verify-otp$/);
    await expect(page.getByRole("heading", { name: "Verify your email" })).toBeVisible();

    const firstMessage = await waitForOTPEmail(request, email, loginStartedAt);
    const resendButton = page.getByRole("button", { name: "Resend code" });
    await expect(resendButton).toBeDisabled();
    await expect(page.getByText(/Resend available in/)).toBeVisible();

    const earlyResend = await callResendDirectly(page);
    expect(earlyResend.ok).toBe(false);
    expect(earlyResend.code).toBe("AUTH_OTP_RESEND_TOO_EARLY");

    await page.getByLabel("Six-digit verification code").fill(wrongCodeFor(firstMessage.otp));
    await page.getByRole("button", { name: "Verify" }).click();
    await expect(page.locator("form").getByRole("alert")).toHaveText(
      "The verification code is invalid. Try again.",
    );

    await expect(resendButton).toBeEnabled({ timeout: 70_000 });
    const resendStartedAt = new Date();
    await resendButton.click();
    await expect(page.getByRole("status")).toHaveText("A new verification code has been sent.");

    const secondMessage = await waitForOTPEmail(request, email, resendStartedAt, firstMessage.id);
    expect(secondMessage.id).not.toBe(firstMessage.id);

    await page.getByLabel("Six-digit verification code").fill(firstMessage.otp);
    await page.getByRole("button", { name: "Verify" }).click();
    await expect(page.locator("form").getByRole("alert")).toHaveText(
      "The verification code is invalid. Try again.",
    );

    await page.getByLabel("Six-digit verification code").fill(secondMessage.otp);
    await page.getByRole("button", { name: "Verify" }).click();
    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
    await expect(page.getByText(email, { exact: true }).first()).toBeVisible();

    await exerciseProtectedApplicationShell(page, email);
  });

  test("locks a real challenge after five wrong codes", async ({ page, request }) => {
    test.skip(authenticationMode !== "otp", "Set E2E_AUTH_MODE=otp for the real OTP flow.");
    test.skip(!hasCredentials(), credentialsSkipReason);
    const { email, password } = credentials();
    const loginStartedAt = new Date();

    await beginSignIn(page, email, password);
    await expect(page).toHaveURL(/\/verify-otp$/);
    const message = await waitForOTPEmail(request, email, loginStartedAt);
    const wrongOTP = wrongCodeFor(message.otp);

    for (let attempt = 1; attempt <= 4; attempt += 1) {
      await page.getByLabel("Six-digit verification code").fill(wrongOTP);
      await page.getByRole("button", { name: "Verify" }).click();
      await expect(page.locator("form").getByRole("alert")).toHaveText(
        "The verification code is invalid. Try again.",
      );
    }

    await page.getByLabel("Six-digit verification code").fill(wrongOTP);
    await page.getByRole("button", { name: "Verify" }).click();
    await expect(
      page.getByRole("heading", { name: "Verification attempts exceeded" }),
    ).toBeVisible();
  });

  test("rejects a correct code after the real challenge expires", async ({ page, request }) => {
    test.skip(authenticationMode !== "otp", "Set E2E_AUTH_MODE=otp for the real OTP flow.");
    test.skip(!hasCredentials(), credentialsSkipReason);
    if (!e2eDatabaseURL) {
      throw new Error(
        "E2E_DATABASE_URL is required and must be the dedicated loopback *_integration_test database used by the API.",
      );
    }
    const { email, password } = credentials();
    const loginStartedAt = new Date();

    await beginSignIn(page, email, password);
    await expect(page).toHaveURL(/\/verify-otp$/);
    const message = await waitForOTPEmail(request, email, loginStartedAt);
    const challengeID = await pendingChallengeID(page);

    await expireChallengeFixture(challengeID);

    const verification = await callVerifyDirectly(page, challengeID, message.otp);
    expect(verification.status).toBe(410);
    expect(verification.code).toBe("AUTH_OTP_EXPIRED");

    await page.reload();

    await expect(page.getByRole("heading", { name: "Verification code expired" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Restart login" })).toBeVisible();
  });

  test("preserves the explicitly configured local bypass flow and exercises the protected application shell", async ({
    page,
  }) => {
    test.skip(authenticationMode !== "bypass", "Set E2E_AUTH_MODE=bypass to test local bypass.");
    test.skip(!hasCredentials(), credentialsSkipReason);
    const { email, password } = credentials();

    await beginSignIn(page, email, password);
    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
    await expect(page.getByText("Local development", { exact: true })).toBeVisible();

    await exerciseProtectedApplicationShell(page, email);
  });
});

async function exerciseProtectedApplicationShell(page: Page, email: string): Promise<void> {
  await page.setViewportSize({ width: 1280, height: 900 });
  await expect(page).toHaveURL(/\/dashboard$/);

  const banner = page.getByRole("banner");
  await expect(banner.getByText(email, { exact: true }).first()).toBeVisible();
  await expect(banner.getByText("Super Admin", { exact: true }).first()).toBeVisible();
  const sidebar = page.getByRole("complementary", { name: "Super Admin navigation" });
  await expect(sidebar.getByText("Sales Agent", { exact: true })).toBeVisible();
  await expect(sidebar.getByText("Super Admin", { exact: true })).toBeVisible();

  const main = page.getByRole("main");
  await expect(
    main.getByRole("heading", { level: 1, name: "Dashboard", exact: true }),
  ).toBeVisible();
  await assertApprovedNavigation(page, "Dashboard");
  await expect(page.getByRole("button", { name: "Open navigation" })).toBeHidden();

  const sectionHeadings = main.getByRole("heading", { level: 2 });
  await expect(sectionHeadings).toHaveCount(dashboardSectionNames.length);
  expect((await sectionHeadings.allTextContents()).map((heading) => heading.trim())).toEqual(
    dashboardSectionNames,
  );

  const aiCostSection = main.getByRole("region", {
    name: "AI Cost & Consumption",
  });
  const organizationsSection = main.getByRole("region", { name: "Organizations" });
  const systemHealthSection = main.getByRole("region", { name: "System Health" });
  await expect(aiCostSection).toBeVisible();
  await expect(organizationsSection).toBeVisible();
  await expect(systemHealthSection).toBeVisible();
  await expect(aiCostSection.getByText(/not implemented|will populate/i)).toBeVisible();
  await expect(organizationsSection.getByText(/not implemented|will appear/i)).toBeVisible();
  await expect(systemHealthSection.getByText("UNKNOWN", { exact: true })).toBeVisible();
  await expect(systemHealthSection.getByText("HEALTHY", { exact: true })).toHaveCount(0);
  await expect(aiCostSection.getByText(/\$\s*0(?:\.00)?\b/)).toHaveCount(0);
  await expect(aiCostSection.getByText(/\b0\s+(?:tokens?|credits?)\b/i)).toHaveCount(0);
  await expect(organizationsSection.getByText(/\b0\s+organizations?\b/i)).toHaveCount(0);

  const createOrganization = main.getByRole("link", {
    name: "Create Organization",
    exact: true,
  });
  await expect(createOrganization).toBeVisible();
  await expect(createOrganization).toHaveAttribute("href", "/organizations");
  await createOrganization.click();
  await assertModulePlaceholder(page, "Organizations", "/organizations");

  await page.getByRole("main").getByRole("link", { name: "Back to Dashboard" }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
  await assertApprovedNavigation(page, "Dashboard");

  for (const destination of approvedNavigation.slice(2)) {
    await page
      .getByRole("navigation", { name: "Primary navigation" })
      .getByRole("link", { name: destination.name, exact: true })
      .click();
    await assertModulePlaceholder(page, destination.name, destination.path);
  }

  await page.reload();
  await assertModulePlaceholder(page, "Audit", "/audit");

  await page
    .getByRole("navigation", { name: "Primary navigation" })
    .getByRole("link", { name: "Dashboard", exact: true })
    .click();
  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(
    page.getByRole("main").getByRole("heading", {
      level: 1,
      name: "Dashboard",
      exact: true,
    }),
  ).toBeVisible();

  await exerciseMobileNavigation(page);

  await page.getByRole("button", { name: "Logout", exact: true }).click();
  await expect(page).toHaveURL(/\/login$/);

  await page.goto("/organizations");
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole("heading", { name: "Super Admin sign in" })).toBeVisible();

  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/login$/);
}

async function assertApprovedNavigation(
  page: Page,
  currentPage: (typeof approvedNavigation)[number]["name"],
): Promise<void> {
  const navigation = page.getByRole("navigation", { name: "Primary navigation" });
  await expect(navigation).toBeVisible();

  const links = navigation.getByRole("link");
  await expect(links).toHaveCount(approvedNavigation.length);
  for (const [index, destination] of approvedNavigation.entries()) {
    await expect(links.nth(index)).toHaveAccessibleName(destination.name);
    await expect(links.nth(index)).toHaveAttribute("href", destination.path);
  }

  await expect(
    navigation.getByRole("link", { name: currentPage, exact: true }),
  ).toHaveAttribute("aria-current", "page");
}

async function assertModulePlaceholder(
  page: Page,
  name: (typeof approvedNavigation)[number]["name"],
  pathName: (typeof approvedNavigation)[number]["path"],
): Promise<void> {
  await assertModulePlaceholderContent(page, name, pathName);
  await assertApprovedNavigation(page, name);
}

async function assertModulePlaceholderContent(
  page: Page,
  name: (typeof approvedNavigation)[number]["name"],
  pathName: (typeof approvedNavigation)[number]["path"],
): Promise<void> {
  await expect(page).toHaveURL(new RegExp(`${escapeRegExp(pathName)}$`));
  const main = page.getByRole("main");
  await expect(main.getByRole("heading", { level: 1, name, exact: true })).toBeVisible();
  await expect(
    main.getByText("This module has not been implemented yet.", { exact: true }),
  ).toBeVisible();
  await expect(main.getByRole("link", { name: "Back to Dashboard" })).toBeVisible();
  await expect(main.getByText("404", { exact: true })).toHaveCount(0);
}

async function exerciseMobileNavigation(page: Page): Promise<void> {
  await page.setViewportSize({ width: 390, height: 844 });

  const navigation = page.getByRole("navigation", { name: "Primary navigation" });
  const openNavigation = page.getByRole("button", { name: "Open navigation" });
  await expect(openNavigation).toBeVisible();
  await expect(openNavigation).toHaveAttribute("aria-expanded", "false");
  await expect(navigation).toBeHidden();

  await openNavigation.focus();
  await page.keyboard.press("Enter");
  const navigationDialog = page.getByRole("dialog", { name: "Super Admin navigation" });
  const closeNavigation = navigationDialog.getByRole("button", { name: "Close navigation" });
  await expect(navigationDialog).toBeVisible();
  await expect(navigation).toBeVisible();
  await expect(closeNavigation).toBeFocused();

  await page.keyboard.press("Shift+Tab");
  await expect(
    navigation.getByRole("link", { name: "Audit", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(closeNavigation).toBeFocused();

  await closeNavigation.click();
  await expect(navigation).toBeHidden();
  await expect(openNavigation).toBeFocused();

  await openNavigation.click();
  await expect(navigation).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(navigation).toBeHidden();
  await expect(openNavigation).toBeFocused();

  await openNavigation.click();
  await navigation.getByRole("link", { name: "AI Agents", exact: true }).click();
  await assertModulePlaceholderContent(page, "AI Agents", "/ai-agents");
  await expect(navigation).toBeHidden();
  await expect(openNavigation).toHaveAttribute("aria-expanded", "false");

  await openNavigation.click();
  await assertApprovedNavigation(page, "AI Agents");
  await page.keyboard.press("Escape");
  await expect(navigation).toBeHidden();

  await expectNoHorizontalOverflow(page);

  await page.setViewportSize({ width: 768, height: 1024 });
  await expect(openNavigation).toBeVisible();
  await expect(navigation).toBeHidden();
  await expectNoHorizontalOverflow(page);
}

async function expectNoHorizontalOverflow(page: Page): Promise<void> {
  const hasHorizontalOverflow = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  );
  expect(hasHorizontalOverflow).toBe(false);
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

async function beginSignIn(page: Page, email: string, password: string) {
  await page.goto("/login");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign in" }).click();
}

async function callResendDirectly(
  page: Page,
): Promise<{ ok: boolean; code: string | null }> {
  return page.evaluate(async () => {
    const rawChallenge = window.sessionStorage.getItem("sales_agent_otp_challenge");
    const parsed: unknown = rawChallenge === null ? null : JSON.parse(rawChallenge);
    const challengeID =
      typeof parsed === "object" &&
      parsed !== null &&
      "challenge_id" in parsed &&
      typeof parsed.challenge_id === "string"
        ? parsed.challenge_id
        : "";
    const response = await fetch("/api/v1/auth/otp/resend", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "same-origin",
      body: JSON.stringify({ challenge_id: challengeID }),
    });
    const payload: unknown = await response.json();
    const code =
      typeof payload === "object" &&
      payload !== null &&
      "error" in payload &&
      typeof payload.error === "object" &&
      payload.error !== null &&
      "code" in payload.error &&
      typeof payload.error.code === "string"
        ? payload.error.code
        : null;

    return { ok: response.ok, code };
  });
}

async function callVerifyDirectly(
  page: Page,
  challengeID: string,
  otp: string,
): Promise<{ status: number; code: string | null }> {
  return page.evaluate(
    async ({ challengeID: browserChallengeID, otp: browserOTP }) => {
      const response = await fetch("/api/v1/auth/otp/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "same-origin",
        body: JSON.stringify({ challenge_id: browserChallengeID, otp: browserOTP }),
      });
      const payload: unknown = await response.json();
      const code =
        typeof payload === "object" &&
        payload !== null &&
        "error" in payload &&
        typeof payload.error === "object" &&
        payload.error !== null &&
        "code" in payload.error &&
        typeof payload.error.code === "string"
          ? payload.error.code
          : null;

      return { status: response.status, code };
    },
    { challengeID, otp },
  );
}

async function pendingChallengeID(page: Page): Promise<string> {
  return page.evaluate(() => {
    const rawChallenge = window.sessionStorage.getItem("sales_agent_otp_challenge");
    const parsed: unknown = rawChallenge === null ? null : JSON.parse(rawChallenge);
    if (
      typeof parsed !== "object" ||
      parsed === null ||
      !("challenge_id" in parsed) ||
      typeof parsed.challenge_id !== "string"
    ) {
      throw new Error("The browser has no pending OTP challenge.");
    }

    return parsed.challenge_id;
  });
}

async function expireChallengeFixture(challengeID: string): Promise<void> {
  if (!e2eDatabaseURL) {
    throw new Error("E2E_DATABASE_URL is required for the expiry fixture.");
  }

  // The fixture is a non-HTTP test command guarded to APP_ENV=test and a
  // dedicated loopback database. Authentication is still exercised through
  // the real Go API before and after this controlled server-state transition.
  await executeFile(
    "go",
    ["run", "./cmd/e2e-auth-fixture", "expire-challenge", challengeID],
    {
      cwd: backendDirectory,
      env: {
        ...process.env,
        APP_ENV: "test",
        GOTOOLCHAIN: "local",
        TEST_DATABASE_URL: e2eDatabaseURL,
      },
      timeout: 30_000,
    },
  );
}

interface MailpitMessageSummary {
  ID: string;
  Created: string;
  To: Array<{ Address: string }>;
}

async function waitForOTPEmail(
  request: APIRequestContext,
  recipient: string,
  createdAfter: Date,
  excludedID?: string,
): Promise<{ id: string; otp: string }> {
  const deadline = Date.now() + 15_000;

  while (Date.now() < deadline) {
    const listResponse = await request.get(`${mailpitOrigin}/api/v1/messages?limit=50`);
    if (listResponse.ok()) {
      const payload: unknown = await listResponse.json();
      const messages = readMailpitMessages(payload)
        .filter(
          (message) =>
            message.ID !== excludedID &&
            message.To.some(
              (mailbox) => mailbox.Address.toLowerCase() === recipient.toLowerCase(),
            ) &&
            Date.parse(message.Created) >= createdAfter.getTime() - 1_000,
        )
        .sort((left, right) => Date.parse(right.Created) - Date.parse(left.Created));

      for (const message of messages) {
        const messageResponse = await request.get(
          `${mailpitOrigin}/api/v1/message/${encodeURIComponent(message.ID)}`,
        );
        if (!messageResponse.ok()) {
          continue;
        }

        const messagePayload: unknown = await messageResponse.json();
        const text = readMailpitText(messagePayload);
        const match = text.match(/(?:^|\D)(\d{6})(?!\d)/m);
        if (match?.[1]) {
          return { id: message.ID, otp: match[1] };
        }
      }
    }

    await new Promise((resolve) => setTimeout(resolve, 250));
  }

  throw new Error(`No recent OTP email appeared in Mailpit for ${recipient}`);
}

function readMailpitMessages(payload: unknown): MailpitMessageSummary[] {
  if (
    typeof payload !== "object" ||
    payload === null ||
    !("messages" in payload) ||
    !Array.isArray(payload.messages)
  ) {
    return [];
  }

  return payload.messages.filter((message): message is MailpitMessageSummary => {
    if (typeof message !== "object" || message === null) {
      return false;
    }

    return (
      "ID" in message &&
      typeof message.ID === "string" &&
      "Created" in message &&
      typeof message.Created === "string" &&
      "To" in message &&
      Array.isArray(message.To) &&
      message.To.every(
        (mailbox: unknown) =>
          typeof mailbox === "object" &&
          mailbox !== null &&
          "Address" in mailbox &&
          typeof mailbox.Address === "string",
      )
    );
  });
}

function readMailpitText(payload: unknown): string {
  if (
    typeof payload === "object" &&
    payload !== null &&
    "Text" in payload &&
    typeof payload.Text === "string"
  ) {
    return payload.Text;
  }

  return "";
}

function wrongCodeFor(correctOTP: string): string {
  return correctOTP === "000000" ? "111111" : "000000";
}

function hasCredentials(): boolean {
  return Boolean(configuredEmail && configuredPassword);
}

function credentials(): { email: string; password: string } {
  if (!configuredEmail || !configuredPassword) {
    throw new Error(credentialsSkipReason);
  }

  return { email: configuredEmail, password: configuredPassword };
}

const credentialsSkipReason =
  "Set E2E_ADMIN_EMAIL and E2E_ADMIN_PASSWORD for the provisioned local Super Admin.";
