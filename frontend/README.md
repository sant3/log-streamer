## Requisites

Tested on **pnpm 10.24.0**

## How to run

`rm -rf node_modules`
`pnpm install`
`pnpm run start` in dev mode on port 3000 *http://localhost:3000*

## Build production

`pnpm run build`

## TODO

* config to enable/disable autocomplete
* add config for list server
  * add right panel to show server list 
* clean up code
* add some tests
* add support for basic auth
* complete support for jwt token
* add config for protocol (https|http)
* add config for auth mode (none|basic|jwt)
* add support for dark/light theme