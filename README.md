# Terraform Provider for Artifactory Pipelines

## Documentation

To use this provider in your Terraform module, follow the documentation on [Terraform Registry](https://registry.terraform.io/providers/jfrog/pipeline/latest/docs).

## License requirements

This provider requires access to Artifactory APIs, which are only available in the _licensed_ enterprise plus editions. You can determine which license you have by accessing the following URL `${host}/artifactory/api/system/licenses/`

You can either access it via API, or web browser - it requires admin level credentials, but it's one of the few APIs that will work without a license (side node: you can also install your license here with a `POST`)

```sh
$ curl -sL ${host}/artifactory/api/system/licenses/ | jq .
```

```js
{
  "type" : "Enterprise Plus Trial",
  "validThrough" : "Jan 29, 2022",
  "licensedTo" : "JFrog Ltd"
}
```

The following 3 license types (`jq .type`) do **NOT** support APIs:
- Community Edition for C/C++
- JCR Edition
- OSS

## Versioning

In general, this project follows [Terraform Versioning Specification](https://www.terraform.io/plugin/sdkv2/best-practices/versioning#versioning-specification) as closely as we can for tagging releases of the package.

## Contributors

See the [contribution guide](https://github.com/jfrog/terraform-provider-pipeline/blob/master/CONTRIBUTIONS.md).

## License

Copyright (c) 2022 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
