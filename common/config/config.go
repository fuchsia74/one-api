package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/songquanpeng/one-api/common/env"
)

var (
	// MaxInlineImageSizeMB limits the size (MB) of images that can be inlined as base64 to prevent oversized payloads from overwhelming upstream providers.
	MaxInlineImageSizeMB = func() int {
		v := env.Int("MAX_INLINE_IMAGE_SIZE_MB", 30)
		if v < 0 {
			panic("MAX_INLINE_IMAGE_SIZE_MB must not be negative")
		}
		return v
	}()

	// SessionSecretEnvValue keeps the raw SESSION_SECRET input so other packages can warn about placeholder values.
	SessionSecretEnvValue = strings.TrimSpace(env.String("SESSION_SECRET", ""))
	// SessionSecret stores the effective session secret. When the provided secret is absent or has an unsupported length it is replaced or hashed to a 32-byte base64 token in init().
	SessionSecret = SessionSecretEnvValue

	// CookieMaxAgeHours controls how long session cookies stay valid. The value is interpreted in hours by the session store.
	CookieMaxAgeHours = env.Int("COOKIE_MAXAGE_HOURS", 168)
	// EnableCookieSecure forces the browser to send session cookies only over HTTPS when set to true.
	EnableCookieSecure = env.Bool("ENABLE_COOKIE_SECURE", false)

	// ServerPort overrides the --port flag when running inside container or PaaS environments.
	ServerPort = strings.TrimSpace(env.String("PORT", ""))
	// GinMode allows forcing Gin into release mode (or other modes) without recompiling.
	GinMode = strings.TrimSpace(env.String("GIN_MODE", ""))

	// ChannelSuspendSecondsFor429 defines the per-ability suspension window after hitting upstream 429 throttling errors.
	ChannelSuspendSecondsFor429 = time.Second * time.Duration(env.Int("CHANNEL_SUSPEND_SECONDS_FOR_429", 60))
	// ChannelSuspendSecondsFor5XX defines how long an ability is paused after upstream 5xx failures.
	ChannelSuspendSecondsFor5XX = time.Second * time.Duration(env.Int("CHANNEL_SUSPEND_SECONDS_FOR_5XX", 30))
	// ChannelSuspendSecondsForAuth defines the backoff window applied after quota/auth/permission errors.
	ChannelSuspendSecondsForAuth = time.Second * time.Duration(env.Int("CHANNEL_SUSPEND_SECONDS_FOR_AUTH", 60))

	// MaxItemsPerPage caps paginated API and UI responses to keep database queries predictable.
	MaxItemsPerPage = env.Int("MAX_ITEMS_PER_PAGE", 100)

	// DebugEnabled toggles verbose structured logging when DEBUG=true.
	DebugEnabled = env.Bool("DEBUG", false)
	// DebugSQLEnabled toggles per-query SQL logging when DEBUG_SQL=true.
	DebugSQLEnabled = env.Bool("DEBUG_SQL", false)
	// MemoryCacheEnabled forces the in-process cache to stay enabled even without Redis.
	MemoryCacheEnabled = env.Bool("MEMORY_CACHE_ENABLED", false)

	// PreconsumeTokenForBackgroundRequest reserves quota for asynchronous background requests that only report usage after completion.
	PreconsumeTokenForBackgroundRequest = env.Int("PRECONSUME_TOKEN_FOR_BACKGROUND_REQUEST", 15000)
	// SyncFrequency controls how frequently option/channel caches refresh from the database when enabled.
	SyncFrequency = env.Int("SYNC_FREQUENCY", 10*60)
	// ForceEmailTLSVerify enforces SMTP TLS certificate validation when sending email.
	ForceEmailTLSVerify = env.Bool("FORCE_EMAIL_TLS_VERIFY", false)

	// BatchUpdateEnabled turns on the background usage batch updater when true.
	BatchUpdateEnabled = env.Bool("BATCH_UPDATE_ENABLED", false)
	// BatchUpdateInterval sets the flush cadence (seconds) for the batch updater.
	BatchUpdateInterval = env.Int("BATCH_UPDATE_INTERVAL", 5)

	// RelayTimeout bounds upstream HTTP requests (seconds) before aborting them.
	RelayTimeout = env.Int("RELAY_TIMEOUT", 0)
	// IdleTimeout controls how long to keep streaming connections alive without traffic (seconds).
	IdleTimeout = env.Int("IDLE_TIMEOUT", 30)
	// BillingTimeoutSec is the maximum time allowed for billing reconciliation (seconds) before failing the request.
	BillingTimeoutSec = env.Int("BILLING_TIMEOUT", 300)
	// StreamingBillingIntervalSec determines how frequently streaming sessions checkpoint usage (seconds).
	StreamingBillingIntervalSec = env.Int("STREAMING_BILLING_INTERVAL", 3)

	// ExternalBillingDefaultTimeoutSec sets the default hold duration (seconds) for external billing reserves.
	ExternalBillingDefaultTimeoutSec = env.Int("EXTERNAL_BILLING_DEFAULT_TIMEOUT", 600)
	// ExternalBillingMaxTimeoutSec caps user-supplied external billing hold durations (seconds).
	ExternalBillingMaxTimeoutSec = env.Int("EXTERNAL_BILLING_MAX_TIMEOUT", 3600)

	// ShutdownTimeoutSec specifies the graceful shutdown timeout (seconds) for the HTTP server and background workers.
	ShutdownTimeoutSec = env.Int("SHUTDOWN_TIMEOUT", 360)

	// GeminiSafetySetting defines the default Gemini safety preset applied to requests without explicit overrides.
	GeminiSafetySetting = env.String("GEMINI_SAFETY_SETTING", "BLOCK_NONE")
	// Theme chooses which bundled frontend theme to render.
	Theme = env.String("THEME", "modern")

	// RequestInterval throttles billing/channel polling loops (seconds).
	RequestInterval = time.Duration(env.Int("POLLING_INTERVAL", 0)) * time.Second
	// ChannelTestFrequencyRaw retains the raw CHANNEL_TEST_FREQUENCY input for validation and documentation.
	ChannelTestFrequencyRaw = strings.TrimSpace(env.String("CHANNEL_TEST_FREQUENCY", ""))
	// ChannelTestFrequency triggers automatic channel health probes when greater than zero (seconds between probes).
	ChannelTestFrequency = func() int {
		if ChannelTestFrequencyRaw == "" {
			return 0
		}
		v, err := strconv.Atoi(ChannelTestFrequencyRaw)
		if err != nil {
			panic(fmt.Sprintf("invalid CHANNEL_TEST_FREQUENCY: %q", ChannelTestFrequencyRaw))
		}
		if v < 0 {
			return 0
		}
		return v
	}()

	// EnableMetric toggles the failure rate monitor that can disable unstable channels.
	EnableMetric = env.Bool("ENABLE_METRIC", false)
	// EnablePrometheusMetrics exposes the /metrics endpoint for Prometheus scrapers when true.
	EnablePrometheusMetrics = env.Bool("ENABLE_PROMETHEUS_METRICS", true)
	// MetricQueueSize configures the buffered queue that aggregates success/failure events before processing.
	MetricQueueSize = env.Int("METRIC_QUEUE_SIZE", 10)
	// MetricSuccessRateThreshold defines the minimum acceptable success ratio before a channel is flagged as unhealthy.
	MetricSuccessRateThreshold = env.Float64("METRIC_SUCCESS_RATE_THRESHOLD", 0.8)
	// MetricSuccessChanSize sizes the buffered success event channel.
	MetricSuccessChanSize = env.Int("METRIC_SUCCESS_CHAN_SIZE", 1024)
	// MetricFailChanSize sizes the buffered failure event channel.
	MetricFailChanSize = env.Int("METRIC_FAIL_CHAN_SIZE", 128)

	// InitialRootToken seeds an initial personal token for the root user on first boot.
	InitialRootToken = env.String("INITIAL_ROOT_TOKEN", "")
	// InitialRootAccessToken seeds an initial access token for the root user on first boot.
	InitialRootAccessToken = env.String("INITIAL_ROOT_ACCESS_TOKEN", "")

	// GeminiVersion selects the default Gemini API version when callers omit it.
	GeminiVersion = env.String("GEMINI_VERSION", "v1")

	// OnlyOneLogFile merges all rotated logs into a single file when true.
	OnlyOneLogFile = env.Bool("ONLY_ONE_LOG_FILE", false)

	// LogRetentionDays determines how many days logs are kept before the retention worker purges them (0 disables cleanup).
	LogRetentionDays = func() int {
		v := env.Int("LOG_RETENTION_DAYS", 0)
		if v < 0 {
			return 0
		}
		return v
	}()

	// LogPushAPI defines the webhook endpoint for escalated log alerts.
	LogPushAPI = env.String("LOG_PUSH_API", "")
	// LogPushType labels outbound log alerts so downstream processors can route them.
	LogPushType = env.String("LOG_PUSH_TYPE", "")
	// LogPushToken authenticates outbound log alert requests.
	LogPushToken = env.String("LOG_PUSH_TOKEN", "")

	// RelayProxy provides an HTTP proxy for outbound relay requests to upstream providers.
	RelayProxy = env.String("RELAY_PROXY", "")
	// UserContentRequestProxy provides an HTTP proxy when fetching user-supplied assets like external images.
	UserContentRequestProxy = env.String("USER_CONTENT_REQUEST_PROXY", "")
	// UserContentRequestTimeout limits fetch time (seconds) for user-supplied assets.
	UserContentRequestTimeout = env.Int("USER_CONTENT_REQUEST_TIMEOUT", 30)

	// TokenKeyPrefix configures the prefix returned when new API tokens are created.
	TokenKeyPrefix = env.String("TOKEN_KEY_PREFIX", "sk-")

	// EnforceIncludeUsage forces upstream adapters to return usage accounting; requests without usage are rejected when true.
	EnforceIncludeUsage = env.Bool("ENFORCE_INCLUDE_USAGE", true)
	// TestPrompt holds the default test prompt used in automated channel diagnostics.
	TestPrompt = env.String("TEST_PROMPT", "2 + 2 = ?")
	// TestMaxTokens caps the tokens requested by the diagnostic test prompt.
	TestMaxTokens = env.Int("TEST_MAX_TOKENS", 1024)

	// OpenrouterProviderSort selects the ordering strategy when listing OpenRouter providers.
	OpenrouterProviderSort = env.String("OPENROUTER_PROVIDER_SORT", "")

	// DefaultMaxToken enforces a global max token value when model-specific limits are unknown.
	DefaultMaxToken = env.Int("DEFAULT_MAX_TOKEN", 2048)
	// DefaultUseMinMaxTokensModel controls whether new channels use the min/max token scheme by default.
	DefaultUseMinMaxTokensModel = env.Bool("DEFAULT_USE_MIN_MAX_TOKENS_MODEL", false)

	// RedisConnString defines the Redis connection string; leaving it empty disables Redis features.
	RedisConnString = strings.TrimSpace(env.String("REDIS_CONN_STRING", ""))
	// RedisMasterName enables Redis sentinel/cluster discovery when provided.
	RedisMasterName = strings.TrimSpace(env.String("REDIS_MASTER_NAME", ""))
	// RedisPassword supplies the Redis authentication password when required.
	RedisPassword = env.String("REDIS_PASSWORD", "")

	// FrontendBaseURL redirects dashboard traffic to an external frontend; follower nodes ignore it.
	FrontendBaseURL = strings.TrimSuffix(strings.TrimSpace(env.String("FRONTEND_BASE_URL", "")), "/")

	// IsMasterNode determines whether this process should serve the web UI (any value other than "slave" treats the node as master).
	IsMasterNode = !strings.EqualFold(env.String("NODE_TYPE", ""), "slave")

	// SQLDSN provides the primary database DSN; empty indicates that SQLite should be used.
	SQLDSN = strings.TrimSpace(env.String("SQL_DSN", ""))

	// GlobalApiRateLimitNum bounds the number of REST API requests per IP within a three minute window.
	GlobalApiRateLimitNum = env.Int("GLOBAL_API_RATE_LIMIT", 480)
	// GlobalApiRateLimitDuration sets the duration (seconds) of the API rate limit window.
	GlobalApiRateLimitDuration int64 = 3 * 60

	// GlobalWebRateLimitNum bounds the number of dashboard requests per IP within a three minute window.
	GlobalWebRateLimitNum = env.Int("GLOBAL_WEB_RATE_LIMIT", 240)
	// GlobalWebRateLimitDuration sets the duration (seconds) of the dashboard rate limit window.
	GlobalWebRateLimitDuration int64 = 3 * 60

	// GlobalRelayRateLimitNum bounds the number of relay API calls per token within a three minute window.
	GlobalRelayRateLimitNum = env.Int("GLOBAL_RELAY_RATE_LIMIT", 480)
	// GlobalRelayRateLimitDuration sets the duration (seconds) of the relay token rate limit window.
	GlobalRelayRateLimitDuration int64 = 3 * 60

	// ChannelRateLimitEnabled toggles per-channel rate limiting when true.
	ChannelRateLimitEnabled = env.Bool("GLOBAL_CHANNEL_RATE_LIMIT", false)
	// ChannelRateLimitDuration sets the duration (seconds) of the per-channel rate limit window.
	ChannelRateLimitDuration int64 = 3 * 60

	// CriticalRateLimitNum defines the burst control for high sensitivity endpoints (seconds window is CriticalRateLimitDuration).
	CriticalRateLimitNum = env.Int("CRITICAL_RATE_LIMIT", 20)
	// CriticalRateLimitDuration sets the window (seconds) for critical rate limiting.
	CriticalRateLimitDuration int64 = 20 * 60

	// UploadRateLimitNum bounds the number of file uploads allowed per client within UploadRateLimitDuration.
	UploadRateLimitNum = 10
	// UploadRateLimitDuration sets the upload rate limit window (seconds).
	UploadRateLimitDuration int64 = 60

	// DownloadRateLimitNum bounds the number of file downloads allowed per client within DownloadRateLimitDuration.
	DownloadRateLimitNum = 10
	// DownloadRateLimitDuration sets the download rate limit window (seconds).
	DownloadRateLimitDuration int64 = 60

	// SQLitePath specifies the SQLite database file path when SQL_DSN is absent.
	SQLitePath = env.String("SQLITE_PATH", "one-api.db")
	// SQLiteBusyTimeout configures SQLite busy timeout in milliseconds to mitigate locking errors.
	SQLiteBusyTimeout = env.Int("SQLITE_BUSY_TIMEOUT", 3000)

	// SQLMaxIdleConns controls the primary database pool's idle connection count.
	SQLMaxIdleConns = env.Int("SQL_MAX_IDLE_CONNS", 200)
	// SQLMaxOpenConns controls the primary database pool's maximum open connections.
	SQLMaxOpenConns = env.Int("SQL_MAX_OPEN_CONNS", 2000)
	// SQLMaxLifetimeSeconds sets how long database connections live before being recycled (seconds).
	SQLMaxLifetimeSeconds = env.Int("SQL_MAX_LIFETIME", 300)

	// LogSQLDSN overrides the DSN used for the logging database; falls back to SQL_DSN when empty.
	LogSQLDSN = env.String("LOG_SQL_DSN", "")

	// OneAPITestAPIBase configures the base URL used by the cmd/test smoke tester.
	OneAPITestAPIBase = strings.TrimSpace(env.String("API_BASE", ""))
	// OneAPITestToken configures the API token consumed by the cmd/test smoke tester.
	OneAPITestToken = strings.TrimSpace(env.String("API_TOKEN", ""))
	// OneAPITestModels lists comma-separated models exercised by the cmd/test smoke tester.
	OneAPITestModels = strings.TrimSpace(env.String("ONEAPI_TEST_MODELS", ""))

	// OneAPITestVariants limits the cmd/test smoke tester to specific API formats (variants).
	OneAPITestVariants = strings.TrimSpace(env.String("ONEAPI_TEST_VARIANTS", ""))
)

