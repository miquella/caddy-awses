package awses

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
)

// An Elasticsearch client factory with a cache that allows concurrent cached client sharing
type ElasticsearchClientFactory struct {
	RootSession *session.Session
	Role        string

	mutex   sync.RWMutex
	clients map[string]*elasticsearchservice.ElasticsearchService
}

func NewElasticsearchClientFactory(rootSession *session.Session, role string) *ElasticsearchClientFactory {
	return &ElasticsearchClientFactory{
		RootSession: rootSession,
		Role:        role,
	}
}

// Returns a new client (does not lock or use the cache)
func (f *ElasticsearchClientFactory) New(region string) *elasticsearchservice.ElasticsearchService {
	config := aws.Config{
		Region: &region,
	}

	if f.Role != "" {
		config.Credentials = stscreds.NewCredentials(f.RootSession.Copy(&config), f.Role)
	}

	return elasticsearchservice.New(f.RootSession.Copy(&config))
}

// Returns a cached client or instantiates a new client and caches it
func (f *ElasticsearchClientFactory) Get(region string) *elasticsearchservice.ElasticsearchService {
	// read lock to check client cache
	client := f.cached(region)
	if client != nil {
		return client
	}

	// write lock to construct a new client (if necessary)
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.clients == nil {
		f.clients = make(map[string]*elasticsearchservice.ElasticsearchService)
	}

	if client = f.clients[region]; client == nil {
		client = f.New(region)
	}

	f.clients[region] = client
	return client
}

func (f *ElasticsearchClientFactory) cached(region string) *elasticsearchservice.ElasticsearchService {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if f.clients == nil {
		return nil
	}

	return f.clients[region]
}
