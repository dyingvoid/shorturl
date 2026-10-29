//go:build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/testcontainers/testcontainers-go"
	tc_kafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	tc_mongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	tc_postgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tc_redis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const (
	postgresDB       = "shorturl_test"
	postgresUser     = "postgres"
	postgresPassword = "postgres"

	mongoDB       = "shorturl_test"
	mongoUser     = "mongo"
	mongoPassword = "mongo"

	kafkaTopic = "click_events"

	jwtSecret = "integration-secret"
)

var suite *testSuite

type testSuite struct {
	ctx          context.Context
	root         string
	binDir       string
	baseURL      string
	client       *http.Client
	postgres     serviceEndpoint
	redis        serviceEndpoint
	mongo        serviceEndpoint
	kafkaBrokers []string
	processes    []*exec.Cmd
	containers   []testcontainers.Container
}

type serviceEndpoint struct {
	host string
	port int
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	s, err := startSuite(ctx)
	if err != nil {
		log.Printf("integration suite setup failed: %v", err)
		if s != nil {
			s.cleanup()
		}
		os.Exit(1)
	}

	suite = s
	code := m.Run()
	s.cleanup()

	os.Exit(code)
}

func startSuite(ctx context.Context) (*testSuite, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	binDir, err := os.MkdirTemp("", "shorturl-integration-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary directory: %w", err)
	}

	s := &testSuite{
		ctx:    ctx,
		root:   filepath.Clean(filepath.Join(workingDir, "..", "..")),
		binDir: binDir,
	}

	if err := s.startContainers(); err != nil {
		return s, err
	}

	if err := s.startServices(); err != nil {
		return s, err
	}

	s.client = &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return s, nil
}

func (s *testSuite) startContainers() error {
	postgresContainer, err := tc_postgres.Run(
		s.ctx,
		"postgres:18-alpine",
		tc_postgres.WithDatabase(postgresDB),
		tc_postgres.WithUsername(postgresUser),
		tc_postgres.WithPassword(postgresPassword),
		tc_postgres.WithInitScripts(filepath.Join(
			s.root, "user", "migrations", "000001_init_schema.up.sql",
		)),
		tc_postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return fmt.Errorf("start postgres container: %w", err)
	}
	s.containers = append(s.containers, postgresContainer)

	if s.postgres, err = containerEndpoint(s.ctx, postgresContainer, "5432/tcp"); err != nil {
		return fmt.Errorf("postgres endpoint: %w", err)
	}

	mongoContainer, err := tc_mongodb.Run(
		s.ctx,
		"mongo:7",
		tc_mongodb.WithUsername(mongoUser),
		tc_mongodb.WithPassword(mongoPassword),
	)
	if err != nil {
		return fmt.Errorf("start mongo container: %w", err)
	}
	s.containers = append(s.containers, mongoContainer)

	if s.mongo, err = containerEndpoint(s.ctx, mongoContainer, "27017/tcp"); err != nil {
		return fmt.Errorf("mongo endpoint: %w", err)
	}

	redisContainer, err := tc_redis.Run(s.ctx, "redis:7-alpine")
	if err != nil {
		return fmt.Errorf("start redis container: %w", err)
	}
	s.containers = append(s.containers, redisContainer)

	if s.redis, err = containerEndpoint(s.ctx, redisContainer, "6379/tcp"); err != nil {
		return fmt.Errorf("redis endpoint: %w", err)
	}

	kafkaContainer, err := tc_kafka.Run(
		s.ctx,
		"confluentinc/confluent-local:7.5.0",
		tc_kafka.WithClusterID("shorturl-integration"),
	)
	if err != nil {
		return fmt.Errorf("start kafka container: %w", err)
	}
	s.containers = append(s.containers, kafkaContainer)

	if s.kafkaBrokers, err = kafkaContainer.Brokers(s.ctx); err != nil {
		return fmt.Errorf("kafka brokers: %w", err)
	}
	if len(s.kafkaBrokers) == 0 {
		return fmt.Errorf("kafka returned no brokers")
	}

	if err := ensureKafkaTopic(s.kafkaBrokers); err != nil {
		return err
	}

	return nil
}

