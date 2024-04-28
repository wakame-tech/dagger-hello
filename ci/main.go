package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// use a node:16-slim container
	// mount the source code directory on the host
	// at /src in the container
	source := client.Container().
		From("node:16-slim").
		WithDirectory("/src", client.Host().Directory(".", dagger.HostDirectoryOpts{
			Exclude: []string{"node_modules/", "ci/", "build/"},
		}))

		// set the working directory in the container
		// install application dependencies
	runner := source.WithWorkdir("/src").
		WithExec([]string{"npm", "install"})

		// run application tests
	test := runner.WithExec([]string{"npm", "test", "--", "--watchAll=false"})

	// build application
	// write the build output to the host
	_, err = test.WithExec([]string{"npm", "run", "build"}).
		Directory("./build").
		Export(ctx, "./build")

	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("docker.io/wakametech/hello-dagger:%0.f", math.Floor(rand.Float64()*10000000))
	// url := fmt.Sprintf("ttl.sh/hello-dagger-%.0f", math.Floor(rand.Float64()*10000000))

	token := client.SetSecret("docker-hub-token", os.Getenv("DOCKER_HUB_TOKEN"))
	fmt.Printf("Using token: %s\n", os.Getenv("DOCKER_HUB_TOKEN"))

	ref, err := client.Container().
		From("nginx:1.23-alpine").
		WithRegistryAuth("docker.io", "wakametech", token).
		WithDirectory("/usr/share/nginx/html", client.Host().Directory("./build")).
		Publish(ctx, url)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Published image to: %s\n", ref)
}
