// Allow fetching Docker images.

package modules

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/skx/marionette/config"
)

// DockerModule stores our state
type DockerModule struct {

	// cfg contains our configuration object.
	cfg *config.Config

	// Cached list of image-tags we've got available on the local host.
	Tags []string
}

// Check is part of the module-api, and checks arguments.
func (dm *DockerModule) Check(args map[string]interface{}) error {

	// Ensure we have an image to pull.
	_, ok := args["image"]
	if !ok {
		return fmt.Errorf("missing 'image' parameter")
	}

	return nil
}

// isInstalled tests if the given image is installed
func (dm *DockerModule) isInstalled(img string) (bool, error) {

	//
	// Cached tag-list already?
	//
	if len(dm.Tags) > 0 {

		//
		// Does the image appear in any of our cached tags?
		//
		for _, x := range dm.Tags {
			if x == img {
				return true, nil
			}
		}

		//
		// Not found.
		//
		return false, nil
	}

	// Create a new client.
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}

	// Get all images
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false, err

	}

	//
	// If we reached here we have no cached tags.
	//
	// Save the tags in the cache before we look
	// for a match.
	//
	found := false
	for _, image := range images {
		for _, x := range image.RepoTags {

			// Update the cache
			dm.Tags = append(dm.Tags, x)
			if x == img {
				found = true
			}
		}
	}

	// Return the result
	return found, nil
}

// installImage pulls the given image from the remote repository.
//
// NOTE: No authentication, or private registries are supported.
func (dm *DockerModule) installImage(img string) error {

	// Create client.
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// Pull the image.
	out, err := cli.ImagePull(ctx, img, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	// Copy output to console.
	//
	// TODO: Clean this up
	defer out.Close()

	if dm.cfg.Verbose {
		_, err := io.Copy(os.Stdout, out)
		if err != nil {
			return err
		}
	}

	// No error.
	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (dm *DockerModule) Execute(args map[string]interface{}) (bool, error) {

	// We might have multiple images to fetch
	var images []string

	// Single image?
	p := StringParam(args, "image")
	if p != "" {
		images = append(images, p)
	}

	// Force the pull?
	force := StringParam(args, "force")

	// Array of packages?
	a := ArrayParam(args, "image")
	if len(a) > 0 {
		images = append(images, a...)
	}

	// installed something?
	installed := false

	// For each image the user wanted to fetch
	for _, img := range images {

		// Check if it is installed
		present, err := dm.isInstalled(img)
		if err != nil {
			return false, err
		}

		// Not installed; fetch.
		if !present || (force == "yes") {

			if dm.cfg.Verbose {
				fmt.Printf("\tPulling docker image %s\n", img)
			}

			err := dm.installImage(img)
			if err != nil {
				return false, err
			}
			installed = true
		}
	}

	// Return whether we installed something.
	return installed, nil
}

// init is used to dynamically register our module.
func init() {
	Register("docker", func(cfg *config.Config) ModuleAPI {
		return &DockerModule{cfg: cfg}
	})
}