func (s *testSuite) startServices() error {
	ports, err := freePorts(3)
	if err != nil {
		return fmt.Errorf("allocate service ports: %w", err)
	}
	userPort, urlPort, gatewayPort := ports[0], ports[1], ports[2]

	userBinary, err := s.buildModule("user", "user")
	if err != nil {
		return err
	}

	urlBinary, err := s.buildModule("url", "url")
	if err != nil {
		return err
	}

	gatewayBinary, err := s.buildModule("gateway", "gateway")
	if err != nil {
		return err
	}

	userEnv := []string{
		"LOG_LEVEL=INFO",
		"TIME_ZONE=UTC",
		"TIMEOUT=10s",
		"POSTGRES_HOST=" + s.postgres.host,
		"POSTGRES_PORT=" + strconv.Itoa(s.postgres.port),
		"POSTGRES_DB=" + postgresDB,
		"POSTGRES_USER=" + postgresUser,
		"POSTGRES_PASSWORD=" + postgresPassword,
		"REDIS_HOST=" + s.redis.host,
		"REDIS_PORT=" + strconv.Itoa(s.redis.port),
		"REDIS_TIMEOUT=5s",
		"JWT_SECRET=" + jwtSecret,
		"ACCESS_TOKEN_TTL=15m",
		"REFRESH_TOKEN_TTL=168h",
		"USER_SERVICE_PORT=" + strconv.Itoa(userPort),
		"GRPC_SHUTDOWN_TIMEOUT=5s",
	}

	urlEnv := []string{
		"LOG_LEVEL=INFO",
		"TIME_ZONE=UTC",
		"MONGO_HOST=" + s.mongo.host,
		"MONGO_PORT=" + strconv.Itoa(s.mongo.port),
		"MONGO_DB=" + mongoDB,
		"MONGO_USER=" + mongoUser,
		"MONGO_PASSWORD=" + mongoPassword,
		"MONGO_AUTH_SOURCE=admin",
		"MONGO_TIMEOUT=10s",
		"REDIS_HOST=" + s.redis.host,
		"REDIS_PORT=" + strconv.Itoa(s.redis.port),
		"REDIS_TIMEOUT=5s",
		"USER_SERVICE_HOST=127.0.0.1",
		"USER_SERVICE_PORT=" + strconv.Itoa(userPort),
		"USER_SERVICE_TIMEOUT=5s",
		"URL_SERVICE_PORT=" + strconv.Itoa(urlPort),
		"GRPC_SHUTDOWN_TIMEOUT=5s",
		"URL_CREATE_ATTEMPTS=5",
		"LINK_CACHE_TTL=5m",
		"KAFKA_BROKERS=" + strings.Join(s.kafkaBrokers, ","),
		"KAFKA_TOPIC=" + kafkaTopic,
		"URL=short.url",
	}

	gatewayEnv := []string{
		"LOG_LEVEL=INFO",
		"GATEWAY_PORT=" + strconv.Itoa(gatewayPort),
		"GATEWAY_REQUEST_TIMEOUT=10s",
		"GATEWAY_ALLOWED_ORIGINS=http://localhost:3000",
		"USER_SERVICE_HOST=127.0.0.1",
		"USER_SERVICE_PORT=" + strconv.Itoa(userPort),
		"URL_SERVICE_HOST=127.0.0.1",
		"URL_SERVICE_PORT=" + strconv.Itoa(urlPort),
	}

	if err := s.startProcess("user", userBinary, userEnv); err != nil {
		return err
	}

	if err := s.startProcess("url", urlBinary, urlEnv); err != nil {
		return err
	}

	if err := waitForTCP(net.JoinHostPort("127.0.0.1", strconv.Itoa(userPort)), 30*time.Second); err != nil {
		return fmt.Errorf("user service is not ready: %w", err)
	}

	if err := waitForTCP(net.JoinHostPort("127.0.0.1", strconv.Itoa(urlPort)), 30*time.Second); err != nil {
		return fmt.Errorf("url service is not ready: %w", err)
	}

	if err := s.startProcess("gateway", gatewayBinary, gatewayEnv); err != nil {
		return err
	}

	gatewayAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(gatewayPort))
	if err := waitForTCP(gatewayAddr, 30*time.Second); err != nil {
		return fmt.Errorf("gateway is not ready: %w", err)
	}

	s.baseURL = "http://" + gatewayAddr

	return nil
}