// RateLimitKeyExpirationDuration controls how long Redis keys for rate limiting remain valid.
var RateLimitKeyExpirationDuration = 20 * time.Minute

func init() {
	if SessionSecretEnvValue == "" {
		fmt.Println("SESSION_SECRET not set, using random secret")
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic(fmt.Sprintf("failed to generate random secret: %v", err))
		}

		SessionSecret = base64.StdEncoding.EncodeToString(key)
	} else if !slices.Contains([]int{16, 24, 32}, len(SessionSecretEnvValue)) {
		hashed := sha256.Sum256([]byte(SessionSecretEnvValue))
		SessionSecret = base64.StdEncoding.EncodeToString(hashed[:32])
	}

	logConsumeEnabled.Store(true)
}

var (
	// SystemName is displayed in the dashboard header and email templates.
	SystemName = "One API"
	// ServerAddress forms absolute URLs in email templates and redirect flows.
	ServerAddress = "http://localhost:3000"
	// Footer supplies custom HTML appended to the dashboard footer.
	Footer = ""
	// Logo provides the dashboard logo URL.
	Logo = ""
	// TopUpLink points users to the recharge page referenced in quota notifications.
	TopUpLink = ""
	// ChatLink links to the default chat UI shown in the dashboard shortcuts.
	ChatLink = ""
	// QuotaPerUnit defines the legacy conversion rate between quota and USD for display.
	QuotaPerUnit = 500 * 1000.0
	// DisplayInCurrencyEnabled toggles quota display in currency instead of raw tokens.
	DisplayInCurrencyEnabled = true
	// DisplayTokenStatEnabled toggles the token statistics card on the dashboard.
	DisplayTokenStatEnabled = true
)

