package awses

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/caddyserver/caddy"
)

var (
	ErrTooManyArgs  = errors.New("[awses] too many arguments provided")
	ErrSingleDomain = errors.New("[awses] a single domain must be provided for the domain directive")
	ErrSingleRegion = errors.New("[awses] a single region must be provided for the region directive")
	ErrSingleRole   = errors.New("[awses] a single role must be provided for the role directive")
)

type Config struct {
	Path string

	Role   string
	Region string

	Domain string
}

func ParseConfigs(c *caddy.Controller) ([]*Config, error) {
	var configs []*Config
	for c.Next() {
		config := &Config{}
		configs = append(configs, config)

		// handle args
		configArgs := c.RemainingArgs()
		switch len(configArgs) {
		case 1:
			if strings.Trim(configArgs[0], "/") == "" {
				config.Path = ""
			} else {
				config.Path = "/" + strings.Trim(configArgs[0], "/")
			}

		case 0:
			config.Path = ""

		default:
			return nil, ErrTooManyArgs
		}

		// handle block directives
		for c.NextBlock() {
			directive := c.Val()
			args := c.RemainingArgs()

			switch directive {
			case "domain":
				if len(args) != 1 {
					return nil, ErrSingleDomain
				}
				config.Domain = args[0]

			case "region":
				if len(args) != 1 {
					return nil, ErrSingleRegion
				}
				config.Region = args[0]

			case "role":
				if len(args) != 1 {
					return nil, ErrSingleRole
				}
				config.Role = args[0]

			default:
				return nil, fmt.Errorf("[awses] invalid directive '%s'", c.Val())
			}
		}
	}

	sortedConfigs := sortableConfigs(configs)
	sort.Stable(sortedConfigs)

	return sortedConfigs, nil
}

type sortableConfigs []*Config

func (c sortableConfigs) Len() int {
	return len(c)
}

func (c sortableConfigs) Less(i, j int) bool {
	return len(c[i].Path) > len(c[j].Path)
}

func (c sortableConfigs) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
