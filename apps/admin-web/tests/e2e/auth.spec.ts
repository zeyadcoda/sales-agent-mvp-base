import { expect, Page, test } from "@playwright/test";

const configuredEmail = process.env.E2E_ADMIN_EMAIL?.trim();
const configuredPassword = process.env.E2E_ADMIN_PASSWORD;

test.describe("Super Admin authentication", () => {
  test.describe.configure({ mode: "serial" });
  test.skip(
    !configuredEmail || !configuredPassword,
    "Set E2E_ADMIN_EMAIL and E2E_ADMIN_PASSWORD for the provisioned local Super Admin.",
  );

  test("redirects direct dashboard access when logged out", async ({ page }) => {
    await page.goto("/dashboard");

    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole("heading", { name: "Super Admin sign in" })).toBeVisible();
  });

  test("keeps the user on login and shows a generic error for a wrong password", async ({ page }) => {
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
  });

  test("logs in, survives refresh, logs out, and rejects later dashboard access", async ({ page }) => {
    const { email, password } = credentials();

    await signIn(page, email, password);
    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
    await expect(page.getByText(email, { exact: true })).toBeVisible();
    await expect(page.getByText("Local development", { exact: true })).toBeVisible();

    await page.reload();
    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
    await expect(page.getByText(email, { exact: true })).toBeVisible();

    await page.getByRole("button", { name: "Logout" }).click();
    await expect(page).toHaveURL(/\/login$/);

    await page.goto("/dashboard");
    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole("heading", { name: "Super Admin sign in" })).toBeVisible();
  });
});

async function signIn(page: Page, email: string, password: string) {
  await page.goto("/login");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
}

function credentials(): { email: string; password: string } {
  if (!configuredEmail || !configuredPassword) {
    throw new Error("E2E credentials are not configured");
  }

  return { email: configuredEmail, password: configuredPassword };
}
