# Milestone 03 Super Admin Dashboard Shell

Status: implemented application-shell vertical slice, 2026-08-20

This is an implementation and validation note. It does not replace or amend the approved product source-of-truth, architecture, ADRs, or authentication security notes.

## Scope

Milestone 03 turns the authenticated landing page into the reusable Super Admin Control Panel shell and an action-led Dashboard. It adds protected destinations for the complete approved MVP navigation while keeping every unimplemented business module explicit and empty.

The only internal actor remains **Super Admin**. This milestone does not introduce roles, tenant selection, business records, sample metrics, or client-selected identity.

The controlling references are the [source-of-truth precedence](../source-of-truth/README.md), [Super Admin Control Panel MVP Specification](../source-of-truth/AI_Sales_Agent_Super_Admin_CP_MVP_Spec_v1.0.md), [Technical Architecture](../architecture/AI_Sales_Agent_Technical_Architecture_v1.0.md), and the existing [Milestone 01](../security/MILESTONE_01_AUTHENTICATION.md), [Milestone 02](../security/MILESTONE_02_AUTHENTICATION.md), and [Milestone 02B](../security/MILESTONE_02B_EMERGENCY_RECOVERY.md) security notes.

## Application shell and navigation

Authenticated pages share the same shell:

- **Sales Agent** product name and **Super Admin** context;
- the current Super Admin's server-resolved display name and email;
- the current page title;
- a keyboard-accessible Logout button;
- a persistent desktop sidebar and a compact mobile navigation drawer.

The navigation contains exactly these items and routes:

| Navigation item | Route | Milestone 03 state |
| --- | --- | --- |
| Dashboard | `/dashboard` | Implemented Dashboard content |
| Organizations | `/organizations` | Honest module placeholder |
| Applications | `/applications` | Honest module placeholder |
| Packages | `/packages` | Honest module placeholder |
| AI & Usage | `/ai-usage` | Honest module placeholder |
| Integrations | `/integrations` | Honest module placeholder |
| AI Agents | `/ai-agents` | Honest module placeholder |
| System Health | `/system-health` | Honest module placeholder with the current runtime-readiness boundary explained |
| Logs | `/logs` | Honest module placeholder |
| Audit | `/audit` | Honest module placeholder |

Every item is a real protected route. Placeholder pages include a page title, the future module's purpose, a precise not-implemented state, and a path back to Dashboard. They expose no mutation controls and create no records.

## Dashboard hierarchy

The Dashboard deliberately follows operational questions instead of a generic KPI-card layout. Its order is:

1. **Needs Attention**
2. **AI Cost & Consumption**
3. **Organizations**
4. **System Health**
5. **Recent Important Activity**

The header contains a concise operational subtitle and the primary **Create Organization** action.

### Create Organization action

**Create Organization** is a link to `/organizations`. That route explains that Organization management has not been implemented yet. Linking to the honest destination keeps the approved principal action visible and lets the user understand the next workflow without presenting a fake form, disabled control, dead-end modal, or backend mutation.

## Dashboard data contract

The browser loads one protected summary endpoint:

```http
GET /api/v1/admin/dashboard
```

The API returns only authoritative information in the repository's normal success envelope. The summary fields are:

```text
needs_attention
  available: false
  reason: SOURCE_NOT_IMPLEMENTED
  items: []

ai_cost_consumption
  available: false
  reason: COST_TRACKING_NOT_IMPLEMENTED

organizations
  available: false
  reason: ORGANIZATIONS_MODULE_NOT_IMPLEMENTED

system_health
  overall_state: UNKNOWN
  reason: PRODUCT_HEALTH_NOT_IMPLEMENTED
  core_runtime_readiness
    available: true | false
    ready: true | false
    reason: CHECK_SUCCEEDED | CHECK_FAILED | CHECKER_UNAVAILABLE

recent_important_activity
  available: false
  reason: AUDIT_QUERY_NOT_IMPLEMENTED
  items: []
```

Reason values are stable machine-readable explanations, not business measurements. They let the frontend render truthful copy without inferring implementation state from an empty array or a zero.

### Section sources and states

| Dashboard section | Milestone 03 source | User-visible behavior |
| --- | --- | --- |
| Needs Attention | No actionable product-module source exists yet. | States that no actionable platform issues are available from implemented modules. It does not claim that the platform has no issues. |
| AI Cost & Consumption | Agent Run cost tracking is not implemented. | Explains that usage and cost attribution will populate after Agent Run cost tracking exists. It shows no currency, token total, chart, or Strategy Credits. |
| Organizations | Organization persistence and workflows are not implemented. | Explains when Organization data will appear and links the approved action to the Organizations placeholder. It shows no count or sample record. |
| System Health | The existing bounded PostgreSQL-and-Redis core runtime readiness checker. | Product-level state is always `UNKNOWN` in this milestone. Core runtime readiness is shown separately when the check is available; a successful readiness result is never relabelled as overall `HEALTHY`. |
| Recent Important Activity | The reusable Platform Audit store exists, but its protected Dashboard query is intentionally deferred. | Explains that recent activity is unavailable and shows no fabricated event. |

