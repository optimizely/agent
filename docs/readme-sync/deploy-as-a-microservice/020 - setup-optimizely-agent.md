---
title: "Set up Optimizely Agent"
excerpt: ""
slug: "setup-optimizely-agent"
hidden: false
metadata: 
  title: "Getting started with Agent - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:27.363Z"
updatedAt: "2020-03-31T23:54:17.841Z"
---
## Running Agent from source (Linux / OSX)

To develop and compile Optimizely Agent from source:

1. Install  [Golang](https://golang.org/dl/)  version 1.13+ .
2. Clone the [Optimizely Agent repo](https://github.com/optimizely/agent). 
3. From the repo directory, open a terminal and start Optimizely Agent:

```bash
make setup
```
Then
```bash
make run
```

This starts the Optimizely Agent with the default configuration in the foreground.

## Running Agent from source (Windows)

You can use a [helper script](https://github.com/optimizely/agent/blob/master/scripts/build.ps1) to install prerequisites (Golang, Git) and compile agent in a Windows environment. Take these steps:

1.  Clone the [Optimizely Agent repo](https://github.com/optimizely/agent)
2. From the repo directory, open a Powershell terminal and run 

```bash
Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser

.\scripts\build.ps1

.\bin\optimizely.exe
```

## Running Agent via Docker

If you have Docker installed, you can start Optimizely Agent as a container. Take these steps:

1. Pull the Docker image:

```bash
docker pull optimizely/agent
```
By default this will pull the "latest" tag. You can also specify a specific version of Agent by providing the version as a tag to the docker command:

```bash
docker pull optimizely/agent:X.Y.Z
```

2. Run the docker container with:

```bash
docker run -p 8080:8080 optimizely/agent
```
This will start Agent in the foreground and expose the container API port 8080 to the host.

3. (Optional) You can alter the configuration by passing in environment variables to the preceding command, without having to create a config.yaml file. See [configure optimizely agent](doc:configure-optimizely-agent) for more options.

Versioning:
When a new version is released, 2 images are pushed to dockerhub. They are distinguished by their tags:
- :latest (same as :X.Y.Z)
- :alpine (same as :X.Y.Z-alpine)

The difference between latest and alpine is that latest is built `FROM scratch` while alpine is `FROM alpine`.
- [latest Dockerfile](https://github.com/optimizely/agent/blob/master/scripts/dockerfiles/Dockerfile.static)
- [alpine Dockerfile](https://github.com/optimizely/agent/blob/master/scripts/dockerfiles/Dockerfile.alpine)