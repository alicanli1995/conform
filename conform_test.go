package conform_test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alicanli1995/conform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig represents a comprehensive test configuration
type TestConfig struct {
	Port     int            `conform:"env=PORT,default=8080,validate=gte:1024,lte:65535"`
	Host     string         `conform:"env=HOST,default=localhost,validate=hostname"`
	Debug    bool           `conform:"env=DEBUG,default=false"`
	Timeout  time.Duration  `conform:"env=TIMEOUT,default=30s"`
	Database DatabaseConfig `conform:"prefix=DB_"`
}

type DatabaseConfig struct {
	URL      string `conform:"env=URL,required,validate=url"`
	MaxConns int    `conform:"env=MAX_CONNS,default=10,validate=gte:1,lte:100"`
}

func TestLoadGeneric_Basic(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("HOST", "api.example.com")
	os.Setenv("DB_URL", "postgres://user:pass@localhost/db")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
		os.Unsetenv("DB_URL")
	}()

	cfg, err := conform.LoadGeneric[TestConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "api.example.com", cfg.Host)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.Database.URL)
	assert.Equal(t, 10, cfg.Database.MaxConns) // default
}

func TestLoadGeneric_Defaults(t *testing.T) {
	// Set DB_URL to satisfy required field
	os.Setenv("DB_URL", "postgres://default/db")
	defer os.Unsetenv("DB_URL")

	// Don't set other env vars
	cfg, err := conform.LoadGeneric[TestConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)              // default
	assert.Equal(t, "localhost", cfg.Host)       // default
	assert.Equal(t, false, cfg.Debug)            // default
	assert.Equal(t, 30*time.Second, cfg.Timeout) // default
}

func TestLoadGeneric_RequiredField(t *testing.T) {
	// Don't set DB_URL (required)
	cfg, err := conform.LoadGeneric[TestConfig](conform.FromEnv())
	require.Error(t, err)
	assert.Nil(t, cfg)

	errList, ok := err.(*conform.ErrorList)
	require.True(t, ok)
	assert.Greater(t, len(errList.Errors), 0)
}

func TestLoadGeneric_Validation(t *testing.T) {
	os.Setenv("PORT", "80")            // Too small
	os.Setenv("DB_URL", "invalid-url") // Invalid URL
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_URL")
	}()

	cfg, err := conform.LoadGeneric[TestConfig](conform.FromEnv())
	require.Error(t, err)
	assert.Nil(t, cfg)

	errList, ok := err.(*conform.ErrorList)
	require.True(t, ok)
	assert.Greater(t, len(errList.Errors), 0)
}

func TestLoadGeneric_FieldPaths(t *testing.T) {
	os.Setenv("DB_MAX_CONNS", "200") // Exceeds max
	defer os.Unsetenv("DB_MAX_CONNS")

	_, err := conform.LoadGeneric[TestConfig](conform.FromEnv())
	require.Error(t, err)

	errList, ok := err.(*conform.ErrorList)
	require.True(t, ok)

	// Check that field path includes nested struct name
	found := false
	for _, e := range errList.Errors {
		if strings.Contains(e.FieldPath, "Database") {
			found = true
			break
		}
	}
	assert.True(t, found, "Field path should include parent struct name")
}

func TestLoadGeneric_FileConfig(t *testing.T) {
	// Create temporary YAML file
	tmpFile := "test_config.yaml"
	yamlContent := `port: 9000
host: "file.example.com"
database:
  url: "postgres://file/db"
  max_conns: 50
`
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0644)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	type FileConfig struct {
		Port     int    `conform:"file=port,default=8080"`
		Host     string `conform:"file=host,default=localhost"`
		Database struct {
			URL      string `conform:"file=database.url,required"`
			MaxConns int    `conform:"file=database.max_conns,default=10"`
		}
	}

	cfg, err := conform.LoadGeneric[FileConfig](conform.FromFile(tmpFile))
	require.NoError(t, err)
	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "file.example.com", cfg.Host)
	assert.Equal(t, "postgres://file/db", cfg.Database.URL)
	assert.Equal(t, 50, cfg.Database.MaxConns)
}

func TestLoadGeneric_TOMLConfig(t *testing.T) {
	tmpFile := "test_config.toml"
	tomlContent := `port = 7000
host = "toml.example.com"
[database]
url = "postgres://toml/db"
max_conns = 25
`
	err := os.WriteFile(tmpFile, []byte(tomlContent), 0644)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	type TOMLConfig struct {
		Port     int    `conform:"file=port,default=8080"`
		Host     string `conform:"file=host,default=localhost"`
		Database struct {
			URL      string `conform:"file=database.url,required"`
			MaxConns int    `conform:"file=database.max_conns,default=10"`
		}
	}

	cfg, err := conform.LoadGeneric[TOMLConfig](conform.FromFile(tmpFile))
	require.NoError(t, err)
	assert.Equal(t, 7000, cfg.Port)
	assert.Equal(t, "toml.example.com", cfg.Host)
	assert.Equal(t, "postgres://toml/db", cfg.Database.URL)
	assert.Equal(t, 25, cfg.Database.MaxConns)
}

