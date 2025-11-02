package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gabrielmelo/tg-forward/internal/matcher"
	"github.com/gabrielmelo/tg-forward/internal/rules"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ExecuteRequest(req *http.Request, router *chi.Mux) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func NewRequest(t *testing.T, method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	return req
}

func NewAuthenticatedRequest(t *testing.T, method string, url string, body io.Reader, token string) *http.Request {
	req := NewRequest(t, method, url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func UnmarshallReqBody[T any](t *testing.T, data *bytes.Buffer) *T {
	var body T
	err := json.Unmarshal(data.Bytes(), &body)
	if err != nil {
		t.Errorf("could not parse body into expected format, actual: \n%s", data.String())
		return nil
	}
	return &body
}

func MarshallBody(t *testing.T, body any) io.Reader {
	if body == nil {
		return nil
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	return bytes.NewBuffer(jsonBody)
}

type MongoDBConfig struct {
	URI      string
	Database string
}

func SetupTestDB(t *testing.T) (*mongo.Client, string, func()) {
	ctx := context.Background()

	mongodbContainer, err := mongodb.Run(ctx, "mongo:6")
	require.NoError(t, err)

	uri, err := mongodbContainer.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("failed to disconnect from mongodb: %v", err)
		}

		if err := mongodbContainer.Terminate(context.Background()); err != nil {
			t.Errorf("failed to terminate mongodb container: %v", err)
		}
	}

	return client, "testdb", cleanup
}

type Fixture struct {
	RulesRepo    *rules.Repository
	RulesService *rules.Service
	Router       *chi.Mux
	Matcher      *matcher.Matcher
}

func NewFixture(t *testing.T, client *mongo.Client, database string, apiToken string, initialPatterns []string) *Fixture {
	rulesRepo, err := rules.NewRepository(
		client,
		database,
		"rules",
	)
	require.NoError(t, err)

	for _, pattern := range initialPatterns {
		_, err := rulesRepo.AddRule("test-rule", pattern, nil)
		require.NoError(t, err)
	}

	patterns, err := rulesRepo.GetPatterns()
	require.NoError(t, err)

	m, err := matcher.New(patterns)
	require.NoError(t, err)

	rulesService := rules.NewService(rulesRepo, m)

	router := rules.NewRouter(rulesService, apiToken)

	return &Fixture{
		RulesRepo:    rulesRepo,
		RulesService: rulesService,
		Router:       router,
		Matcher:      m,
	}
}