var (
	// OptionMap caches key/value pairs loaded from the database options table.
	OptionMap map[string]string
	// OptionMapRWMutex guards concurrent reads/writes to OptionMap.
	OptionMapRWMutex sync.RWMutex
)

var (
	// DefaultItemsPerPage controls pagination defaults for tables that do not override the value.
	DefaultItemsPerPage = 10
	// MaxRecentItems limits the number of recent actions retained in memory for widgets like Recent Logs.
	MaxRecentItems = 100
)

var (
	// PasswordLoginEnabled toggles email/password login support.
	PasswordLoginEnabled = true
	// PasswordRegisterEnabled toggles self-service registration.
	PasswordRegisterEnabled = true
	// EmailVerificationEnabled forces email verification during registration.
	EmailVerificationEnabled = false
	// GitHubOAuthEnabled toggles GitHub OAuth login.
	GitHubOAuthEnabled = false
	// OidcEnabled toggles generic OIDC login.
	OidcEnabled = false
	// WeChatAuthEnabled toggles WeChat login support.
	WeChatAuthEnabled = false
	// TurnstileCheckEnabled toggles Cloudflare Turnstile verification on the login UI.
	TurnstileCheckEnabled = false
	// RegisterEnabled disables all new-user registration when set to false.
	RegisterEnabled = true
)

