package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"portfolio-rebalancer/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
)

var esClient *elasticsearch.Client
var ErrUserNotFound = errors.New("user not found")
var ErrRequestNotFound = errors.New("rebalance request not found")

// InitElastic initializes elasticsearch connection with retry logic
func InitElastic() error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_URL"),
		},
	}

	var client *elasticsearch.Client
	var err error

	for i := 1; i <= 5; i++ {
		client, err = elasticsearch.NewClient(cfg)
		if err != nil {
			log.Printf("Failed to create client: %v", err)
		} else {
			_, err = client.Info()
			if err == nil {
				log.Println("Connected to Elasticsearch")
				esClient = client
				return nil
			}
			log.Printf("Client created, but ES not ready: %v", err)
		}

		log.Printf("Retrying connection to Elasticsearch... (%d/5)", i)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to connect to Elasticsearch after retries: %w", err)
}

func SavePortfolio(ctx context.Context, p *models.Portfolio) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	res, err := esClient.Index("portfolios", bytes.NewReader(body), esClient.Index.WithDocumentID(p.UserID))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving portfolio: %s", res.String())
	}

	log.Printf("Portfolio saved for user %s", p.UserID)
	return nil
}

func GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	res, err := esClient.Get("portfolios", userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, ErrUserNotFound
	}

	var esResp struct {
		Source models.Portfolio `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	return &esResp.Source, nil
}

func SaveRebalanceRequest(ctx context.Context, p *models.RebalanceRequest) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	res, err := esClient.Index("rebalance_requests", bytes.NewReader(body), esClient.Index.WithDocumentID(p.UserID))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving rebalance request: %s", res.String())
	}

	log.Printf("Rebalance request saved for user %s", p.UserID)
	return nil
}

func GetRebalanceRequest(ctx context.Context, userID string) (*models.RebalanceRequest, error) {
	res, err := esClient.Get("rebalance_requests", userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, ErrRequestNotFound
	}

	var esResp struct {
		Source models.RebalanceRequest `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	return &esResp.Source, nil
}

func SaveRebalanceTransactions(ctx context.Context, txs []models.RebalanceTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, tx := range txs {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "rebalance_transactions" } }%s`, "\n"))
		data, err := json.Marshal(tx)
		if err != nil {
			return err
		}
		data = append(data, '\n')
		buf.Write(meta)
		buf.Write(data)
	}

	res, err := esClient.Bulk(bytes.NewReader(buf.Bytes()), esClient.Bulk.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving rebalance transactions: %s", res.String())
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return fmt.Errorf("failure to to parse response body: %s", err)
	}

	if hasErrors, ok := raw["errors"].(bool); ok && hasErrors {
		return fmt.Errorf("bulk request contained errors: %v", raw)
	}

	log.Printf("Saved %d rebalance transactions for user %s", len(txs), txs[0].UserID)
	return nil
}
