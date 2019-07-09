caddy-awses
===========

[Caddy][Caddy] plugin for signing and proxying requests to
[AWS Elasticsearch][AWS Elasticsearch] (AWS ES).

Configuring access to an AWS ES domain can be tricky. The access policy of an
AWS ES domain is based on a principal (which necessitates a signed request) or
an IP address whitelist. Whitelisting IP addresses often isn't a viable option
and standard tools (such as `curl` or a browser) can't properly sign requests.

This is exactly the problem this plugin aims to address. Standard tools can
make unauthenticated requests to the Caddy server which are then signed and
proxied to the AWS ES service.


Getting Started
---------------

The simplest way to get started is by invoking `caddy` with the `awses`
directive, like so:

```sh
caddy awses
```

Or by adding the `awses` directive to your [`Caddyfile`][Caddyfile].


Syntax
------

```Caddyfile
awses [/prefix] {
    domain <DOMAIN>
    region <REGION>
    role <ROLE_ARN>
}
```

### `/prefix`

The prefix the path must match for `awses` to match and handle the request.
Defaults to `/`, matching all requests.

_Note: The prefix is always considered to be a full path segment. i.e. a prefix
of `/abc` **will not** match a request for `/abcdef`, but will match `/abc/def/`._

### `domain`

The name of the AWS ES domain to proxy requests to. Derived from the request path
unless set (see [URLs](#URLs) below).

_Note: `awses` will lookup the AWS ES domain endpoint automatically and should
not be provided._

### `region`

The AWS region containing the AWS ES domains to proxy for. Derived from the
request path unless set (see [URLs](#URLs) below).

### `role`

The AWS IAM role to assume to sign requests. This can be useful to assume a
role that has the permissions necessary to access the domain. This can also be
used for cross-account access of a domain. By default, no role is assumed.


Required Permissions
--------------------

For any AWS ES domain that `awses` proxies to, the following permission is
always required (to lookup the domain's endpoint):

 * `es:DescribeElasticsearchDomain`

Additionally, the following actions must be allowed for any method you intend
`awses` to proxy:

 * `es:ESHttpDelete`
 * `es:ESHttpGet`
 * `es:ESHttpHead`
 * `es:ESHttpPost`
 * `es:ESHttpPut`

Optionally, if no domain is specified the following permission can be granted
to get a list of available domains (within a region):

 * `es:ListDomainNames`


URLs
----

Requests to `awses` take the form:

`[/region][/domain]/<destination>`

If `region` and/or `domain` are specified in the configuration, they will not
be derived from the request path.

See [Examples](#Examples) below for more details.


Examples
--------

### All regions and domains

```Caddyfile
awses
```

Allows requests in the following form:

 * `/<region>/<domain>/<destination>`

### Specific region (all domains)

```Caddyfile
awses {
    region us-west-2
}
```

Allows requests in the following form:

 * `/<domain>/<destination>`

### Specific domain (all regions)

```Caddyfile
awses {
    domain es-logs
}
```

Allows requests in the following form:

 * `/<region>/<destination>`

### Specific region and domain

```Caddyfile
awses {
    region us-west-2
    domain es-logs
}
```

Allows requests in the following form:

 * `/<destination>`

### Multiple prefixes

```Caddyfile
awses /docs/ {
    region us-east-1
    domain the-docs
}

awses /logs/ {
    domain es-logs
}

awses /other-account/logs/ {
    domain es-logs
    role arn:aws:iam::123456789012:role/elasticsearch-logs-us-east-2
}
```

Allows requests in the following forms:

 * `/docs/<destination>`
 * `/logs/<region>/<destination>`
 * `/other-account/logs/<region>/<destination>`


Kibana
------

Please note that Kibana appears to have issues when hosted at a path other than
`/`, but I haven't had enough time to track down why that is just yet.

If you're looking to use Kibana through `awses`, the configuration will need to
omit the `/prefix` and will need to include `region` and `domain` parameters.
This will leave Kibana accessible at `/_plugin/kibana/`.


[AWS Elasticsearch]: https://aws.amazon.com/documentation/elasticsearch-service/
[Caddy]: https://github.com/caddyserver/caddy
[Caddyfile]: https://caddyserver.com/tutorial/caddyfile
