package config

type Config struct {
	AppEnv             string
	Port               string
	DatabaseURL        string
	JWTSecret          string
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
	CORSAllowedOrigins string
}

func Load() Config {
	return Config{
		AppEnv:             GetEnv("APP_ENV", "development"),
		Port:               GetEnv("PORT", "8080"),
		DatabaseURL:        GetEnv("DATABASE_URL", ""),
		JWTSecret:          GetEnv("JWT_SECRET", ""),
		SupabaseURL:        GetEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:    GetEnv("SUPABASE_ANON_KEY", GetEnv("SUPABASE_KEY", "")),
		SupabaseServiceKey: GetEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		CORSAllowedOrigins: GetEnv("CORS_ALLOWED_ORIGINS", "*"),
	}
}
