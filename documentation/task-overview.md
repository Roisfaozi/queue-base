# QMS-Reborn Reverse Engineering & Refactor Improvement Document

## 1. Project Identity

**Project name:** `qms-reborn`
**Type:** Go backend API
**Domain:** Hospital / clinic queue-management system
**Core stack:** Go, Echo, GORM, MySQL, JWT, Casbin RBAC, Medifrans/Bithealth-style integrations.

The system manages patient queues, users, roles, permissions, menus, submenus, branches, companies, lockets/counters, devices, assets, dashboard data, scanner station check-ins, and appointment-related integration flows. The project uses a layered feature structure with `presentation`, `business`, and `data` packages, wired through a factory layer.

The reverse-engineering approach follows the uploaded RE process: identify the subject, structure, context, expert roles, standards, and final optimized output rather than producing a shallow summary.

---

## 2. High-Level Architecture

```text
HTTP Request
   ↓
Echo Router
   ↓
Global Middleware
- Recover
- CORS
- Logger
- Timeout
- JWT Middleware
- Casbin Middleware
   ↓
Presentation Layer
   ↓
Business Layer
   ↓
Data / Repository Layer
   ↓
GORM + MySQL
   ↓
External APIs
- Medifrans
- Bithealth / Doctor Search
- Notification API
```

The Medifrans integration exposes doctor, schedule, specialty, appointment, patient, scanner, station menu, visit journey, and patient queue routes under `/medifrans`. Scanner login uses `x-client-id` and `x-api-key`, while scanner check-in routes use scanner JWT middleware.

---

## 3. Main Bad Logic / Bad Flow Found

### 3.1 Casbin authorization path matcher is risky

Current Casbin policy files mix path styles such as:

```csv
/branch/*
/branch/:id
/customer/analytic*
```

But the model uses `keyMatch`, which is not ideal for `:id` style route parameters. This can cause authorization bugs where valid routes are denied or mismatched.

**Fix:**

```conf
# File: config/casbin_auth_model.conf

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
```

Then standardize all policies to one style:

```csv
p, 3, /branch/:id, GET
```

or:

```csv
p, 3, /branch/*, GET
```

Do not mix route styles unless the matcher supports them.

---

### 3.2 CORS allows everything with credentials

Current route setup uses:

```go
AllowOrigins: []string{"*"}
AllowCredentials: true
```

This is bad security flow for authenticated APIs because credentialed requests should only be allowed from trusted origins.

**Improved flow:**

```go
// File: routes/routes.go

func getAllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:3000"}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))

	for _, origin := range parts {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}
```

Then:

```go
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	AllowOrigins:     getAllowedOrigins(),
	AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
	AllowCredentials: true,
}))
```

---

### 3.3 Scanner auth middleware only checks presence, not validity

`NewAuthScannerMiddleware` only checks that `x-client-id` and `x-api-key` exist, then stores them in context. It does not validate whether the API key belongs to the client.

**Bad flow:**

```text
Client sends any x-client-id + any x-api-key
   ↓
Middleware accepts presence
   ↓
Request continues
```

**Better flow:**

```text
Client sends x-client-id + x-api-key
   ↓
Middleware checks DB
   ↓
Compare API key securely
   ↓
Reject invalid pair before handler
```

Refactor the middleware to accept a scanner repository:

```go
func NewAuthScannerMiddleware(repo model.ScannerRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			clientID := c.Request().Header.Get("x-client-id")
			apiKey := c.Request().Header.Get("x-api-key")

			if clientID == "" || apiKey == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"status":  "Unauthorized",
					"message": "missing scanner credentials",
				})
			}

			client, err := repo.GetClientByDeviceID(ctx, clientID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"status":  "Unauthorized",
					"message": "invalid scanner credentials",
				})
			}

			if client.ApiKey != apiKey {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"status":  "Unauthorized",
					"message": "invalid scanner credentials",
				})
			}

			c.Set("client_id", clientID)
			c.Set("api_key", apiKey)
			c.Set("branch_id", client.BranchID)

			return next(c)
		}
	}
}
```

