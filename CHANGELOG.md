## 1.0.4 (July 11, 2022)
BUG FIXES:

* updated to latest shared provider (internal ticket)
* update makefile to be consistent with other providers. Still doesn't do version substitution correctly PR: [#47](https://github.com/jfrog/terraform-provider-project/pull/47)

## 1.0.3 (July 1, 2022)

BUG FIXES:

* provider: Fix hardcoded HTTP user-agent string. PR: [#8](https://github.com/jfrog/terraform-provider-pipeline/pull/8)

## 1.0.2 (June 21, 2022)

IMPROVEMENTS:

* Bump shared module version

## 1.0.1 (May 27, 2022)

IMPROVEMENTS:

* Upgrade `gopkg.in/yaml.v3` to v3.0.0 for [CVE-2022-28948](https://nvd.nist.gov/vuln/detail/CVE-2022-28948) PR [#4](https://github.com/jfrog/terraform-provider-pipeline/pull/4)

## 1.0.0 (May 16, 2022)

IMPROVEMENTS:

* Initial release to Terraform registry
  * Use code from terraform-provider-shared module
  * Add documentation
  * Add telemetry

PR [#3](https://github.com/jfrog/terraform-provider-pipeline/pull/3)
