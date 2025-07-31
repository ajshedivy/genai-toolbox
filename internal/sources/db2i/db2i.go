package db2i

import (
	"context"
	"fmt"

	mapepire "github.com/deady54/mapepire-go"
	"github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"go.opentelemetry.io/otel/trace"
)

const SourceKind string = "db2i"

// validate interface
var _ sources.SourceConfig = Config{}

func init() {
	if !sources.Register(SourceKind, newConfig) {
		panic(fmt.Sprintf("source kind %q already registered", SourceKind))
	}
}

func newConfig(ctx context.Context, name string, decoder *yaml.Decoder) (sources.SourceConfig, error) {
	actual := Config{Name: name}
	if err := decoder.DecodeContext(ctx, &actual); err != nil {
		return nil, err
	}
	return actual, nil
}

type Config struct {
	Name     string `yaml:"name" validate:"required"`
	Kind     string `yaml:"kind" validate:"required"`
	Host     string `yaml:"host" validate:"required"`
	Port     string `yaml:"port" validate:"required"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Database string `yaml:"database" validate:"required"`
}

func (r Config) SourceConfigKind() string {
	return SourceKind
}

func (r Config) Initialize(ctx context.Context, tracer trace.Tracer) (sources.Source, error) {

	pool, err := initDb2iConnectionPool(ctx, tracer, r.Name, r.Host, r.Port, r.User, r.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to create pool: %w", err)
	}

	s := &Source{
		Name: r.Name,
		Kind: SourceKind,
		Pool: pool,
	}
	return s, nil
}

type Source struct {
	Name string            `yaml:"name"`
	Kind string            `yaml:"kind"`
	Pool *mapepire.JobPool `yaml:"-"`
}

func (s *Source) SourceKind() string {
	return SourceKind
}

func (s *Source) Db2iPool() *mapepire.JobPool {
	return s.Pool
}

func initDb2iConnectionPool(ctx context.Context, tracer trace.Tracer, name, host, port, user, password string) (*mapepire.JobPool, error) {
	//nolint:all // Reassigned ctx
	ctx, span := sources.InitConnectionSpan(ctx, tracer, SourceKind, name)
	defer span.End()

	// Validate required connection parameters
	if host == "" {
		return nil, fmt.Errorf("host is required for db2i connection")
	}
	if port == "" {
		return nil, fmt.Errorf("port is required for db2i connection")
	}
	if user == "" {
		return nil, fmt.Errorf("user is required for db2i connection")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required for db2i connection")
	}

	creds := mapepire.DaemonServer{
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		IgnoreUnauthorized: true,
		Technique:          "tcp", // Use TCP connection
	}
	options := mapepire.PoolOptions{Creds: creds, MaxSize: 5, StartingSize: 3, MaxWaitTime: 1}
	pool, err := mapepire.NewPool(options)
	if err != nil {
		return nil, fmt.Errorf("unable to create mapepire pool: %w", err)
	}
	if pool == nil {
		return nil, fmt.Errorf("mapepire pool is nil")
	}

	res, err := pool.ExecuteSQL("select 1 from sysibm.sysdummy1")
	if err != nil {
		return nil, fmt.Errorf("unable to execute test query: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("test query result is nil")
	}
	if !res.Success {
		return nil, fmt.Errorf("unable to connect successfully: %s", res.Error)
	}

	return pool, nil
}
