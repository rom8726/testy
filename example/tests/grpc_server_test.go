package tests

import (
	"context"
	"net"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/rom8726/pgfixtures"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"

	"github.com/rom8726/testy/v2"
	"testyexample"
	"testyexample/userspb"
)

func TestGRPCServer(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container with testcontainers
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		postgres.WithDatabase("db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres container: %v", err)
		}
	}()

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}
	// Convert postgres:// to postgresql:// for compatibility with lib/pq
	if len(connStr) >= 11 && connStr[:11] == "postgres://" {
		connStr = "postgresql://" + connStr[11:]
	}

	// Apply migration
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	migrationPath := filepath.Join(testDir, "../grpc_migration.sql")
	if err := applyMigration(ctx, connStr, migrationPath); err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	// Create gRPC service servers
	userService, err := testyexample.NewUserServiceServer(connStr)
	if err != nil {
		t.Fatalf("failed to create user service: %v", err)
	}
	defer userService.Close()

	adminService, err := testyexample.NewAdminServiceServer(connStr)
	if err != nil {
		t.Fatalf("failed to create admin service: %v", err)
	}
	defer adminService.Close()

	// Start gRPC server
	grpcServer := grpc.NewServer()
	userspb.RegisterUserServiceServer(grpcServer, userService)
	userspb.RegisterAdminServiceServer(grpcServer, adminService)

	// Start listening on a random port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	// Get the actual address
	grpcAddr := listener.Addr().String()

	// Start server in goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()
	defer grpcServer.GracefulStop()

	// Get test directory paths (relative to test file location)
	casesDir := filepath.Join(testDir, "cases", "grpc")
	fixturesDir := filepath.Join(testDir, "fixtures")

	// Ensure paths are absolute for reliability
	casesDir, err = filepath.Abs(casesDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for cases: %v", err)
	}
	fixturesDir, err = filepath.Abs(fixturesDir)
	if err != nil {
		t.Fatalf("failed to get absolute path for fixtures: %v", err)
	}

	// Run testy tests
	cfg := testy.Config{
		GRPCAddr:    grpcAddr,
		DBType:      pgfixtures.PostgreSQL,
		CasesDir:    casesDir,
		FixturesDir: fixturesDir,
		ConnStr:     connStr,
		BeforeReq: func() error {
			return nil
		},
		AfterReq: func() error {
			return nil
		},
		JUnitReport: filepath.Join(testDir, "junit_grpc.xml"),
	}

	testy.Run(t, &cfg)
}
