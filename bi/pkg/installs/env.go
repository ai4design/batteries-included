package installs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"bi/pkg/specs"

	"bi/pkg/cluster"
	"bi/pkg/cluster/kind"

	"github.com/adrg/xdg"
)

// This wraps all the paths and locations for caching data of an install
// the spec is used to tell us what should be running.
type InstallEnv struct {
	// The Slug of the customer install
	Slug            string
	clusterProvider cluster.Provider
	Spec            *specs.InstallSpec
	source          string
}

// Init Function generate all needed
func (env *InstallEnv) init(ctx context.Context) error {
	slog.Debug("Initializing install", slog.String("slug", env.Slug))
	// Create the install directory in the xdg state home
	if err := os.MkdirAll(env.InstallStateHome(), 0o700); err != nil {
		return fmt.Errorf("error creating install directory: %w", err)
	}

	// Try writing the spec and summary to the install directory
	// Don't overwrite
	if err := env.WriteSpec(false); err != nil {
		return fmt.Errorf("error checking spec is writeable: %w", err)
	}

	if err := env.WriteSummary(false); err != nil {
		return fmt.Errorf("error checking summary is writeable: %w", err)
	}

	provider := env.Spec.KubeCluster.Provider

	switch provider {
	case "kind":
		needsLocalGateway, err := env.Spec.NeedsLocalGateway()
		if err != nil {
			return fmt.Errorf("error checking if local gateway is needed: %w", err)
		}
		dockerDesktop, _ := kind.IsDockerDesktop(ctx)
		podman, _ := kind.IsPodmanAvailable()

		gatewayEnabled := needsLocalGateway && (dockerDesktop || podman)
		env.clusterProvider = kind.NewClusterProvider(slog.Default(), env.Slug, gatewayEnabled)
	case "aws":
		env.clusterProvider = cluster.NewPulumiProvider(env.Spec)
	case "provided":
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	if err := env.clusterProvider.Init(ctx); err != nil {
		return fmt.Errorf("error initializing cluster provider: %w", err)
	}

	return nil
}

func NewEnv(ctx context.Context, slugOrURL string) (*InstallEnv, error) {
	installEnv, err := readInstallEnv(slugOrURL)
	if err != nil {
		return nil, fmt.Errorf("error reading install env: %w", err)
	}

	return installEnv, nil
}

func readInstallEnv(slugOrURL string) (*InstallEnv, error) {
	type potentialPath struct{ source, path string }
	for _, p := range []potentialPath{
		{source: "file", path: filepath.Join(xdg.StateHome, "bi", "installs", slugOrURL, "spec.json")},
		{source: "url", path: slugOrURL},
	} {
		l := slog.With(slog.String("path", p.path), slog.String("source", p.source))

		spec, err := specs.GetSpecFromURL(p.path)
		if err != nil {
			l.Debug("Didn't find install", slog.Any("error", err))
			continue
		}
		l.Debug("Found install")
		return &InstallEnv{Slug: spec.Slug, Spec: spec, source: p.source}, nil
	}

	return nil, errors.New("no spec found")
}

// NeedsKubeCleanup returns true if we should remove all resources in an install
func (env *InstallEnv) NeedsKubeCleanup() bool {
	// Returns true if the cluster provider is in [provided, aws]
	provider := env.Spec.KubeCluster.Provider
	if provider != "provided" && provider != "aws" {
		return false
	}

	// Do we have a kube config? Eg. we finished bootstrapping.
	_, err := os.Stat(env.KubeConfigPath())
	return err == nil
}

func (env *InstallEnv) Init(ctx context.Context, remove bool) error {
	// since NeedsKubeCleanup is predicated on their being a kube config, allow skipping removal
	if env.source == "url" && remove {
		// We got from the url so we should remove everything
		_ = env.Remove()
	}

	if err := env.init(ctx); err != nil {
		return fmt.Errorf("error initializing install: %w", err)
	}
	return nil
}
