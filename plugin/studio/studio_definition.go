// Package studio defines the common functionality required by each
// of the UiPath Studio command plugins like analyze, pack, publish
// restore and test run.
package studio

const StudioDefinition = `
openapi: 3.0.1
info:
  title: UiPath Studio
  description: UiPath Studio
  version: v1
servers:
  - url: https://cloud.uipath.com/{organization}/studio_/backend
    description: The production url
    variables:
      organization:
        description: The organization name (or id)
        default: my-org
paths:
  {}
`
