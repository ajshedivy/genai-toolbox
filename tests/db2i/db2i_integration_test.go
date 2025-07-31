package db2i

import (
	"fmt"
	"os"
	"testing"

	mapepire "github.com/deady54/mapepire-go"
)

var (
	DB2iSourceKind = "db2i"
	DB2iToolKind   = "db2i-sql"
	DB2iDatabase   = os.Getenv("DB2I_DATABASE")
	DB2iHost       = os.Getenv("DB2I_HOST")
	DB2iPort       = os.Getenv("DB2I_PORT")
	DB2iUser       = os.Getenv("DB2I_USER")
	DB2iPass       = os.Getenv("DB2I_PASS")
)

func getDB2iVars(t *testing.T) map[string]any {
	switch "" {
	case DB2iHost:
		t.Skip("'DB2I_HOST' not set - skipping integration test")
	case DB2iPort:
		t.Skip("'DB2I_PORT' not set - skipping integration test")
	case DB2iUser:
		t.Skip("'DB2I_USER' not set - skipping integration test")
	case DB2iPass:
		t.Skip("'DB2I_PASS' not set - skipping integration test")
	}

	// Use specified database name
	database := DB2iDatabase
	if database == "" {
		database = "my-db2i-db"
	}

	return map[string]any{
		"kind":     DB2iSourceKind,
		"host":     DB2iHost,
		"port":     DB2iPort,
		"database": database,
		"user":     DB2iUser,
		"password": DB2iPass,
	}
}

// Initialize DB2i connection pool for testing
func initDB2iConnectionPool(host, port, user, pass string) (*mapepire.JobPool, error) {
	creds := mapepire.DaemonServer{
		Host:               host,
		Port:               port,
		User:               user,
		Password:           pass,
		IgnoreUnauthorized: true,
	}
	options := mapepire.PoolOptions{Creds: creds, MaxSize: 5, StartingSize: 3, MaxWaitTime: 1}
	pool, err := mapepire.NewPool(options)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	res, err := pool.ExecuteSQL("SELECT 1 FROM SYSIBM.SYSDUMMY1")
	if err != nil {
		return nil, fmt.Errorf("unable to execute test query: %w", err)
	}
	if !res.Success {
		return nil, fmt.Errorf("test query failed: %s", res.Error)
	}

	return pool, nil
}

func TestDB2iBasicConnection(t *testing.T) {
	_ = getDB2iVars(t) // This will skip the test if env vars aren't set

	pool, err := initDB2iConnectionPool(DB2iHost, DB2iPort, DB2iUser, DB2iPass)
	if err != nil {
		t.Fatalf("unable to create DB2i connection pool: %s", err)
	}

	// Test basic query
	res, err := pool.ExecuteSQL("SELECT 1 FROM SYSIBM.SYSDUMMY1")
	if err != nil {
		t.Fatalf("unable to execute test query: %s", err)
	}

	if !res.Success {
		t.Fatalf("test query failed: %s", res.Error)
	}

	if len(res.Data) == 0 {
		t.Fatalf("expected at least one row in result")
	}

	t.Logf("DB2i connection test successful. Got %d rows", len(res.Data))
}