var (
	// EmailDomainRestrictionEnabled allows limiting registrations to EmailDomainWhitelist.
	EmailDomainRestrictionEnabled = false
	// EmailDomainWhitelist lists domains allowed when EmailDomainRestrictionEnabled is true.
	EmailDomainWhitelist = []string{
		"gmail.com",
		"163.com",
		"126.com",
		"qq.com",
		"outlook.com",
		"hotmail.com",
		"icloud.com",
		"yahoo.com",
		"foxmail.com",
	}
)

var (
	// SMTPServer holds the SMTP hostname for outbound email.
	SMTPServer = ""
	// SMTPPort holds the SMTP port.
	SMTPPort = 587
	// SMTPAccount stores the SMTP username.
	SMTPAccount = ""
	// SMTPFrom defines the From address used in outbound email.
	SMTPFrom = ""
	// SMTPToken stores the SMTP password or token.
	SMTPToken = ""
)

var (
	// GitHubClientId stores the OAuth client ID for GitHub login.
	GitHubClientId = ""
	// GitHubClientSecret stores the OAuth client secret for GitHub login.
	GitHubClientSecret = ""
)

var (
	// LarkClientId stores the OAuth client ID for Lark login.
	LarkClientId = ""
	// LarkClientSecret stores the OAuth client secret for Lark login.
	LarkClientSecret = ""
)