The runtime readiness check uses the same injected checker as `/health/ready`. A successful check reports that the core runtime is ready. A checker failure reports core runtime not ready using safe copy. A missing checker reports readiness unavailable. Dependency names, credentials, connection strings, SQL/Redis errors, and provider topology are never returned.

## UX state model

The protected shell and Dashboard define the following states:

- **Loading:** session validation and Dashboard loading use bounded, understandable status content without continuous announcements.
- **Authenticated:** the shell renders only after `GET /api/v1/auth/session` succeeds.
- **Unauthenticated:** a missing, expired, or revoked session redirects to `/login`.
- **Backend unavailable:** the page presents controlled retry guidance and never displays a raw backend or dependency error.
- **Partial Dashboard data:** section components use their own availability/reason fields, so an unavailable module is described in place instead of becoming a fake zero.
- **Empty:** implemented collections may render a specific empty explanation; currently unavailable collections remain explicitly unavailable.
- **Module not implemented:** protected placeholder pages identify the exact deferred module and link back to Dashboard.
- **Mobile navigation open/closed:** the drawer has an explicit toggle, closes after navigation and by its dismissal control, and does not create horizontal overflow.

The summary is intentionally small. No Dashboard table or additional per-card endpoint was introduced.

## Security model

Frontend route protection improves the experience but is not an authorization boundary.

- `GET /api/v1/auth/session` remains the source of authenticated identity for the shell.
- `GET /api/v1/admin/dashboard` independently resolves the current PostgreSQL-backed session cookie on every request.
- The browser never submits a Super Admin ID, Organization ID, tenant selector, role, or identity query parameter.
- The Dashboard endpoint accepts no query contract. Query strings, including attempted admin or Organization selectors, are rejected after authentication.
- The browser stores no authentication identity or token in `localStorage` or `sessionStorage`. The OTP page's pre-authentication challenge storage remains the narrow Milestone 02 behavior.
- The session secret remains in the host-only, `HttpOnly`, `SameSite=Strict` cookie. Session JSON contains only the established safe session view and CSRF token.
- Logout continues to call the existing server-side endpoint with exact-origin and synchronizer-CSRF validation, revokes the server session, and clears the cookie.
- API responses use the existing safe error envelope and correlation behavior and are marked `Cache-Control: no-store`.
- No wildcard CORS policy, client-trusted role, hidden-URL authorization, raw dependency error, or sensitive credential is introduced.

An attacker may know every route, request, response shape, and frontend branch without gaining an authenticated session or authority.

## Responsive behavior

At desktop widths the sidebar remains visible beside the header and main content. At tablet and mobile widths it becomes a compact drawer controlled by a clearly named menu button. The content grid collapses to a single usable column, long text wraps, controls keep usable target sizes, and the viewport has no horizontal overflow.

Opening the mobile drawer moves focus into its navigation context. The drawer can be dismissed with its close control and the Escape key, and focus returns to the menu trigger. Selecting a destination closes the drawer. Background content is not treated as an alternative navigation surface while the drawer is open.

## Accessibility behavior

- The primary navigation is a semantic, labelled `nav` landmark.
- The current route is exposed with `aria-current="page"` in addition to its visual style.
- Links are used for navigation; buttons are used for menu disclosure, retry, and logout actions.
- Page and section headings preserve a meaningful hierarchy.
- Loading, unavailable, and error copy remains understandable without color or an icon.
- Keyboard focus is visible on links and controls.
- The mobile menu exposes its accessible name and expanded state, supports Escape dismissal, and restores focus.
- Logout, Dashboard return links, and all navigation destinations are reachable by keyboard.
- Repeated loading or polling does not produce excessive live-region announcements.

## Intentionally not implemented

This milestone does **not** implement:

- Organization creation, persistence, lists, detail pages, or lifecycle operations;
- Applications review or approval workflows;
- Packages, numeric limits, or overrides;
- AI usage accounting, cost attribution, charts, or Strategy Credit reporting;
- Organization or core integrations;
- AI Agent registry, versions, testing, or operations;
- product-level System Health, provider checks, job controls, operational issues, or status history;
- Logs or Agent Run inspection;
- Audit list/detail UI or a Dashboard Audit query;
- new database tables or demo business data;
- additional internal actors or role-management UI.

The future System Health module may add AI Runtime, Core Providers, Search/Research, Meta dependencies, Notification Email, Background Jobs, File Processing where applicable, and Agent Operations. None of those checks are implied by core runtime readiness today.

