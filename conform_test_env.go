package conform

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGeneric_EnvironmentSpecific(t *testing.T) {
	// Create environment-specific config files
	devFile := "test_config.dev.yaml"
	prodFile := "test_config.production.yaml"

	devContent := `database:
  host: "dev.db.com"
  port: 5432
server:
  port: 8080
`
	prodContent := `database:
  host: "prod.db.com"
  port: 5433
server:
  port: 443
`

	err := os.WriteFile(devFile, []byte(devContent), 0644)
	require.NoError(t, err)
	defer os.Remove(devFile)

	err = os.WriteFile(prodFile, []byte(prodContent), 0644)
	require.NoError(t, err)
	defer os.Remove(prodFile)

	type EnvConfig struct {
		Database struct {
			Host string `conform:"file=database.${ENV}.host,default=localhost"`
			Port int    `conform:"file=database.${ENV}.port,default=5432"`
		}
		Server struct {
			Port int `conform:"file=server.${ENV}.port,default=8080"`
		}
	}

	// Test development environment
	devCfg, err := LoadGeneric[EnvConfig](
		WithEnvironment("development"),
		FromFile("test_config.${ENV}.yaml"),
	)
	require.NoError(t, err)
	assert.Equal(t, "dev.db.com", devCfg.Database.Host)
	assert.Equal(t, 5432, devCfg.Database.Port)
	assert.Equal(t, 8080, devCfg.Server.Port)

	// Test production environment
	prodCfg, err := LoadGeneric[EnvConfig](
		WithEnvironment("production"),
		FromFile("test_config.${ENV}.yaml"),
	)
	require.NoError(t, err)
	assert.Equal(t, "prod.db.com", prodCfg.Database.Host)
	assert.Equal(t, 5433, prodCfg.Database.Port)
	assert.Equal(t, 443, prodCfg.Server.Port)
}

func TestLoadGeneric_VariableSubstitution(t *testing.T) {
	os.Setenv("DB_HOST", "custom.db.com")
	os.Setenv("DB_PORT", "9999")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()

	type SubConfig struct {
		DatabaseURL string `conform:"env=DB_URL,default=postgres://${DB_USER:-postgres}:${DB_PASSWORD}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-mydb}"`
		APIURL      string `conform:"env=API_URL,default=https://api.${ENV:-dev}.example.com"`
	}

	cfg, err := LoadGeneric[SubConfig](
		WithEnvironment("production"),
		FromEnv(),
	)
	require.NoError(t, err)

	// Check variable substitution
	assert.Contains(t, cfg.DatabaseURL, "custom.db.com")
	assert.Contains(t, cfg.DatabaseURL, "9999")
	assert.Contains(t, cfg.APIURL, "production")
}

func TestLoadGeneric_VariableSubstitutionWithDefaults(t *testing.T) {
	type DefaultConfig struct {
		DatabaseURL string `conform:"default=postgres://${DB_USER:-postgres}:${DB_PASS:-secret}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-mydb}"`
	}

	cfg, err := LoadGeneric[DefaultConfig](
		FromEnv(),
	)
	require.NoError(t, err)

	// Should use defaults when variables not set
	assert.Contains(t, cfg.DatabaseURL, "postgres")
	assert.Contains(t, cfg.DatabaseURL, "localhost")
	assert.Contains(t, cfg.DatabaseURL, "5432")
	assert.Contains(t, cfg.DatabaseURL, "mydb")
}
