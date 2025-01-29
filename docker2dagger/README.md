# Docker2Dagger
- https://github.com/samalba/dagger-modules/tree/main/docker2dagger

This module helps you convert any existing Dockerfile into a Dagger Module. It aims to ease the adoption of Dagger by starting a module from what you already know well.

Maintaining large Dockerfiles is hard, a Dagger module can replace a Dockerfile entirely with the same programming language you use in your app.

## How to use

This example assumes:
1. [You have Dagger already installed](https://docs.dagger.io/install/).
2. You have an env variable named `ANTHROPIC_API_KEY` [set to a valid Anthropic API key](https://docs.anthropic.com/en/docs/initial-setup).

First you need a to choose a Dockerfile to convert, we'll pick the project ["Getting Started Todo App"](https://github.com/docker/getting-started-todo-app) from Docker.

```terminal
git clone https://github.com/docker/getting-started-todo-app.git
```

Then we convert the Dockerfile into a new Dagger module:

```terminal
dagger -m github.com/samalba/dagger-modules/docker2dagger \
  call from-dockerfile \
  --anthropic-api-key env:ANTHROPIC_API_KEY \
  --dockerfile ./getting-started-todo-app/Dockerfile \
  export --path my-new-module
```

The command will be convert the existing Dockerfile from previously cloned repository and create a new Dagger module that you can use as-is or modify to your needs in a new `my-new-module` directory.

Here is how to build the todo app with Dagger, using the new generated module:

```terminal
dagger -m ./my-new-module \
  call build \
  --local-directory ./getting-started-todo-app
```

As-is, the Dagger build function returns a `dagger.Container` which is configured to expose the port 3000 as a service. We can reuse this container somewhere else in our Dagger pipeline, for example to do things like: publish the container to a registry, run a vulnerability scan, etc...

We can also start the app as a Dagger service by invoking `up` on the same container:

```terminal
dagger -m ./my-new-module \
  call build \
  --local-directory ./getting-started-todo-app \
  up
```

You can then access the running app at http://localhost:3000/