The project already has a repository method that can retrieve scanner clients by device ID.

---

### 3.4 Role permission update duplicates Casbin policies

The role update logic appends the same policy object twice:

```go
casbin.Policy = append(casbin.Policy, obj)
casbin.Policy = append(casbin.Policy, obj)
```

and does similar duplication for `LastPolicy`.

**Fix:**

Create a deduplication helper before calling Casbin:

```go
func uniquePolicies(input []middleware.PolicyPath) []middleware.PolicyPath {
	seen := map[string]bool{}
	output := make([]middleware.PolicyPath, 0, len(input))

	for _, p := range input {
		key := p.Path + "|" + p.Method
		if seen[key] {
			continue
		}
		seen[key] = true
		output = append(output, p)
	}

	return output
}
```

Then:

```go
casbin.Policy = uniquePolicies(casbin.Policy)
casbin.LastPolicy = uniquePolicies(casbin.LastPolicy)
```

---

### 3.5 RegisterUser ignores Casbin registration error

The user registration flow calls:

```go
middleware.RegisterUser(casbin)
if err != nil {
    ...
}
```

But the return value from `middleware.RegisterUser(casbin)` is not assigned to `err`, so the error check is ineffective.

**Fix:**

```go
if err := middleware.RegisterUser(casbin); err != nil {
	config.Errorf("[RegisterUser] Error registering user in Casbin: %v", err)
	return users.Core{}, err
}
```

---

### 3.6 Pharmacy validator logic is too fragile

The pharmacy rule checks menu name using this regex:

```go
regexp.MustCompile(`(?i)resep`)
```

That means any menu containing “resep” may pass, not only “Penerimaan Resep”.

**Better logic:**

Add a stable database flag or code:

```text
menus.code = "PENERIMAAN_RESEP"
menus.is_pharmacy_reception = true
```

Then validate using code/flag instead of menu name.

**Better validator condition:**

```go
func isPenerimaanResep(menu MenuInfo) bool {
	return menu.Code == "PENERIMAAN_RESEP" || menu.IsPharmacyReception
}
```

This avoids accidental approval from names like `Cetak Resep`, `Review Resep`, or `Riwayat Resep`.

---

### 3.7 Queue estimate is fake

`GetQueueEstimate` currently returns `1` no matter what:

```go
return 1, nil
```

That is bad business logic because users receive inaccurate estimated wait time.

**Better flow:**

```text
Get queue estimate
   ↓
Count waiting customers before current queue
   ↓
Calculate historical average service duration for menu
   ↓
Multiply queue-left by average duration
   ↓
Return estimate in minutes
```

Suggested implementation:

```go
func (ub *CustomerBusiness) GetQueueEstimate(menuID uint, isPharmacy bool, prescriptionType string) (int, error) {
	avgDuration, err := ub.conferenceData.GetAverageServiceDuration(menuID)
	if err != nil {
		return 0, err
	}

	queueLeft, err := ub.conferenceData.GetQueueLeft(menuID)
	if err != nil {
		return 0, err
	}

	estimate := int(avgDuration.Minutes()) * queueLeft
	if estimate < 1 {
		estimate = 1
	}

	return estimate, nil
}
```

---

### 3.8 Order parameter can become unsafe

The customer pagination logic builds order using string concatenation:

```go
query = query.Order("customers.id " + order)
```

This should be restricted to `ASC` or `DESC`.

**Fix:**

```go
func normalizeOrder(order string) string {
	switch strings.ToUpper(order) {
	case "ASC":
		return "ASC"
	case "DESC":
		return "DESC"
	default:
		return "ASC"
	}
}
```

Then:

```go
query = query.Order("customers.id " + normalizeOrder(order))
```

---

