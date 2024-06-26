# Zerops GOlang probe!!

```yaml
project:
  name: zerops-golang-probe
  tags:
    - zerops
    - probe

services:
  - hostname: nginx
    type: go@1
    ports:
      - port: 8080
        httpSupport: true
      - port: 8081
        httpSupport: true
      - port: 8082
        httpSupport: true
    enableSubdomainAccess: true
    buildFromGit: https://github.com/jansaidl/golang-probe
  - hostname: hey
    type: go@1
    ports:
      - port: 8080
        httpSupport: true
    enableSubdomainAccess: true
    buildFromGit: https://github.com/jansaidl/golang-probe
  - hostname: backend
    type: go@1
    ports:
      - port: 8080
        httpSupport: true
    enableSubdomainAccess: true
    buildFromGit: https://github.com/jansaidl/golang-probe
  - hostname: apibuild
    type: go@1
    ports:
      - port: 8080
        httpSupport: true
    enableSubdomainAccess: true
    buildFromGit: https://github.com/jansaidl/golang-probe
  - hostname: api
    type: go@1
    ports:
      - port: 8080
        httpSupport: true
    enableSubdomainAccess: true
    buildFromGit: https://github.com/jansaidl/golang-probe
  - hostname: test2
    type: go@1
    ports:
      - port: 8080
        httpSupport: true
    enableSubdomainAccess: true
    buildFromGit: https://github.com/jansaidl/golang-probe
```
