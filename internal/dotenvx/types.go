package dotenvx

// EnvFile represents a discovered encrypted environment file.
type EnvFile struct {
	Path  string // relative path from target dir (e.g., "apps/api/.env.staging")
	Scope string // directory component relative to target ("." for root, "apps/api", etc.)
	Env   string // environment suffix after ".env." (e.g., "local", "staging", "production")
}