### 3.9 Debug DB stats route should be protected

The project exposes `/debug/dbstats`, returning SQL connection pool stats. This should not be public in production.

**Fix:**

```go
func RegisterDebugRoutes(e *echo.Echo, jwtMiddleware echo.MiddlewareFunc) {
	debug := e.Group("/debug")
	debug.Use(jwtMiddleware)

	debug.GET("/dbstats", func(c echo.Context) error {
		sqlDB, err := config.DB.DB()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "database stats unavailable",
			})
		}

		return c.JSON(http.StatusOK, sqlDB.Stats())
	})
}
```

Or disable it unless:

```env
APP_ENV=development
```

---

## 4. Refactor Plan

### Phase 1 — Safe fixes

1. Fix Casbin matcher to `keyMatch2`.
2. Remove duplicate Casbin policy appends.
3. Capture and handle `RegisterUser` Casbin errors.
4. Add order whitelist for pagination.
5. Protect or disable `/debug/dbstats`.
6. Remove logging of encryption key length from JWT creation.
7. Replace raw internal error messages with safe API messages.

---

### Phase 2 — Security hardening

1. Replace wildcard CORS with environment allowlist.
2. Validate scanner API credentials before scanner login.
3. Require production startup validation for:
   - `JWT_SECRET`
   - `KEY`
   - DB config
   - Casbin model path
   - Casbin policy path
   - external API URLs and keys

4. Add file upload validation:
   - extension whitelist
   - MIME sniffing
   - max size
   - path sanitization

5. Add structured error response helper.

---

### Phase 3 — Bad flow cleanup

Rename confusing internal variables:

| Current                | Better                  |
| ---------------------- | ----------------------- |
| `conferenceData`       | `repo`                  |
| `custommer`            | `customerBusiness`      |
| `cutomersBusiness`     | `customersBusiness`     |
| `CompanysPresentation` | `CompaniesPresentation` |
| `BranchsPresentation`  | `BranchesPresentation`  |
| `enviroment.go`        | `environment.go`        |
| `GetAllconference`     | `GetAll`                |

These names appear in multiple areas and make the code harder to understand and maintain.

---

### Phase 4 — Queue flow refactor

Create a dedicated queue domain service:

```text
features/customers/business/queue_service.go
```

Responsibilities:

```text
- assign queue number
- calculate queue-left
- calculate estimated wait time
- enforce queue reset time
- validate state transitions
- validate pharmacy flow
```

The project already uses a 04:00 Asia/Jakarta queue reset rule, so this should become a named domain rule instead of scattered date logic.

---

### Phase 5 — Integration flow refactor

Current Medifrans factory wires many services manually and has typos like `cutomersBusiness` and `custommerBusiness`.

Refactor to:

```text
integration/medifrans/
  client/
  repository/
  service/
  handler/
  routes/
  module.go
```

Use:

```go
type Module struct {
	Router routes.Router
}

func NewModule(deps Dependencies) *Module
```

This makes dependencies clearer and easier to test.

---

## 5. Final Improved Flow

### Authentication

```text
User login
   ↓
Validate email/username + password
   ↓
Update user status
   ↓
Create encrypted JWT claims
   ↓
Return token
```

Fix required: validate config before login and stop logging encryption key length. JWT expiry currently uses the next 04:00 Asia/Jakarta cycle in production, with shorter expiry in non-production.

---

### Authorization

```text
JWT validates identity
   ↓
Casbin checks subject → role → policy
   ↓
Path matching uses keyMatch2
   ↓
Request allowed or denied
```

Fix required: normalize policy paths and remove duplicate Casbin rules.

---

### Scanner flow

```text
Scanner sends x-client-id + x-api-key
   ↓
Middleware validates client against DB
   ↓
Scanner login creates scanner JWT
   ↓
Scanner check-in uses scanner JWT
   ↓
Station/menu/branch relationship is validated
```