func TestLoadGeneric_MultiSource(t *testing.T) {
	os.Setenv("PORT", "3000")

	tmpFile := "test_config.yaml"
	yamlContent := `port: 9000
host: "file.example.com"
`
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0644)
	require.NoError(t, err)
	defer os.Remove(tmpFile)
	defer os.Unsetenv("PORT")

	type MultiConfig struct {
		Port int    `conform:"env=PORT,file=port,default=8080"`
		Host string `conform:"env=HOST,file=host,default=localhost"`
	}

	cfg, err := conform.LoadGeneric[MultiConfig](
		conform.FromEnv(),
		conform.FromFile(tmpFile),
	)
	require.NoError(t, err)

	// Env should take priority
	assert.Equal(t, 3000, cfg.Port)
	// File should be used if env not set
	assert.Equal(t, "file.example.com", cfg.Host)
}

func TestLoadGeneric_Slices(t *testing.T) {
	type SliceConfig struct {
		Hosts []string `conform:"env=HOSTS,separator=|"`
		Ports []int    `conform:"env=PORTS,separator=,"`
	}

	os.Setenv("HOSTS", "host1|host2|host3")
	os.Setenv("PORTS", "8080,8081,8082")
	defer func() {
		os.Unsetenv("HOSTS")
		os.Unsetenv("PORTS")
	}()

	cfg, err := conform.LoadGeneric[SliceConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, []string{"host1", "host2", "host3"}, cfg.Hosts)
	assert.Equal(t, []int{8080, 8081, 8082}, cfg.Ports)
}

func TestLoadGeneric_Maps(t *testing.T) {
	type MapConfig struct {
		Settings map[string]string `conform:"env=SETTINGS"`
		Counts   map[string]int    `conform:"env=COUNTS"`
	}

	os.Setenv("SETTINGS", "key1=value1,key2=value2")
	os.Setenv("COUNTS", "one=1,two=2,three=3")
	defer func() {
		os.Unsetenv("SETTINGS")
		os.Unsetenv("COUNTS")
	}()

	cfg, err := conform.LoadGeneric[MapConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, cfg.Settings)
	assert.Equal(t, map[string]int{"one": 1, "two": 2, "three": 3}, cfg.Counts)
}

func TestLoadGeneric_MapsCustomKeys(t *testing.T) {
	type MapConfig struct {
		IntKeys map[int]string `conform:"env=INT_KEYS"`
	}

	os.Setenv("INT_KEYS", "1=one,2=two,3=three")
	defer os.Unsetenv("INT_KEYS")

	cfg, err := conform.LoadGeneric[MapConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, map[int]string{1: "one", 2: "two", 3: "three"}, cfg.IntKeys)
}

func TestLoadGeneric_TimeParsing(t *testing.T) {
	type TimeConfig struct {
		StartTime time.Time `conform:"env=START_TIME,format=2006-01-02T15:04:05Z"`
		BirthDate time.Time `conform:"env=BIRTH_DATE,format=2006-01-02"`
	}

	os.Setenv("START_TIME", "2024-01-01T00:00:00Z")
	os.Setenv("BIRTH_DATE", "1990-05-15")
	defer func() {
		os.Unsetenv("START_TIME")
		os.Unsetenv("BIRTH_DATE")
	}()

	cfg, err := conform.LoadGeneric[TimeConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, 2024, cfg.StartTime.Year())
	assert.Equal(t, 1, int(cfg.StartTime.Month()))
	assert.Equal(t, 1, cfg.StartTime.Day())
	assert.Equal(t, 1990, cfg.BirthDate.Year())
	assert.Equal(t, 5, int(cfg.BirthDate.Month()))
	assert.Equal(t, 15, cfg.BirthDate.Day())
}

func TestLoadGeneric_CustomValidator(t *testing.T) {
	type CustomConfig struct {
		Password string `conform:"env=PASSWORD,required,validate=strong_password"`
	}

	conform.RegisterValidator("strong_password", func(value interface{}, params []string) error {
		str := value.(string)
		if len(str) < 8 {
			return fmt.Errorf("password too short")
		}
		return nil
	})

	os.Setenv("PASSWORD", "weak")
	defer os.Unsetenv("PASSWORD")

	_, err := conform.LoadGeneric[CustomConfig](conform.FromEnv())
	require.Error(t, err)
}

