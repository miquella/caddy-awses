package awses

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
)

var (
	ErrDomainNotFound    = errors.New("AWS ES domain not found")
	ErrInvalidDomainName = errors.New("The provided AWS ES domain is invalid")
)

type elasticsearchDomain struct {
	region string
	domain string
}

type ElasticsearchManager struct {
	ClientFactory ElasticsearchClientFactory

	mutex   sync.RWMutex
	proxies map[elasticsearchDomain]*httputil.ReverseProxy
}

func NewElasticsearchManager(rootSession *session.Session, role string) *ElasticsearchManager {
	return &ElasticsearchManager{
		ClientFactory: ElasticsearchClientFactory{
			RootSession: rootSession,
			Role:        role,
		},
	}
}

func (m *ElasticsearchManager) ListDomains(region string) ([]string, error) {
	client := m.ClientFactory.Get(region)
	output, err := client.ListDomainNames(nil)
	if err != nil {
		return nil, err
	}

	domains := []string{}
	for _, domain := range output.DomainNames {
		domains = append(domains, *domain.DomainName)
	}
	sort.Strings(domains)
	return domains, nil
}

func (m *ElasticsearchManager) NewProxy(region, domain string) (*httputil.ReverseProxy, error) {
	// lookup the domain endpoint
	client := m.ClientFactory.Get(region)
	output, err := client.DescribeElasticsearchDomain(&elasticsearchservice.DescribeElasticsearchDomainInput{DomainName: &domain})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				return nil, ErrDomainNotFound
			} else if awsErr.Code() == "ValidationException" {
				return nil, ErrInvalidDomainName
			}
		}
		return nil, err
	} else if output.DomainStatus == nil || output.DomainStatus.Endpoint == nil {
		return nil, ErrDomainNotFound
	}

	// construct the reverse proxy
	signer := v4.NewSigner(client.Config.Credentials)
	endpointHost := *output.DomainStatus.Endpoint

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Rewrite the request
			req.Host = ""
			req.URL.Scheme = "https"
			req.URL.Host = endpointHost

			// if the client hasn't set a User-Agent, don't allow it to be set to the default value
			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "")
			}

			// why does the signing fail when we have "Connection: keep-alive"??
			req.Header.Set("Connection", "close")

			// Read the body
			var body io.ReadSeeker
			if req.Body != nil {
				defer req.Body.Close()
				bodyBytes, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return
				}

				body = bytes.NewReader(bodyBytes)
			}

			// Sign the request
			signer.Sign(req, body, "es", region, time.Now().Add(-10*time.Second))
		},
	}, nil
}

func (m *ElasticsearchManager) GetProxy(region, domain string) (*httputil.ReverseProxy, error) {
	// read lock to check proxy cache
	proxy := m.cachedProxy(region, domain)
	if proxy != nil {
		return proxy, nil
	}

	// write lock to construct a new proxy (if necessary)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := elasticsearchDomain{region: region, domain: domain}

	if m.proxies == nil {
		m.proxies = make(map[elasticsearchDomain]*httputil.ReverseProxy)
	}

	if proxy = m.proxies[key]; proxy == nil {
		var err error
		proxy, err = m.NewProxy(region, domain)
		if err != nil {
			return nil, err
		}
	}

	m.proxies[key] = proxy
	return proxy, nil
}

func (m *ElasticsearchManager) cachedProxy(region, domain string) *httputil.ReverseProxy {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.proxies == nil {
		return nil
	}

	return m.proxies[elasticsearchDomain{region: region, domain: domain}]
}