The repository already contains client lookup and station relationship validation methods.

---

### Pharmacy queue flow

```text
Patient queue history checked
   ↓
Target pharmacy menu detected
   ↓
System verifies patient passed Penerimaan Resep
   ↓
Forward/register allowed only if rule passes
```

Fix required: do not rely on menu-name regex only. Use a stable menu code or database flag.

---

## 6. Recommended File Changes

```text
config/casbin_auth_model.conf
- Change keyMatch to keyMatch2.

config/casbin_auth_policy.csv
config/casbin_auth_policy_bithealth.csv
- Standardize dynamic route syntax.

routes/routes.go
- Replace wildcard CORS.
- Protect debug route.
- Reduce DB stats logging per request.

middleware/casbin.go
- Add nil enforcer guard.
- Deduplicate policies before saving.

features/users/business/business.go
- Fix ignored RegisterUser error.
- Remove misleading error block.

features/roles/business/business.go
- Remove duplicate policy appends.
- Add transactional update around DB permission changes + Casbin policy changes.

features/customers/business/pharmacy_validator.go
- Replace regex-only rule with menu code/flag.

features/customers/business/business.go
- Replace GetQueueEstimate placeholder.
- Move queue state logic into QueueService.

features/customers/data/mysql.go
- Whitelist order values.
- Add indexes for queue queries.
- Avoid repeated JSON_EXTRACT where possible.

integration/medifrans/routes/medifrans_middleware.go
- Validate scanner API key against repository.

integration/medifrans/factory.go
- Rename typo variables.
- Introduce dependency struct.
```

---

## 7. Priority Checklist

**Critical**

- Fix Casbin matcher.
- Fix scanner API-key validation.
- Replace wildcard CORS.
- Protect debug routes.

**High**

- Fix ignored Casbin error during user registration.
- Remove duplicate Casbin policies.
- Whitelist sort/order input.
- Replace fake queue estimate.

**Medium**

- Refactor pharmacy validation.
- Rename confusing variables and methods.
- Standardize response errors.
- Improve queue service boundaries.

**Low**

- Clean old comments.
- Rename files with typos.
- Add architecture documentation.
- Improve test naming.

---

## 8. Testing Plan

### Authorization tests

```text
GET /branch/1 should match /branch/:id with keyMatch2.
GET /branch/1 should not fail because policy used colon syntax.
Unauthorized role should receive 403.
Duplicate policies should not be created.
```

### Scanner tests

```text
Missing x-client-id returns 401.
Missing x-api-key returns 401.
Invalid client returns 401.
Invalid API key returns 401.
Valid scanner credentials return token.
Scanner token can access /medifrans/scanner/check-in.
```

### Queue tests

```text
Queue day starts at 04:00 Asia/Jakarta.
Queue at 03:59 belongs to previous queue day.
Queue at 04:00 starts new queue day.
Queue-left counts only not_called records.
Queue estimate is based on real average duration.
```

### Pharmacy tests

```text
Regular menu → pharmacy menu rejected unless source is Penerimaan Resep.
Penerimaan Resep → pharmacy menu allowed.
Pharmacy → pharmacy allowed.
Pharmacy → non-pharmacy allowed.
Regex-only false positives are rejected after refactor.
```

### Security tests

```text
CORS rejects unknown origin.
Debug route is unavailable in production.
JWT secret missing fails startup.
Invalid encryption key length fails startup.
Sort injection input is ignored or rejected.
```

---

## 9. Final Verdict

The project is functional but needs refactoring around **authorization, scanner authentication, queue business rules, pharmacy validation, and naming consistency**. The most dangerous issues are the Casbin matcher mismatch, wildcard credentialed CORS, scanner middleware that only checks header presence, duplicate Casbin policy logic, and fake queue estimates.

The safest improvement path is not a full rewrite. Start with targeted fixes, add tests, then gradually extract queue, scanner, and authorization logic into dedicated services.
