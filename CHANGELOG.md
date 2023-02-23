## 1.2.2 (February 23, 2023)

SECURITY:
* provider:
  * Bumps golang.org/x/net from 0.0.0-20211029224645-99673261e6eb to 0.7.0. PR [#21](https://github.com/jfrog/terraform-provider-pipeline/pull/21)
  * Bump golang.org/x/crypto from 0.0.0-20210921155107-089bfa567519 to 0.1.0. PR [#22](https://github.com/jfrog/terraform-provider-pipeline/pull/22)

## 1.2.1 (February 23, 2023)

SECURITY:
* provider: Bumps golang.org/x/text from 0.3.7 to 0.3.8. PR [#20](https://github.com/jfrog/terraform-provider-pipeline/pull/20)

## 1.2.0 (December 6, 2022)

BUG FIXES:
* data source/pipeline_project: Fix crash when no Project is found.
* resource/pipeline_project_integration: Fix crashes when either `key` or `name` attributes are not specified for `project` attribute. The `project` attribute changes from a map to a set (of 1 item) to allow for schema validation.

Issue [#18](https://github.com/jfrog/terraform-provider-pipeline/issues/18) PR: [#19](https://github.com/jfrog/terraform-provider-pipeline/pull/19)

## 1.1.0 (November 30, 2022)

IMPROVEMENTS:

* resource/pipeline_project_integration: Add support for specifying if a form JSON values from Pipeline integration contains sensitive value or not. This also fixes state drift issue. PR: [#16](https://github.com/jfrog/terraform-provider-pipeline/pull/16)

## 1.0.5 (August 9, 2022)

IMPROVEMENTS:

* Update package `github.com/Masterminds/goutils` to 1.1.1 for [Dependeabot alert](https://github.com/jfrog/terraform-provider-pipeline/security/dependabot/3)

## 1.0.4 (July 11, 2022)

BUG FIXES:

* Updated makefile to be consistent with others. Issue: [#10](https://github.com/jfrog/terraform-provider-pipeline/issues/10) PR: [#13](https://github.com/jfrog/terraform-provider-pipeline/pull/13)

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
