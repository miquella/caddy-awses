package awses_test

import (
	"reflect"
	"testing"

	"github.com/caddyserver/caddy"
	"github.com/miquella/caddy-awses"
)

type TestCase struct {
	Caddyfile string
	Configs   []awses.Config
}

var TestCases = map[string]TestCase{
	"Basic": TestCase{
		Caddyfile: `
			awses
		`,
		Configs: []awses.Config{
			{
				Path: "",
			},
		},
	},

	"Prefix": TestCase{
		Caddyfile: `
			awses /with/prefix
		`,
		Configs: []awses.Config{
			{
				Path: "/with/prefix",
			},
		},
	},

	"Domain": TestCase{
		Caddyfile: `
			awses {
				domain some-domain
			}
		`,
		Configs: []awses.Config{
			{
				Domain: "some-domain",
			},
		},
	},

	"Region": TestCase{
		Caddyfile: `
			awses {
				region ap-northeast-1
			}
		`,
		Configs: []awses.Config{
			{
				Region: "ap-northeast-1",
			},
		},
	},

	"Role": TestCase{
		Caddyfile: `
			awses {
				role arn:aws:iam::123456789012:role/some-role
			}
		`,
		Configs: []awses.Config{
			{
				Role: "arn:aws:iam::123456789012:role/some-role",
			},
		},
	},

	"Full": TestCase{
		Caddyfile: `
			awses /a/prefix/ {
				domain a-domain
				region ap-southeast-2
				role arn:aws:iam::123456789012:role/xacct
			}
		`,
		Configs: []awses.Config{
			{
				Path:   "/a/prefix",
				Domain: "a-domain",
				Region: "ap-southeast-2",
				Role:   "arn:aws:iam::123456789012:role/xacct",
			},
		},
	},

	"Multi": TestCase{
		Caddyfile: `
			awses /middle {
				domain middle
			}

			awses /longest {
				region us-east-1
			}

			awses /last {
				role arn:aws:iam::123456789012:role/last
			}
		`,
		Configs: []awses.Config{
			{
				Path:   "/longest",
				Region: "us-east-1",
			},
			{
				Path:   "/middle",
				Domain: "middle",
			},
			{
				Path: "/last",
				Role: "arn:aws:iam::123456789012:role/last",
			},
		},
	},
}

func TestParseConfigs(t *testing.T) {
	for key, testCase := range TestCases {
		controller := caddy.NewTestController("", testCase.Caddyfile)
		configs, err := awses.ParseConfigs(controller)
		if err != nil {
			t.Errorf("Failed to parse '%s' Caddyfile: %v\n%s", key, err, testCase.Caddyfile)
			continue
		}

		if len(configs) != len(testCase.Configs) {
			t.Errorf("Wrong number of configs parsed: %d, expected %d", len(configs), len(testCase.Configs))
			continue
		}

		for i := range testCase.Configs {
			if !reflect.DeepEqual(*configs[i], testCase.Configs[i]) {
				t.Errorf("Incorrect config:\n%#v\nexpected:\n%#v", *configs[i], testCase.Configs[i])
			}
		}
	}
}
