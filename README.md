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

Syntax
------

In the simplest form `awses` is enabled with:

```Caddyfile
awses
```

However, there are several advanced features that can be utilized with the
expanded syntax:

```Caddyfile
awses [/prefix] {
  domain <DOMAIN>
  region <REGION>
  role <ROLE_ARN>
}
```

* **`/prefix`** is the prefix of the path that must match in order for `awses`
  to handle the request. The prefix is always considered to be a full path
  segment, even when not ended with a slash (`/`) and will not match a partial
  path segment. Defaults to `/`.
* **`domain`** is the name of the AWS ES domain to proxy requests to. If not
  set, it will be derived from the path.
* **`region`** is the AWS region that will be searched to locate the domain.
  If not set, it will be derived from the path.
* **`role`** is the AWS IAM role to assume before signing requests. This can be
  useful to assume a role that has necessary permissions to access the domain
  or can be used for cross-account access of a domain. By default, no role is
  assumed.

URLs
----

Unless otherwise configured, the `region` and `domain` are derived from the
path of the request URL.

For example:

 * Request `/us-west-2/my-es/_cluster/health?pretty` proxies as:
   * Destination: `/_cluster/health?pretty`
   * Domain: `my-es`
   * Region: `us-west-2`

If `domain` were configured as `bob`:

 * Request `/us-west-2/_cluster/health?pretty` proxies as:
   * Destination: `/_cluster/health?pretty`
   * Domain: `bob`
   * Region: `us-west-2`

If `region` were configured as `ap-south-1`:

 * Request `/timmy/_cluster/health?pretty` proxies as:
   * Destination: `/_cluster/health?pretty`
   * Domain: `timmy`
   * Region: `ap-south-1`

If `domain` were configured as `bob` and `region` as `eu-west-1`:

 * Request `/_cluster/health?pretty` proxies as:
   * Destination: `/_cluster/health?pretty`
   * Domain: `bob`
   * Region: `eu-west-1`

Examples
--------

To enable proxying to all AWS ES domains in all regions, a simple
configuration may be used:

```Caddyfile
awses
```

A prefix can also be specified:

```Caddyfile
awses /logs/
```

Multiple configurations may be specified (although only one should be specified
for any given prefix):

```Caddyfile
awses /regions/

awses /logs/ {
  domain logs
}

awses /other-account/ {
  region us-east-2
  role arn:aws:iam::123456789012:role/elasticsearch-logs-us-east-2
}
```

[AWS Elasticsearch]: https://aws.amazon.com/documentation/elasticsearch-service/
[Caddy]: https://github.com/mholt/caddy
