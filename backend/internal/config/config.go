package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

type Config struct {
	MongoUser     string
	MongoPassword string
	MongoHost     string
	MongoPort     string
	MongoName     string

	RedisHost     string
	RedisPort     string
	RedisUser     string
	RedisPassword string
	RedisDB       int

	StorageProvider string
	BucketName      string
	FrontendURL     string
	UserAgent       string

	MinioEndpoint string
	MinioUser     string
	MinioPassword string
	MinioUseSSL   bool

	AppEnv string
	Host   string
	Port   string

	TimeToBet            int // seconds; env: TIME_TO_BET, default 60
	TimeToPass           int // seconds; env: TIME_TO_PASS, default 60
	QuestionDemoDuration int // seconds; env: QUESTION_DEMO_DURATION, default 5
	IdleRoomTTL          int // seconds; env: IDLE_ROOM_TTL, default 600

	ValidatorType       string // env: VALIDATOR_TYPE; "ollama" | "gemini" | "" (disabled)
	OllamaURL           string // env: OLLAMA_URL
	OllamaSystemPrompt  string // env: OLLAMA_SYSTEM_PROMPT
	GeminiProjectID     string // env: GEMINI_PROJECT_ID
	GeminiLocation      string // env: GEMINI_LOCATION
	GeminiSystemPrompt  string // env: GEMINI_SYSTEM_PROMPT
	AIValidationTimeout int    // env: AI_VALIDATION_TIMEOUT; seconds; default 10
}

func Load() (*Config, error) {
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	minioUseSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	timeToBet, _ := strconv.Atoi(os.Getenv("TIME_TO_BET"))
	if timeToBet == 0 {
		timeToBet = 60
	}
	timeToPass, _ := strconv.Atoi(os.Getenv("TIME_TO_PASS"))
	if timeToPass == 0 {
		timeToPass = 60
	}
	questionDemoDuration, _ := strconv.Atoi(os.Getenv("QUESTION_DEMO_DURATION"))
	if questionDemoDuration == 0 {
		questionDemoDuration = 5
	}
	idleRoomTTL, _ := strconv.Atoi(os.Getenv("IDLE_ROOM_TTL"))
	if idleRoomTTL == 0 {
		idleRoomTTL = 600
	}

	aiValidationTimeout, _ := strconv.Atoi(os.Getenv("AI_VALIDATION_TIMEOUT"))
	if aiValidationTimeout == 0 {
		aiValidationTimeout = 10
	}

	cfg := &Config{
		MongoHost:     os.Getenv("MONGO_HOST"),
		MongoPort:     os.Getenv("MONGO_PORT"),
		MongoUser:     os.Getenv("MONGO_ROOT_USER"),
		MongoPassword: os.Getenv("MONGO_ROOT_PASSWORD"),
		MongoName:     os.Getenv("MONGO_NAME"),

		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisUser:     os.Getenv("REDIS_USER"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,

		StorageProvider: os.Getenv("STORAGE_PROVIDER"),
		BucketName:      os.Getenv("BUCKET_NAME"),

		MinioEndpoint: os.Getenv("MINIO_ENDPOINT"),
		MinioUser:     os.Getenv("MINIO_ROOT_USER"),
		MinioPassword: os.Getenv("MINIO_ROOT_PASSWORD"),
		MinioUseSSL:   minioUseSSL,

		AppEnv:      os.Getenv("APP_ENV"),
		Host:        os.Getenv("HOST"),
		Port:        os.Getenv("PORT"),
		FrontendURL: os.Getenv("FRONTEND_URL"),
		UserAgent:   os.Getenv("USER_AGENT"),

		TimeToBet:            timeToBet,
		TimeToPass:           timeToPass,
		QuestionDemoDuration: questionDemoDuration,
		IdleRoomTTL:          idleRoomTTL,

		ValidatorType:       os.Getenv("VALIDATOR_TYPE"),
		OllamaURL:           os.Getenv("OLLAMA_URL"),
		OllamaSystemPrompt:  os.Getenv("OLLAMA_SYSTEM_PROMPT"),
		GeminiProjectID:     os.Getenv("GEMINI_PROJECT_ID"),
		GeminiLocation:      os.Getenv("GEMINI_LOCATION"),
		GeminiSystemPrompt:  os.Getenv("GEMINI_SYSTEM_PROMPT"),
		AIValidationTimeout: aiValidationTimeout,
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	required := map[string]string{
		"MONGO_ROOT_USER":     c.MongoUser,
		"MONGO_ROOT_PASSWORD": c.MongoPassword,
		"MONGO_HOST":          c.MongoHost,
		"MONGO_PORT":          c.MongoPort,
		"MONGO_NAME":          c.MongoName,
		"REDIS_HOST":          c.RedisHost,
		"REDIS_PORT":          c.RedisPort,
		"STORAGE_PROVIDER":    c.StorageProvider,
		"BUCKET_NAME":         c.BucketName,
		"FRONTEND_URL":        c.FrontendURL,
		"HOST":                c.Host,
		"PORT":                c.Port,
	}

	var missing []string
	for key, val := range required {
		if val == "" {
			missing = append(missing, key)
		}
	}

	if c.StorageProvider == "minio" {
		minioRequired := map[string]string{
			"MINIO_ENDPOINT":      c.MinioEndpoint,
			"MINIO_ROOT_USER":     c.MinioUser,
			"MINIO_ROOT_PASSWORD": c.MinioPassword,
		}
		for key, val := range minioRequired {
			if val == "" {
				missing = append(missing, key)
			}
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing required env vars: %v", missing)
	}
	return nil
}