var (
	// OidcClientId stores the client ID for generic OIDC login.
	OidcClientId = ""
	// OidcClientSecret stores the client secret for generic OIDC login.
	OidcClientSecret = ""
	// OidcWellKnown caches the OIDC discovery endpoint.
	OidcWellKnown = ""
	// OidcAuthorizationEndpoint overrides the authorization endpoint when discovery is unavailable.
	OidcAuthorizationEndpoint = ""
	// OidcTokenEndpoint overrides the token endpoint when discovery is unavailable.
	OidcTokenEndpoint = ""
	// OidcUserinfoEndpoint overrides the userinfo endpoint when discovery is unavailable.
	OidcUserinfoEndpoint = ""
)

var (
	// WeChatServerAddress stores the WeChat auth server URL.
	WeChatServerAddress = ""
	// WeChatServerToken stores the WeChat auth token.
	WeChatServerToken = ""
	// WeChatAccountQRCodeImageURL points to the QR code shown during WeChat login onboarding.
	WeChatAccountQRCodeImageURL = ""
)

var (
	// MessagePusherAddress is the endpoint for optional push notification integrations.
	MessagePusherAddress = ""
	// MessagePusherToken authenticates optional push notification requests.
	MessagePusherToken = ""
)

var (
	// TurnstileSiteKey holds the Cloudflare Turnstile site key for frontend validation.
	TurnstileSiteKey = ""
	// TurnstileSecretKey holds the Cloudflare Turnstile secret for server-side verification.
	TurnstileSecretKey = ""
)

