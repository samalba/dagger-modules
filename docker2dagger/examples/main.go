// A generated module for MyModule functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/my-module/internal/dagger"
	"fmt"
)

type MyModule struct{}

// Build a container from a Python application
//
// # Dockerfile used:
// FROM python:3.12
// WORKDIR /usr/local/app
// # Install the application dependencies
// COPY requirements.txt ./
// RUN pip install --no-cache-dir -r requirements.txt
// # Copy in the source code
// COPY src ./src
// EXPOSE 5000
// # Setup an app user so the container doesn't run as the root user
// RUN useradd app
// USER app
// CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
func (m *MyModule) Example1(localDirectory *dagger.Directory) *dagger.Container {
	// Dagger cannot access the local directory, so the directory on the host is passed as an argument to the function
	// Here is an example of calling the function with the current working directory passed as the argument:
	// dagger call example1 --local-directory .
	return dag.Container().
		From("python:3.12").
		WithWorkdir("/usr/local/app").
		WithFile("./requirements.txt", localDirectory.File("requirements.txt")).
		WithExec([]string{"pip", "nnstall", "--no-cache-dir", "-r", "requirements.txt"}).
		WithDirectory("./src", localDirectory.Directory("src")).
		WithExposedPort(5000).
		WithExec([]string{"useradd", "app"}).
		WithUser("app").
		WithDefaultArgs([]string{"uvicorn", "app.main:app", "--host", "8080"})
}

// Build a container from a Java application
//
// # Dockerfile used:
// # Stage 1: build stage
// FROM eclipse-temurin:21.0.2_13-jdk-jammy AS builder
// WORKDIR /opt/app
// COPY .mvn/ .mvn
// COPY mvnw pom.xml ./
// RUN ./mvnw dependency:go-offline
// COPY ./src ./src
// RUN ./mvnw clean install
// # Stage 2: final container which has some files copied from the build container
// FROM eclipse-temurin:21.0.2_13-jre-jammy AS final
// WORKDIR /opt/app
// EXPOSE 8080
// COPY --from=builder /opt/app/target/*.jar /opt/app/*.jar
// ENTRYPOINT ["java", "-jar", "/opt/app/*.jar"]
func (m *MyModule) Example2(localDirectory *dagger.Directory) *dagger.Container {
	// Dagger cannot access the local directory, so the directory on the host is passed as an argument to the function
	// Here is an example of calling the function with the current working directory passed as the argument:
	// dagger call example1 --local-directory .

	// Stage 1: Build Environment
	buildStage := dag.Container().
		From("eclipse-temurin:21.0.2_13-jdk-jammy").
		WithWorkdir("/opt/app").
		WithDirectory(".mvn/", localDirectory.Directory(".mvn")).
		WithFile("mvnw", localDirectory.File("mvnw"), dagger.ContainerWithFileOpts{
			Permissions: 0755,
		}).
		WithFile("pom.xml", localDirectory.File("pom.xml")).
		WithExec([]string{"./mvnw", "dependency:go-offline"}).
		WithDirectory("./src", localDirectory.Directory("src")).
		WithExec([]string{"./mvnw", "clean", "install"})

	// Stage 2: final container which has some files copied from the build container
	finalContainer := dag.Container().
		From("eclipse-temurin:21.0.2_13-jre-jammy").
		WithWorkdir("/opt/app").
		WithExposedPort(8080).
		WithDirectory("/opt/app", buildStage.Directory("/opt/app/target"), dagger.ContainerWithDirectoryOpts{
			Include: []string{"*.jar"},
		}).
		WithEntrypoint([]string{"java", "-jar", "/opt/app/*.jar"})

	return finalContainer
}

