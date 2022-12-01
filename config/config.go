package config

type Config struct {
	Environment   *string
	MongoUri      *string
	MongoDbName   *string
	JwtPrivateKey *string
}

var (
	EnvIsProd = "production"
	EnvIsDev  = "development"
	Conf      *Config
)