var (
	// QuotaForNewUser awards quota when a new user registers.
	QuotaForNewUser int64 = 0
	// QuotaForInviter awards quota to the inviter when a referral activates.
	QuotaForInviter int64 = 0
	// QuotaForInvitee awards quota to the invitee when they register via referral.
	QuotaForInvitee int64 = 0
	// ChannelDisableThreshold defines the failure ratio that triggers automatic channel disablement.
	ChannelDisableThreshold = 5.0
	// AutomaticDisableChannelEnabled enables automatic channel disabling when thresholds are exceeded.
	AutomaticDisableChannelEnabled = false
	// AutomaticEnableChannelEnabled re-enables channels automatically when health recovers.
	AutomaticEnableChannelEnabled = false
	// QuotaRemindThreshold determines when low quota notifications are sent.
	QuotaRemindThreshold int64 = 1000
	// PreConsumedQuota sets the default quota reservation for requests to avoid race conditions.
	PreConsumedQuota int64 = 500
	// ApproximateTokenEnabled toggles approximate token counting when exact counts are unavailable.
	ApproximateTokenEnabled = false
	// RetryTimes configures default retry attempts for certain background jobs.
	RetryTimes = 0
)

var (
	// RootUserEmail records the email for the built-in root account when seeded manually.
	RootUserEmail = ""
)

var (
	// logConsumeEnabled toggles quota consumption logging and is mutated at runtime via SetLogConsumeEnabled.
	logConsumeEnabled atomic.Bool
)

// ValidThemes enumerates the built-in frontend themes.
var ValidThemes = map[string]bool{
	"default": true,
	"berry":   true,
	"air":     true,
	"modern":  true,
}

// IsLogConsumeEnabled reports whether consumption logging is enabled.
func IsLogConsumeEnabled() bool {
	return logConsumeEnabled.Load()
}

// SetLogConsumeEnabled toggles consumption logging in a concurrency-safe way.
func SetLogConsumeEnabled(enabled bool) {
	logConsumeEnabled.Store(enabled)
}