func TestLoadGeneric_CustomConverter(t *testing.T) {
	type CustomType string

	type CustomConfig struct {
		CustomField CustomType `conform:"env=CUSTOM_FIELD,default=test"`
	}

	conform.RegisterConverter(
		reflect.TypeOf(CustomType("")),
		func(s string) (interface{}, error) {
			return CustomType("custom_" + s), nil
		},
	)

	os.Setenv("CUSTOM_FIELD", "value")
	defer os.Unsetenv("CUSTOM_FIELD")

	cfg, err := conform.LoadGeneric[CustomConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, CustomType("custom_value"), cfg.CustomField)
}

func TestLoadGeneric_ErrorMessages(t *testing.T) {
	os.Setenv("PORT", "80")          // Invalid
	os.Setenv("DB_URL", "not-a-url") // Invalid
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_URL")
	}()

	cfg, err := conform.LoadGeneric[TestConfig](conform.FromEnv())
	require.Error(t, err)
	assert.Nil(t, cfg)

	errStr := err.Error()
	assert.Contains(t, errStr, "Configuration validation failed")
	assert.Contains(t, errStr, "Port")
	assert.Contains(t, errStr, "Suggestion")
}

func TestLoadGeneric_HotReload(t *testing.T) {
	type ReloadConfig struct {
		Port int `conform:"env=PORT,default=8080"`
	}

	os.Setenv("PORT", "3000")

	reloadCount := 0
	watcher, err := conform.Watch[ReloadConfig](func(newCfg ReloadConfig) {
		reloadCount++
	}, conform.FromEnv())
	require.NoError(t, err)
	defer watcher.Stop()

	// Get initial config
	cfg := watcher.Get()
	assert.Equal(t, 3000, cfg.Port)

	// Change config
	os.Setenv("PORT", "4000")
	time.Sleep(2 * time.Second)

	// Check if reloaded
	cfg = watcher.Get()
	assert.Equal(t, 4000, cfg.Port)

	os.Unsetenv("PORT")
}

func TestLoad_BackwardCompatibility(t *testing.T) {
	type Config struct {
		Port int `conform:"env=PORT,default=8080"`
	}

	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")

	var cfg Config
	err := conform.Load(&cfg, conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, 3000, cfg.Port)
}

func TestLoadGeneric_AllValidators(t *testing.T) {
	type ValidatorConfig struct {
		Email    string `conform:"env=EMAIL,validate=email"`
		URL      string `conform:"env=URL,validate=url"`
		IP       string `conform:"env=IP,validate=ip"`
		Hostname string `conform:"env=HOSTNAME,validate=hostname"`
		Min      int    `conform:"env=MIN,validate=min:5"`
		Max      int    `conform:"env=MAX,validate=max:100"`
		Gte      int    `conform:"env=GTE,validate=gte:10"`
		Lte      int    `conform:"env=LTE,validate=lte:50"`
		OneOf    string `conform:"env=ONEOF,validate=oneof:red:green:blue"`
		Regex    string `conform:"env=REGEX,validate=regex:^[0-9]+$"`
		Len      string `conform:"env=LEN,validate=len:5"`
	}

	os.Setenv("EMAIL", "user@example.com")
	os.Setenv("URL", "https://example.com")
	os.Setenv("IP", "192.168.1.1")
	os.Setenv("HOSTNAME", "example.com")
	os.Setenv("MIN", "10")
	os.Setenv("MAX", "50")
	os.Setenv("GTE", "20")
	os.Setenv("LTE", "40")
	os.Setenv("ONEOF", "red")
	os.Setenv("REGEX", "12345")
	os.Setenv("LEN", "12345")
	defer func() {
		os.Unsetenv("EMAIL")
		os.Unsetenv("URL")
		os.Unsetenv("IP")
		os.Unsetenv("HOSTNAME")
		os.Unsetenv("MIN")
		os.Unsetenv("MAX")
		os.Unsetenv("GTE")
		os.Unsetenv("LTE")
		os.Unsetenv("ONEOF")
		os.Unsetenv("REGEX")
		os.Unsetenv("LEN")
	}()

	cfg, err := conform.LoadGeneric[ValidatorConfig](conform.FromEnv())
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", cfg.Email)
	assert.Equal(t, "https://example.com", cfg.URL)
	assert.Equal(t, "192.168.1.1", cfg.IP)
	assert.Equal(t, "example.com", cfg.Hostname)
	assert.Equal(t, 10, cfg.Min)
	assert.Equal(t, 50, cfg.Max)
	assert.Equal(t, 20, cfg.Gte)
	assert.Equal(t, 40, cfg.Lte)
	assert.Equal(t, "red", cfg.OneOf)
	assert.Equal(t, "12345", cfg.Regex)
	assert.Equal(t, "12345", cfg.Len)
}
