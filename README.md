# Wallarm node Docker image (NGINX-based)

Wallarm end-to-end API Security protects websites, APIs and microservices from OWASP Top 10, bots and application abuse with no manual rule configuration and ultra-low false positives. This Docker image contains all subsystems of the NGINX-based Wallarm node to be deployed to your environment to protect it.

* [Image on Docker Hub](https://hub.docker.com/r/wallarm/node)
* [Detailed instructions for running the Docker container](https://docs.wallarm.com/admin-en/installation-docker-en/)

### Software requirements
* docker
* docker-compose
* gnu-sed (insted of `sed` util, just for MacOS)
  ```
    brew install gnu-sed
    export PATH="/opt/homebrew/opt/gnu-sed/libexec/gnubin:$PATH"
  
### Build
Use `make build` to build `wallarm/node:test` docker image.
### Tests
Use `make smoke-test` to run smoke tests.