## Automated validation

Backend tests cover session enforcement, logged-out/expired/revoked rejection, rejection of browser-selected identity, safe responses, honest unavailable modules, `UNKNOWN` overall health, real readiness outcomes, failure isolation at the supported boundary, and dependency-error redaction.

Frontend unit tests cover protected-shell identity, the exact navigation, current-page indication, the five Dashboard sections, the Create Organization action, unavailable-state copy without fake zeroes, redirect/logout behavior, placeholders, mobile navigation behavior, and controlled backend-unavailable handling.

The Playwright suite uses the real Go backend and the existing approved authentication fixtures. In OTP mode it completes email/password login, retrieves the real OTP from Mailpit, verifies the code, and then exercises the shell. The explicitly configured local bypass remains a separate local-only mode. The acceptance journey covers:

1. Dashboard shell and exact approved navigation;
2. Organizations placeholder and return to Dashboard;
3. Applications and every remaining protected navigation route without a 404;
4. refresh persistence on a protected route;
5. absence of fake Organization and AI cost values;
6. mobile drawer open, navigation, close, Escape/focus behavior, and horizontal-overflow smoke checks;
7. server-side logout and direct protected-route redirect afterward.

No Dashboard or authentication API is mocked by E2E.

## Beginner-friendly manual browser validation

These steps use the existing real local OTP flow. Do not edit the real `.env` just for this validation; use the repository's already configured local environment or temporary shell overrides approved for your setup.

### Start the services

1. From the repository root, start PostgreSQL, Redis, and Mailpit:

   ```bash
   make infra-up
   ```

2. Apply migrations:

   ```bash
   make migrate-up
   ```

3. If no local Super Admin exists, create one and follow the hidden password prompts:

   ```bash
   make bootstrap-super-admin EMAIL=admin@example.com NAME="Super Admin"
   ```

4. Start the Go API in one terminal:

   ```bash
   make api
   ```

5. Start the web application in a second terminal:

   ```bash
   make web
   ```

### Inspect the desktop shell

1. Open `http://127.0.0.1:3001/login`.
2. Enter the provisioned email and password, then select **Sign in**.
3. Open Mailpit at `http://127.0.0.1:8025`, open the newest message for that email, and copy its six-digit code.
4. Return to the verification page, enter the code, and select **Verify**. Confirm the browser reaches `/dashboard`.
5. Confirm the page shows **Sales Agent**, **Super Admin**, the authenticated name/email, **Dashboard**, and **Logout**.
6. Confirm the navigation contains exactly Dashboard, Organizations, Applications, Packages, AI & Usage, Integrations, AI Agents, System Health, Logs, and Audit. Dashboard should be identified as current.
7. Confirm the five sections appear in the documented order. Verify that no Organization count, AI currency/token total, demo record, chart, or overall `HEALTHY` claim appears.

### Inspect every route and action

1. Select **Create Organization**. Confirm it opens `/organizations` and explains that Organization management is not implemented; no form or record creation should occur.
2. Use the page's Dashboard link to return to `/dashboard`.
3. Open **Applications**, **Packages**, **AI & Usage**, **Integrations**, **AI Agents**, **System Health**, **Logs**, and **Audit** one at a time.
4. On every route, confirm the URL is correct, the page title matches the navigation item, the current navigation item is identified, a precise not-implemented explanation is present, and the page is not a 404.
5. On `/system-health`, confirm the copy distinguishes core runtime readiness from the unimplemented product-level health module and does not label the whole platform `HEALTHY`.

### Inspect refresh, mobile navigation, and logout

1. While on any placeholder route, refresh the browser. Confirm the same protected route returns and the authenticated shell remains visible.
2. Open browser developer tools, enable responsive/device mode, and set a common mobile width such as 390 pixels.
3. Confirm the desktop sidebar is replaced by a named menu button and there is no horizontal page scrollbar.
4. Using the keyboard, focus and activate the menu button. Confirm the navigation opens and focus moves into it.
5. Press Escape. Confirm the menu closes and focus returns to the menu button. Open it again, select a destination, and confirm the destination loads and the menu closes.
6. Return to Dashboard and select **Logout**. Confirm the browser returns to `/login`.
7. Enter `http://127.0.0.1:3001/dashboard` directly. Confirm it redirects to `/login`. Repeat with one placeholder URL such as `/organizations`.

If a dependency is unavailable during the test, the browser should show controlled retry/unavailable guidance, never raw PostgreSQL, Redis, network, or credential text.

## Next recommended visible milestone

The next visible milestone should implement **Organizations**, beginning with the authenticated Organization directory and the real **Create Organization** workflow. That turns the already prominent primary action into an authoritative backend mutation while preserving the shell, route protection, validation, and Audit boundaries established here. Applications, Packages, and the other navigation modules should remain deferred until their own milestones.
