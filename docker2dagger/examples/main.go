// A generated module for Examples functions
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
	"dagger/examples/internal/dagger"
)

type Examples struct{}

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
func (m *Examples) Example1(localDirectory *dagger.Directory) *dagger.Container {
	// Dagger cannot access the local directory, so the directory on the host is passed as an argument to the function
	// Here is an example of calling the function with the current working directory passed as the argument:
	// dagger call example1 --local-directory .
	return dag.Container().
		From("python:3.12").
		WithWorkdir("/usr/local/app").
		WithFile("./requirements.txt", localDirectory.File("requirements.txt")).
		WithExec([]string{"pip", "install", "--no-cache-dir", "-r", "requirements.txt"}).
		WithDirectory("./src", localDirectory.Directory("src")).
		WithExposedPort(5000).
		WithExec([]string{"useradd", "app"}).
		WithUser("app").
		WithDefaultArgs([]string{"uvicorn", "app.main:app", "--host", "8080"})
}

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
func (m *Examples) Example2(localDirectory *dagger.Directory) *dagger.Container {
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