func (s *testSuite) buildModule(name string, moduleDir string) (string, error) {
	binary := filepath.Join(s.binDir, name)

	cmd := exec.CommandContext(s.ctx, "go", "build", "-o", binary, "./cmd")
	cmd.Dir = filepath.Join(s.root, moduleDir)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build %s: %w: %s", name, err, output.String())
	}

	return binary, nil
}

func (s *testSuite) startProcess(name string, binary string, overrides []string) error {
	cmd := exec.Command(binary)
	cmd.Env = mergeEnv(overrides)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	s.processes = append(s.processes, cmd)

	return nil
}

func (s *testSuite) cleanup() {
	for i := len(s.processes) - 1; i >= 0; i-- {
		cmd := s.processes[i]
		if cmd.Process == nil {
			continue
		}

		_ = cmd.Process.Signal(syscall.SIGTERM)

		done := make(chan struct{})
		go func(c *exec.Cmd) {
			_ = c.Wait()
			close(done)
		}(cmd)

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	}

	for i := len(s.containers) - 1; i >= 0; i-- {
		_ = testcontainers.TerminateContainer(s.containers[i])
	}

	if s.binDir != "" {
		_ = os.RemoveAll(s.binDir)
	}
}

func containerEndpoint(
	ctx context.Context,
	container testcontainers.Container,
	port string,
) (serviceEndpoint, error) {
	endpoint, err := container.PortEndpoint(ctx, port, "")
	if err != nil {
		return serviceEndpoint{}, fmt.Errorf("container endpoint %s: %w", port, err)
	}

	host, rawPort, err := net.SplitHostPort(endpoint)
	if err != nil {
		return serviceEndpoint{}, fmt.Errorf("split endpoint %q: %w", endpoint, err)
	}

	parsedPort, err := strconv.Atoi(rawPort)
	if err != nil {
		return serviceEndpoint{}, fmt.Errorf("parse endpoint port %q: %w", rawPort, err)
	}

	return serviceEndpoint{host: host, port: parsedPort}, nil
}

func ensureKafkaTopic(brokers []string) error {
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	if err != nil {
		return fmt.Errorf("create kafka admin: %w", err)
	}
	defer admin.Close()

	err = admin.CreateTopic(kafkaTopic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return fmt.Errorf("create kafka topic %s: %w", kafkaTopic, err)
	}

	return nil
}

func freePorts(count int) ([]int, error) {
	listeners := make([]net.Listener, 0, count)
	defer func() {
		for _, listener := range listeners {
			_ = listener.Close()
		}
	}()

	ports := make([]int, 0, count)
	for range count {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, err
		}

		listeners = append(listeners, listener)
		ports = append(ports, listener.Addr().(*net.TCPAddr).Port)
	}

	return ports, nil
}

func waitForTCP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		lastErr = err
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("wait for %s: %w", addr, lastErr)
}

func mergeEnv(overrides []string) []string {
	keys := make(map[string]struct{}, len(overrides))
	for _, override := range overrides {
		key, _, _ := strings.Cut(override, "=")
		keys[key] = struct{}{}
	}

	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		if _, exists := keys[key]; exists {
			continue
		}
		env = append(env, value)
	}

	return append(env, overrides...)
}
