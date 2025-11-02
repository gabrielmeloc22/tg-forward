package rules

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Rule struct {
	ID      string `json:"id" bson:"_id"`
	Name    string `json:"name" bson:"name"`
	Pattern string `json:"pattern" bson:"pattern"`
}

type Repository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewRepository(uri, database, collection string) (*Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	coll := client.Database(database).Collection(collection)

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	}
	if _, err := coll.Indexes().CreateOne(ctx, indexModel); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &Repository{
		client:     client,
		collection: coll,
	}, nil
}

func (r *Repository) GetRules() ([]Rule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find rules: %w", err)
	}
	defer cursor.Close(ctx)

	var rules []Rule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, fmt.Errorf("failed to decode rules: %w", err)
	}

	if rules == nil {
		rules = []Rule{}
	}

	return rules, nil
}

func (r *Repository) GetRuleByID(id string) (*Rule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rule Rule
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("rule not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find rule: %w", err)
	}

	return &rule, nil
}

func (r *Repository) SetRules(rules []Rule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := r.collection.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	if len(rules) == 0 {
		return nil
	}

	docs := make([]interface{}, len(rules))
	for i, rule := range rules {
		docs[i] = rule
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to insert rules: %w", err)
	}

	return nil
}

func (r *Repository) AddRule(name, pattern string) (*Rule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rule := Rule{
		ID:      generateID(),
		Name:    name,
		Pattern: pattern,
	}

	_, err := r.collection.InsertOne(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to insert rule: %w", err)
	}

	return &rule, nil
}

func (r *Repository) RemoveRule(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("rule not found: %s", id)
	}

	return nil
}

func (r *Repository) GetPatterns() ([]string, error) {
	rules, err := r.GetRules()
	if err != nil {
		return nil, err
	}

	patterns := make([]string, len(rules))
	for i, rule := range rules {
		patterns[i] = rule.Pattern
	}

	return patterns, nil
}

func (r *Repository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Disconnect(ctx)
}

func generateID() string {
	return uuid.New().String()
}
