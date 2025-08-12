## [3.2.2](https://github.com/Levy-Tal/site-availability/compare/v3.2.1...v3.2.2) (2025-08-12)


### Bug Fixes

* **site:** bug origin url is not currect ([0945060](https://github.com/Levy-Tal/site-availability/commit/0945060ca4a968584c1558b31dbca7603c156cab))

## [3.2.1](https://github.com/Levy-Tal/site-availability/compare/v3.2.0...v3.2.1) (2025-08-11)


### Bug Fixes

* **logo:** format from png to svg ([de1981b](https://github.com/Levy-Tal/site-availability/commit/de1981b89a963833f1836a5fff63c41d5d5abfbc))

# [3.2.0](https://github.com/Levy-Tal/site-availability/compare/v3.1.0...v3.2.0) (2025-08-10)


### Features

* **ui:** add spinner on loading ([4dab3ce](https://github.com/Levy-Tal/site-availability/commit/4dab3cec7153ec609a1fbc0637c6de1c65f19812))

# [3.1.0](https://github.com/Levy-Tal/site-availability/compare/v3.0.1...v3.1.0) (2025-08-10)


### Bug Fixes

* **style:** lable ([3f4bbfe](https://github.com/Levy-Tal/site-availability/commit/3f4bbfee6e9b27acd1041ab0af0a5eecd93f0ae2))


### Features

* **lables:** add to app status ([6b867cc](https://github.com/Levy-Tal/site-availability/commit/6b867ccee04a7c2f6a0d3abb8712c2d9a99e1d6b))
* **lables:** show lables by app ([b6d2ca3](https://github.com/Levy-Tal/site-availability/commit/b6d2ca37052648d2e805bcf6012df2df6bbdd325))

## [3.0.1](https://github.com/Levy-Tal/site-availability/compare/v3.0.0...v3.0.1) (2025-08-07)


### Bug Fixes

* **helm:** helm chart dashboard ([2aaead2](https://github.com/Levy-Tal/site-availability/commit/2aaead2e5752e1aa5a5eddea2ba766384f3220d1))

# [3.0.0](https://github.com/Levy-Tal/site-availability/compare/v2.4.1...v3.0.0) (2025-08-07)


### Bug Fixes

* **metrics:** add app lables to metrics ([fe1fa51](https://github.com/Levy-Tal/site-availability/commit/fe1fa519463ab47778983f21dd78c863bf753922))
* **oidc:** bug when local user was disabled ([c636d7a](https://github.com/Levy-Tal/site-availability/commit/c636d7aa92601fe84707482004685e8dbbd0f965))
* **oidc:** bug when local user was disabled ([df0a259](https://github.com/Levy-Tal/site-availability/commit/df0a2590926994203c7185b8da85bd177f9e6fb4))
* **oidc:** bugs ([3f08ad6](https://github.com/Levy-Tal/site-availability/commit/3f08ad6bffbe80919579bab2a473abc3fa2adf66))
* **proxy:** trust proxy headers ([477c92b](https://github.com/Levy-Tal/site-availability/commit/477c92b6f8889b013f052619c68833aaaa8b43e3))
* **proxy:** trust proxy headers ([7b8cf6d](https://github.com/Levy-Tal/site-availability/commit/7b8cf6d1b7472c4195a1ac19050c4c3bdd95279e))
* **proxy:** trust proxy headers ([272b38e](https://github.com/Levy-Tal/site-availability/commit/272b38e8596a072d2595ebd57944a0bbe523702e))
* Refactor Sidebar.js to properly handle key suggestions and selected labels ([fb46a41](https://github.com/Levy-Tal/site-availability/commit/fb46a41103898df7f22e239b4cfb7af608b9b663))
* **security:** bugs ([91fbe48](https://github.com/Levy-Tal/site-availability/commit/91fbe481780bac9c859ff4a67bf94b89e1bfac80))
* **security:** fix bug ([55f692a](https://github.com/Levy-Tal/site-availability/commit/55f692a88eed33a78a867e5c3c26605c901fab7d))
* **simple:** bugs ([f1a4378](https://github.com/Levy-Tal/site-availability/commit/f1a43780e81094546826ffcd68c6443c4c4c309f))
* **style:** colors ([eca7f54](https://github.com/Levy-Tal/site-availability/commit/eca7f540ae584562ce268d3518b2a6a52457b470))
* **test:** failed tests ([dfc4184](https://github.com/Levy-Tal/site-availability/commit/dfc41843f73f17d259ba8a527222ed86c033bd3a))


### Code Refactoring

* **lables:** change lables from struct to map ([b3b4d95](https://github.com/Levy-Tal/site-availability/commit/b3b4d955f7b23904300f53f6a16dd24db15aa367))


### Features

* **helm:** add improvments to the chart ([02cc2ef](https://github.com/Levy-Tal/site-availability/commit/02cc2efaf9b11835ee11eed239d672845768e256))
* **login:** add admin-local login rc1 ([606f25d](https://github.com/Levy-Tal/site-availability/commit/606f25d5f347f49ed1ddd5a6fc0cfb9b98e32440))
* **login:** add oidc login rc3 ([555d6e6](https://github.com/Levy-Tal/site-availability/commit/555d6e66f99a02be724087bc885ff590765dc7f7))
* **login:** add oidc login- working rc4 ([cdf70e9](https://github.com/Levy-Tal/site-availability/commit/cdf70e928f9ca497eb09c3a1c3d67e9b1794e09b))
* **metrics:** add auth login ([dd1a723](https://github.com/Levy-Tal/site-availability/commit/dd1a7235091b37a53b499097325d3c2712d3c6ce))
* **status:** add numbers nextto checkbox ([236a9a3](https://github.com/Levy-Tal/site-availability/commit/236a9a3455ed454c4bdf50be1e2b01f537e3ab82))


### BREAKING CHANGES

* **lables:** schema change

# [Unreleased]

### Features

- **Metrics Authentication**: Added authentication support for the `/metrics` endpoint
  - Support for basic authentication (username/password)
  - Support for bearer token authentication
  - Configurable via `server_settings.metrics_auth` in config.yaml or credentials.yaml
  - Default behavior remains unauthenticated (backward compatible)
  - Proper HTTP 401 responses with WWW-Authenticate headers
  - Comprehensive test coverage for all authentication scenarios

### Documentation

- Added comprehensive documentation for metrics authentication configuration
- Updated Prometheus configuration examples with authentication
- Added troubleshooting guide for metrics authentication issues
- Updated server configuration documentation with new metrics_auth section

## [2.4.1](https://github.com/Levy-Tal/site-availability/compare/v2.4.0...v2.4.1) (2025-07-28)

### Bug Fixes

- **sync:** fix locations and apps sync ([bc78f30](https://github.com/Levy-Tal/site-availability/commit/bc78f30969ced08bf88d0ce0871cb539352c3589))

# [2.4.0](https://github.com/Levy-Tal/site-availability/compare/v2.3.1...v2.4.0) (2025-07-13)

### Features

- **localstorage:** save user filters on the ui ([3205fb5](https://github.com/Levy-Tal/site-availability/commit/3205fb5cf8a983ac78d55ae5b3e9d86b4ed4c34e))

## [2.3.1](https://github.com/Levy-Tal/site-availability/compare/v2.3.0...v2.3.1) (2025-07-13)

### Bug Fixes

- **ui:** fix ui bugs ([1d66f59](https://github.com/Levy-Tal/site-availability/commit/1d66f59d60c61ac1cd078e5787ed6a4fb2d6458b))

# [2.3.0](https://github.com/Levy-Tal/site-availability/compare/v2.2.0...v2.3.0) (2025-07-09)

### Security Fixes

- **OIDC Callback URL Vulnerability**: Fixed Host header injection vulnerability in OIDC callback URL generation
  - Added required `host_url` field to `server_settings` configuration
  - OIDC callback URLs now use trusted configuration instead of unvalidated request headers
  - Prevents potential authorization code theft through malicious redirects
  - Ensures correct protocol (http/https) handling for OIDC authentication

### Breaking Changes

- **Configuration**: The `host_url` field is now required in `server_settings`
  - Must include scheme (http:// or https://) and host
  - Used to construct secure OIDC callback URLs
  - Format: `https://myserver.com` or `http://localhost:8080`

### Documentation

- Updated all configuration examples to include the required `host_url` field
- Added troubleshooting section for host_url configuration issues
- Updated OIDC setup guides with new callback URL format

# [2.3.0](https://github.com/Levy-Tal/site-availability/compare/v2.2.0...v2.3.0) (2025-07-09)

### Features

- **group by lable:** Add group by filter in the app status panel ([5896d0c](https://github.com/Levy-Tal/site-availability/commit/5896d0cfa82732e57bd78be11f0bc99cc537958e))

# [2.2.0](https://github.com/Levy-Tal/site-availability/compare/v2.1.1...v2.2.0) (2025-07-09)

### Bug Fixes

- **docs:** ui show the docs url and title ([8a5680b](https://github.com/Levy-Tal/site-availability/commit/8a5680bdb12729a1a8b35a076bafcc56c1a0e699))

### Features

- **http:** add http source ([38ae7e1](https://github.com/Levy-Tal/site-availability/commit/38ae7e1fa079a323c7ceeeeae09a0486635896d3))
- **source:** new source architecture ([8747ef9](https://github.com/Levy-Tal/site-availability/commit/8747ef92e136ffd967bda9d1433e063524e61cf0))

## [2.1.1](https://github.com/Levy-Tal/site-availability/compare/v2.1.0...v2.1.1) (2025-07-03)

### Bug Fixes

- **ca_cert:** fix ca_cert ([9655175](https://github.com/Levy-Tal/site-availability/commit/965517529c127681a41c25cc560eecfa4ddc24be))
- **ca_cert:** fix ca_cert ([4e21ab2](https://github.com/Levy-Tal/site-availability/commit/4e21ab2f1c6da8d5ff2e449095398cb2d850e2e7))
- **ca:** fix ca_cert not loading ([dce3ee4](https://github.com/Levy-Tal/site-availability/commit/dce3ee4f94a8ea53313d2e805ace5f80f4c1b19d))
- **ca:** fix ca_cert not loading ([67ccc85](https://github.com/Levy-Tal/site-availability/commit/67ccc85a205ce2ad986961d5b59f76e4a070f7cd))

# [2.1.0](https://github.com/Levy-Tal/site-availability/compare/v2.0.1...v2.1.0) (2025-07-02)

### Bug Fixes

- **lables:** not showing app status ([df3cec4](https://github.com/Levy-Tal/site-availability/commit/df3cec491cb2281ef7788c29a231a5e4284b8431))
- **lables:** remove metrics from files ([c371851](https://github.com/Levy-Tal/site-availability/commit/c371851adfcfee322e5bcd4cec05aa2b0a3aa69c))

### Features

- **api:** search locations and lables ([9b2967e](https://github.com/Levy-Tal/site-availability/commit/9b2967ea0bcc458030d238942ff3a23ba270bc79))
- **labels:** add labels to apps ([3126de9](https://github.com/Levy-Tal/site-availability/commit/3126de9a89096aa8ec3ec4f1ec53a4f0aed25bbb))

## [2.0.1](https://github.com/Levy-Tal/site-availability/compare/v2.0.0...v2.0.1) (2025-06-28)

### Bug Fixes

- **location:** locations from other sites are not showing ([4c2a7f1](https://github.com/Levy-Tal/site-availability/commit/4c2a7f15fbb8954bbed33f1c46d2c13ae44e0dca))

# [2.0.0](https://github.com/Levy-Tal/site-availability/compare/v1.6.0...v2.0.0) (2025-06-26)

- Merge pull request [#2](https://github.com/Levy-Tal/site-availability/issues/2) from Levy-Tal/feat(site-sync) ([612deca](https://github.com/Levy-Tal/site-availability/commit/612deca21e5946c5a3605baf070dc9489b9fd140))

### Bug Fixes

- **handlers_test.go:** I've updated another test to not rely on the order of items in the cache. ([63cff8a](https://github.com/Levy-Tal/site-availability/commit/63cff8ad841ffb658f31b927d1d57e11ca8d14d9))
- **handlers_test.go:** I've updated the test to not rely on the order of items in the cache. ([d64b515](https://github.com/Levy-Tal/site-availability/commit/d64b515bf46a5f3c09b1b0c7e7a3a30bb203aafc))

### Features

- **site:** site scraping ([6354381](https://github.com/Levy-Tal/site-availability/commit/6354381ecb9ddb41203ce1aa2ad6b774a7a83df5))
- **sync:** Add the ability to sync status from other sites ([bf3ed51](https://github.com/Levy-Tal/site-availability/commit/bf3ed51e7039e3e8de57570984378c99f20737ae))

### BREAKING CHANGES

- New project structure.
- **site:** New project structure.

# [1.6.0](https://github.com/Levy-Tal/site-availability/compare/v1.5.0...v1.6.0) (2025-06-08)

### Features

- **scrape:** add the option to scrape prometheus with authentication ([c67852a](https://github.com/Levy-Tal/site-availability/commit/c67852a8e35acfa49a0598aac485b8716c162a12))

# [1.5.0](https://github.com/Levy-Tal/site-availability/compare/v1.4.11...v1.5.0) (2025-06-08)

### Features

- **pre-commit:** add ([7eacda2](https://github.com/Levy-Tal/site-availability/commit/7eacda25f9d8e7812f0c1e8c1ee7b8d8bcd80f76))

## [1.4.11](https://github.com/Levy-Tal/site-availability/compare/v1.4.10...v1.4.11) (2025-06-07)

### Bug Fixes

- **helm:** security ([d6bce5f](https://github.com/Levy-Tal/site-availability/commit/d6bce5f1b6951a6ab42e0a6d1a50efd3a6e73b3f))

## [1.4.10](https://github.com/Levy-Tal/site-availability/compare/v1.4.9...v1.4.10) (2025-06-05)

### Bug Fixes

- **release:** typo ([5e7e24a](https://github.com/Levy-Tal/site-availability/commit/5e7e24ab9a942b1a8a24dfd37ec92d67d158e771))

## [1.4.9](https://github.com/Levy-Tal/site-availability/compare/v1.4.8...v1.4.9) (2025-06-05)

### Bug Fixes

- **release:** typo ([35ac5e4](https://github.com/Levy-Tal/site-availability/commit/35ac5e485c2ca1cfcfa7dfeedd41910f3e004653))

## [1.4.8](https://github.com/Levy-Tal/site-availability/compare/v1.4.7...v1.4.8) (2025-06-05)

### Bug Fixes

- **release:** typo ([1f43b42](https://github.com/Levy-Tal/site-availability/commit/1f43b42197079def51932b280c9a83c537cd4779))
- **release:** typo ([54cc3ff](https://github.com/Levy-Tal/site-availability/commit/54cc3ff87bbf18c7479edd41b920dd7c449e0cec))
- **release:** workflow ([5486082](https://github.com/Levy-Tal/site-availability/commit/5486082b0720320629983bee4593319f1e2a9af4))

## [1.4.7](https://github.com/Levy-Tal/site-availability/compare/v1.4.6...v1.4.7) (2025-06-04)

### Bug Fixes

- **docker:** workflow ([252c316](https://github.com/Levy-Tal/site-availability/commit/252c316f1045cb16633f714720a8a15b68d290d8))

## [1.4.6](https://github.com/Levy-Tal/site-availability/compare/v1.4.5...v1.4.6) (2025-06-04)

### Bug Fixes

- **semantic-release:** not pushing files to git ([eab713e](https://github.com/Levy-Tal/site-availability/commit/eab713e3b8c1c134b796adcec9ae6977d2b48a8f))

## [1.4.3](https://github.com/Levy-Tal/site-availability/compare/v1.4.2...v1.4.3) (2025-06-04)

### Bug Fixes

- **semantic-release:** file name in release ([0ceefb4](https://github.com/Levy-Tal/site-availability/commit/0ceefb49ce7c9c0ce80264f658c11a8699ca1861))

## [1.4.2](https://github.com/Levy-Tal/site-availability/compare/v1.4.1...v1.4.2) (2025-06-04)

### Bug Fixes

- **semantic-release:** file permissions ([0336baf](https://github.com/Levy-Tal/site-availability/commit/0336bafe747ee36359dca5d95c4ffd955d4c88bd))

## [1.4.1](https://github.com/Levy-Tal/site-availability/compare/v1.4.0...v1.4.1) (2025-06-04)

### Bug Fixes

- **semantic-release:** add semantic release for github release ([9cad75b](https://github.com/Levy-Tal/site-availability/commit/9cad75bdf6177a5063fc5ce4df41e8f1ccddae3b))

## [1.4.1](https://github.com/Levy-Tal/site-availability/compare/v1.4.0...v1.4.1) (2025-06-04)

### Bug Fixes

- **semantic-release:** add semantic release for github release ([9cad75b](https://github.com/Levy-Tal/site-availability/commit/9cad75bdf6177a5063fc5ce4df41e8f1ccddae3b))

## [1.4.1](https://github.com/Levy-Tal/site-availability/compare/v1.4.0...v1.4.1) (2025-06-04)

### Bug Fixes

- **semantic-release:** add semantic release for github release ([9cad75b](https://github.com/Levy-Tal/site-availability/commit/9cad75bdf6177a5063fc5ce4df41e8f1ccddae3b))

# [1.4.0](https://github.com/Levy-Tal/site-availability/compare/v1.3.0...v1.4.0) (2025-06-04)

### Features

- add automated status reports generation ([61308fe](https://github.com/Levy-Tal/site-availability/commit/61308feba138d050c2d102f6a333e7584f1d8dec))
- add semantic release ([23bb615](https://github.com/Levy-Tal/site-availability/commit/23bb6155c6b37723b3f0ab72c9457ee2009f9d4e))
