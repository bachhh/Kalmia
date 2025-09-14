package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	muxHandlers "github.com/gorilla/handlers"
)

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}

type S3 struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Bucket          string `json:"bucket"`
	UsePathStyle    bool   `json:"usePathStyle"`
	PublicUrlFormat string `json:"publicUrlFormat"`
}

type GithubOAuth struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURL  string `json:"callbackUrl"`
}

type MicrosoftOAuth struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	DirectoryID  string `json:"directoryId"`
	RedirectURL  string `json:"callbackUrl"`
}

type GoogleOAuth struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURL  string `json:"callbackUrl"`
}

type Config struct {
	LogSubCmd      bool           `json:"logSubCmd"`
	Environment    string         `json:"environment"`
	Host           string         `json:"host"`
	Port           int            `json:"port"`
	Security       Security       `json:"security"`
	Database       string         `json:"database"`
	LogLevel       string         `json:"logLevel"`
	AssetStorage   string         `json:"assetStorage"`
	MaxFileSize    int64          `json:"maxFileSize"` // in MB
	SessionSecret  string         `json:"sessionSecret"`
	Admins         []User         `json:"users"`
	DataPath       string         `json:"dataPath"`
	S3             S3             `json:"s3"`
	GithubOAuth    GithubOAuth    `json:"githubOAuth"`
	MicrosoftOAuth MicrosoftOAuth `json:"microsoftOAuth"`
	GoogleOAuth    GoogleOAuth    `json:"googleOAuth"`
	BodyLimitMb    int64          `json:"bodyLimitMb"`
	PathToSecret   string         `json:"pathToSecretFile"`
	Secret         Secret         `json:"-"`
}

type Security struct {
	CORSConfig   CORSConfig   `json:"corsConfig"`
	CookieConfig CookieConfig `json:"cookieConfig"`
	// TODO CFRSConfig CFRSConfig
	// other security config here
}

type CookieConfig struct {
	Domain   string   `json:"domain"`
	Path     string   `json:"path"`
	AgeDays  uint     `json:"ageDays"`
	Secure   bool     `json:"secure"`
	SameSite SameSite `json:"sameSite"` // lax / strict / none
	HttpOnly bool     `json:"httpOnly"`
}

type SameSite http.SameSite

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *SameSite) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch strings.ToLower(str) {
	case "lax":
		*s = SameSite(http.SameSiteLaxMode)
	case "strict":
		*s = SameSite(http.SameSiteStrictMode)
	case "none":
		*s = SameSite(http.SameSiteNoneMode)
	default:
		return fmt.Errorf("unknown SameSite value: %q", str)
	}
	return nil
}

type (
	// TODO CFRSConfig struct{}
	CORSConfig struct {
		AllowedOrigins   []string `json:"allowedOrigins"`
		AllowedMethods   []string `json:"allowedMethods"`
		AllowedHeaders   []string `json:"allowedHeaders"`
		AllowCredentials bool     `json:"allowCredentials"`
		ExposedHeaders   []string `json:"exposedHeaders"`
		CORSMaxAge       int      `json:"corsMaxAge"`
	}
)

// ToMuxCORSOptions converts the CORSConfig struct into a slice of gorilla/handlers.CORSOption.
func (c *CORSConfig) ToMuxCORSOptions() []muxHandlers.CORSOption {
	options := []muxHandlers.CORSOption{
		muxHandlers.AllowedOrigins(c.AllowedOrigins),
		muxHandlers.AllowedMethods(c.AllowedMethods),
		muxHandlers.AllowedHeaders(c.AllowedHeaders),
		muxHandlers.ExposedHeaders(c.ExposedHeaders),
		muxHandlers.MaxAge(c.CORSMaxAge),
		muxHandlers.AllowCredentials(),
	}

	if c.AllowCredentials {
		options = append(options, muxHandlers.AllowCredentials())
	}

	return options
}

//nolint:gochecknoglobals
var (
	DefaultAllowedOrigins   = []string{"*"}
	DefaultAllowedMethods   = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	DefaultAllowedHeaders   = []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}
	DefaultExposedHeaders   = []string{"Link"}
	DefaultAllowCredentials = true
)

// set default values to CORSConfig if they are empty.
func (cfg *CORSConfig) SetDefault() {
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = DefaultAllowedOrigins
	}
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = DefaultAllowedMethods
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = DefaultAllowedHeaders
	}
	if len(cfg.ExposedHeaders) == 0 {
		cfg.ExposedHeaders = DefaultExposedHeaders
	}

	if !cfg.AllowCredentials {
		cfg.AllowCredentials = DefaultAllowCredentials
	}
}

type Secret struct {
	JwtSecretKey string `json:"JwtSecretKey"`
}

//nolint:gochecknoglobals
var ParsedConfig *Config

func ParseConfig(path string) *Config {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	ParsedConfig = &Config{}
	err = decoder.Decode(ParsedConfig)
	if err != nil {
		panic(err)
	}

	err = SetupDataPath()
	if err != nil {
		panic(err)
	}

	if ParsedConfig.AssetStorage == "local" {
		SetupLocalS3Storage()
	}

	// INFO: Adds the default max file size of 10
	// Added for backwards compatibility
	if ParsedConfig.MaxFileSize == 0 {
		ParsedConfig.MaxFileSize = 10
	}

	// sensible default for body limit
	if ParsedConfig.BodyLimitMb == 0 {
		ParsedConfig.BodyLimitMb = 50
	}

	// sensible defaualt for cors
	ParsedConfig.Security.CORSConfig.SetDefault()

	if ParsedConfig.PathToSecret == "" {
		panic("path to secret file is empty")
	}
	f, err := os.Open(ParsedConfig.PathToSecret)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	rd, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(rd, &ParsedConfig.Secret)
	if err != nil {
		panic(err)
	}

	return ParsedConfig
}

func SetupDataPath() error {
	if ParsedConfig.DataPath == "" {
		ParsedConfig.DataPath = "./data"
	}

	if _, err := os.Stat(ParsedConfig.DataPath); os.IsNotExist(err) {
		err := os.Mkdir(ParsedConfig.DataPath, 0o755)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(ParsedConfig.DataPath + "/rspress_data"); os.IsNotExist(err) {
		err := os.Mkdir(ParsedConfig.DataPath+"/rspress_data", 0o755)
		if err != nil {
			return err
		}
	}

	return nil
}

func SetupLocalS3Storage() {
	if ParsedConfig.Environment == "dev" {
		ParsedConfig.S3.Endpoint = "http://localhost:9000"
	} else {
		ParsedConfig.S3.Endpoint = "http://minio:9000"
	}
	envAccessKeyID := os.Getenv("KAL_MINIO_ROOT_USER")
	envSecretAccessKey := os.Getenv("KAL_MINIO_ROOT_PASSWORD")
	if len(envAccessKeyID) == 0 {
		envAccessKeyID = "minio_kalmia_user"
	}
	if len(envSecretAccessKey) == 0 {
		//nolint:gosec
		envSecretAccessKey = "minio_kalmia_password"
	}
	ParsedConfig.S3.AccessKeyId = envAccessKeyID
	ParsedConfig.S3.SecretAccessKey = envSecretAccessKey
	ParsedConfig.S3.Bucket = "uploads"
	ParsedConfig.S3.UsePathStyle = true
	ParsedConfig.S3.Region = "auto"
	// TODO: change this to be something different/dynamic which could be fetched as per config
	ParsedConfig.S3.PublicUrlFormat = "http://localhost:9000/%s/%s"
}
