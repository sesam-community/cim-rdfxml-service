# Minimal Microservice Scaffold in Golang / Go.

  Example simple microservice project template providing:

  - **quick bootstrap of Golang microservice development**,
  - useful project structure,
  - some best practices for microservices,
  - `ginkgo` BDD-style unit and integration tests,
  - optional `.config.json` initialisation file,
  - environment variable configurations,
  - `http.HandlerFunc` middleware compatible examples,
  - uses performant `httprouter` (HTTP-handling otherwise based on the standard Go libraries),
  - JSON stream processing of incoming data (array of JSON objects),
  - ETL schema-based and ELT schemaless JSON data transformation examples,
  - CI/CD scripts help ship microservice container artifact,
  - minimally sized container image from `scratch`,
  - single static executable as only possible entrypoint,
  - based on secure and lightweight _system call library_ `musl-libc` (**Alpine Linux**),
  - includes _ca-certificates_ for Golang HTTPS/TLS connectivity.

## Runtime configuration

  * `/.config.json` is an empty optional configuration file which is included into the Docker build.

## Editor integration

 - It is recommended to use the `gopls` Golang Language Server when working with Golang files.
   The Language Server was introduced in summer 2019 and is still in alpha state by October 2019, but is already very usable, efficient and works well with **Golang modules**.

### VSCode - Visual Studio Code

  * `.vscode/tasks.json` includes a default build task (ctrl+b) which executes `.ci/build-docker.sh` script.

## Build system

  - Build scripts in `.ci/` will build the microservice static executable using **Alpine Linux** for a more secure system calls library (`musl-libc`) and being very lightweight with minimal bloat to the executable.

### Docker image

  * microservice executable placed as `/service` in Docker image,
  * single entrypoint for container image,
  * `Dockerfile` can accept some arguments via the build environment (but must otherwise must be edited for e.g static resources).


## See also

  * https://onsi.github.io/ginkgo/
  * https://github.com/julienschmidt/httprouter
  * https://godoc.org/github.com/julienschmidt/httprouter
  * https://github.com/antchfx/jsonquery
  * https://github.com/json-iterator/go
