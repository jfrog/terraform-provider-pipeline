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

Pull requests, issues and comments are welcomed. For pull requests:

* Add tests for new features and bug fixes
* Follow the existing style
* Separate unrelated changes into multiple pull requests

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating an issue and explaining the intended change.

JFrog requires contributors to sign a Contributor License Agreement, known as a CLA. This serves as a record stating that the contributor is entitled to contribute the code/documentation/translation to the project and is willing to have it used in distributions and derivative works (or is willing to transfer ownership).

## License

Copyright (c) 2022 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