// Build a Rust application using cross-compilation with Zig
//
// # Dockerfile used:
// FROM --platform=$BUILDPLATFORM ubuntu AS build
// ENV HOME="/root"
// WORKDIR $HOME
//
// RUN apt update && apt install -y build-essential curl python3-venv
//
// # Setup zig as cross compiling linker
// RUN python3 -m venv $HOME/.venv
// RUN .venv/bin/pip install cargo-zigbuild
// ENV PATH="$HOME/.venv/bin:$PATH"
//
// # Install rust
// ARG TARGETPLATFORM
//
//	RUN case "$TARGETPLATFORM" in \
//	    "linux/arm64") echo "aarch64-unknown-linux-musl" > rust_target.txt ;; \
//	    "linux/amd64") echo "x86_64-unknown-linux-musl" > rust_target.txt ;; \
//	    *) exit 1 ;; \
//	    esac
//
// # Update rustup whenever we bump the rust version
// COPY rust-toolchain.toml rust-toolchain.toml
// RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --target $(cat rust_target.txt) --profile minimal --default-toolchain none
// ENV PATH="$HOME/.cargo/bin:$PATH"
// # Installs the correct toolchain version from rust-toolchain.toml and then the musl target
// RUN rustup target add $(cat rust_target.txt)
//
// # Build
// COPY crates crates
// COPY Cargo.toml Cargo.toml
// COPY Cargo.lock Cargo.lock
// RUN cargo zigbuild --bin ruff --target $(cat rust_target.txt) --release
// RUN cp target/$(cat rust_target.txt)/release/ruff /ruff
//
// FROM scratch
// COPY --from=build /ruff /ruff
// WORKDIR /io
// ENTRYPOINT ["/ruff"]
func (m *MyModule) Example3(
	ctx context.Context,
	// Local directory to use for the build
	localDirectory *dagger.Directory,
	// Target platform for the build, omit to use the default platform
	//
	// +optional
	platform string,
) *dagger.Container {
	buildStage := dag.Container().
		From("ubuntu").
		WithEnvVariable("HOME", "/root").
		WithWorkdir("/root").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "build-essential", "curl", "python3-venv"}).
		WithExec([]string{"python3", "-m", "venv", "/root/.venv"}).
		WithExec([]string{"/root/.venv/bin/pip", "install", "cargo-zigbuild"}).
		// When the environment variable value contains another environment variable, we must set the Expand flag
		WithEnvVariable("PATH", "/root/.venv/bin:$PATH", dagger.ContainerWithEnvVariableOpts{
			Expand: true,
		})

	if platform == "" {
		defaultPlatform, err := dag.DefaultPlatform(ctx)
		if err != nil {
			panic("cannot get the default platform")
		}
		platform = string(defaultPlatform)
	}

	rustTarget := ""
	switch platform {
	case "linux/arm64":
		rustTarget = "aarch64-unknown-linux-musl"
	case "linux/amd64":
		rustTarget = "x86_64-unknown-linux-musl"
	default:
		panic(fmt.Sprintf("unsupported platform: %s", platform))
	}

	buildStage = buildStage.
		WithFile("rust-toolchain.toml", localDirectory.File("rust-toolchain.toml")).
		WithExec([]string{"sh", "-c", "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --target " + rustTarget + " --profile minimal --default-toolchain none"}).
		// When the environment variable value contains another environment variable, we must set the Expand flag
		WithEnvVariable("PATH", "/root/.cargo/bin:$PATH", dagger.ContainerWithEnvVariableOpts{
			Expand: true,
		}).
		WithExec([]string{"rustup", "target", "add", rustTarget}).
		WithDirectory("crates", localDirectory.Directory("crates")).
		WithFile("Cargo.toml", localDirectory.File("Cargo.toml")).
		WithFile("Cargo.lock", localDirectory.File("Cargo.lock")).
		WithExec([]string{"cargo", "zigbuild", "--bin", "ruff", "--target", rustTarget, "--release"}).
		WithExec([]string{"cp", "target/" + rustTarget + "/release/ruff", "/ruff"})

	// Dagger has no equivalent to the "scratch" image, just omit the From to have an empty container.
	return dag.Container().
		WithFile("/ruff", buildStage.File("/ruff")).
		WithWorkdir("/io").
		WithEntrypoint([]string{"/ruff"})
}
