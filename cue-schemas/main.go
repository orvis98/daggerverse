// A module for vendoring and publishing CUE schemas to registry

package main

import (
	"context"
	"dagger/cue-schemas/internal/dagger"
	_ "embed"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v67/github"
	yamlv3 "gopkg.in/yaml.v3"
)

type CueSchemas struct {
	// returns the timoni version
	TimoniVersion string
	// returns the cue version
	CueVersion string
}

type GithubSource struct {
	Tag    string   `yaml:"tag"`
	Ref    string   `yaml:"ref"`
	Owner  string   `yaml:"owner"`
	Repo   string   `yaml:"repo"`
	Files  []string `yaml:"files"`
	Dirs   []string `yaml:"dirs"`
	Assets []string `yaml:"assets"`
}

type KubernetesSource struct {
	Version string `yaml:"version"`
}

type Sources struct {
	Github     []GithubSource     `yaml:"github"`
	Kubernetes []KubernetesSource `yaml:"kubernetes"`
}

//go:embed schema.cue
var schemaFile string

func New(
	// +optional
	// +default="v0.23.0"
	// the desired timoni version
	timoniVersion string,
	// +optional
	// +default="v0.11.0"
	// the desired CUE version
	cueVersion string,
) *CueSchemas {
	return &CueSchemas{
		TimoniVersion: timoniVersion,
		CueVersion:    cueVersion,
	}
}

// returns a container with the timoni and cue binaries
func (m *CueSchemas) Container() *dagger.Container {
	return dag.Container().
		From("golang").
		WithExec([]string{"go", "install", fmt.Sprintf("github.com/stefanprodan/timoni/cmd/timoni@%s", m.TimoniVersion)}).
		WithExec([]string{"go", "install", fmt.Sprintf("cuelang.org/go/cmd/cue@%s", m.CueVersion)})
}

// vendor Kubernetes API CUE schemas
func (m *CueSchemas) VendorKubernetes(version string) *dagger.Directory {
	semver := semver.MustParse(version)
	dir := m.Container().
		WithExec([]string{"cue", "mod", "init"}).
		WithExec([]string{"timoni", "mod", "vendor", "k8s", "-v", fmt.Sprintf("%d.%d", semver.Major(), semver.Minor())}).
		WithWorkdir("cue.mod/gen/k8s.io").
		WithExec([]string{"cue", "mod", "init", fmt.Sprintf("k8s.io@v%d", semver.Major()), "--source=self"}).
		Directory(".")
	return dag.Container().
		WithDirectory(fmt.Sprintf("k8s.io-%s", version), dir).
		Directory(".")
}

// vendor Timoni CUE schemas for the current version
func (m *CueSchemas) VendorTimoni() *dagger.Directory {
	semver := semver.MustParse(m.TimoniVersion)
	dir := m.Container().
		WithExec([]string{"timoni", "mod", "init", "derp"}).
		WithWorkdir("derp/cue.mod/pkg/timoni.sh").
		WithExec([]string{"cue", "mod", "init", fmt.Sprintf("timoni.sh@v%d", semver.Major()), "--source=self"}).
		Directory(".")
	return dag.Container().
		WithDirectory(fmt.Sprintf("timoni.sh-%s", m.TimoniVersion), dir).
		Directory(".")
}

// vendor Kubernetes CRD CUE schemas from GitHub
func (m *CueSchemas) VendorGithub(
	ctx context.Context,
	// the desired tag
	tag string,
	// the github ref
	ref string,
	// the github owner
	owner string,
	// the github repo
	repo string,
	// +optional
	// the repo files to vendor
	file []string,
	// +optional
	// the repo directories to vendor
	dir []string,
	// +optional
	// the repo release assets to vendor
	asset []string,
) (*dagger.Directory, error) {
	semver := semver.MustParse(tag)
	client := github.NewClient(nil)
	if ref == "" {
		ref = tag
	}
	var files []string
	for _, f := range file {
		files = append(files, fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/tags/%s/%s", owner, repo, ref, f))
	}
	for _, d := range dir {
		_, entries, _, err := client.Repositories.GetContents(ctx, owner, repo, d, &github.RepositoryContentGetOptions{Ref: ref})
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if strings.HasSuffix(e.GetName(), ".yml") || strings.HasSuffix(e.GetName(), ".yaml") {
				files = append(files, e.GetDownloadURL())
			}
		}
	}
	for _, a := range asset {
		files = append(files, fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, ref, a))
	}
	ctr := m.Container().
		WithExec([]string{"cue", "mod", "init"})
	for _, f := range files {
		ctr = ctr.WithExec([]string{"timoni", "mod", "vendor", "crds", "-f", f})
	}
	ctr = ctr.WithWorkdir("cue.mod/gen")
	mods, _ := ctr.Directory(".").Entries(ctx)
	for _, mod := range mods {
		ctr = ctr.WithWorkdir(mod).
			WithExec([]string{"cue", "mod", "init", fmt.Sprintf("%s@v%d", mod, semver.Major()), "--source=self"}).
			WithWorkdir("..")
		ctr = ctr.WithDirectory(fmt.Sprintf("%s-%s", mod, tag), ctr.Directory(mod)).
			WithoutDirectory(mod)
	}
	return ctr.Directory("."), nil
}

// validate a sources.yaml file
func (m *CueSchemas) Validate(ctx context.Context, file *dagger.File) error {
	cctx := cuecontext.New()
	schema := cctx.CompileString(schemaFile).LookupPath(cue.ParsePath("#Schema"))
	contents, _ := file.Contents(ctx)
	return yaml.Validate([]byte(contents), schema)
}

// vendor CUE schemas from a sources.yaml file
func (m *CueSchemas) Vendor(ctx context.Context, file *dagger.File) (*dagger.Directory, error) {
	if err := m.Validate(ctx, file); err != nil {
		return nil, err
	}
	contents, _ := file.Contents(ctx)
	var sources Sources
	if err := yamlv3.Unmarshal([]byte(contents), &sources); err != nil {
		return nil, err
	}
	ctr := dag.Container()
	for _, s := range sources.Github {
		mods, err := m.VendorGithub(ctx, s.Tag, s.Ref, s.Owner, s.Repo, s.Files, s.Dirs, s.Assets)
		if err != nil {
			return nil, err
		}
		entries, _ := mods.Entries(ctx)
		for _, e := range entries {
			ctr = ctr.WithDirectory(e, mods.Directory(e))
		}
	}
	for _, s := range sources.Kubernetes {
		mods := m.VendorKubernetes(s.Version)
		ctr = ctr.WithDirectory("k8s.io-"+s.Version, mods.Directory("k8s.io-"+s.Version))
	}
	ctr = ctr.WithDirectory("timoni.sh-"+m.TimoniVersion, m.VendorTimoni().Directory("timoni.sh-"+m.TimoniVersion))
	return ctr.Directory("."), nil
}

// publish CUE schemas from a sources.yaml file
func (m *CueSchemas) Publish(
	ctx context.Context,
	file *dagger.File,
	// +optional
	// the registry URL
	registry string,
	// +optional
	// +default="derp"
	// the registry username
	username string,
	// +optional
	// the registry password
	password *dagger.Secret,
	// +optional
	// the registry service
	service *dagger.Service,
) (string, error) {
	dir, err := m.Vendor(ctx, file)
	if err != nil {
		return "", err
	}
	mods, _ := dir.Entries(ctx)
	ctr := m.Container().WithEnvVariable("CUE_REGISTRY", registry)
	if registry == "" && service == nil {
		return "", fmt.Errorf("one of registry or service is required")
	} else if registry == "" {
		endpoint, _ := service.Endpoint(ctx)
		ctr = ctr.WithServiceBinding("registry", service).
			WithEnvVariable("CUE_REGISTRY", fmt.Sprintf("%s+insecure", endpoint))
	}
	if password != nil {
		docker := dag.Container().
			From("docker").
			WithSecretVariable("REGISTRY_PASSWORD", password).
			WithExec([]string{"sh", "-c", fmt.Sprintf("docker login -u %s -p $REGISTRY_PASSWORD %s", username, registry)}).
			File("/root/.docker/config.json")
		ctr = ctr.WithFile("/root/.docker/config.json", docker)
	}
	var result string
	for _, m := range mods {
		parts := strings.Split(m, "-")
		version := parts[len(parts)-1]
		stdout, err := ctr.WithDirectory(m, dir.Directory(m)).
			WithWorkdir(m).
			WithExec([]string{"cue", "mod", "publish", version}).
			Stdout(ctx)
		if err != nil {
			return result, err
		}
		result += stdout
	}
	return result, nil
}

// export Kubernetes CRDs from GitHub
func (m *CueSchemas) ExportGithub(
	ctx context.Context,
	// +optional
	// the desired ref
	tag string,
	// the desired ref
	ref string,
	// the github owner
	owner string,
	// the github repo
	repo string,
	// +optional
	// the repo files to vendor
	file []string,
	// +optional
	// the repo directories to vendor
	dir []string,
	// +optional
	// the repo release assets to vendor
	asset []string,
) (*dagger.File, error) {
	client := github.NewClient(nil)
	if ref == "" {
		ref = tag
	}
	var files []string
	for _, f := range file {
		files = append(files, fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/tags/%s/%s", owner, repo, ref, f))
	}
	for _, d := range dir {
		_, entries, _, err := client.Repositories.GetContents(ctx, owner, repo, d, &github.RepositoryContentGetOptions{Ref: ref})
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if strings.HasSuffix(e.GetName(), ".yml") || strings.HasSuffix(e.GetName(), ".yaml") {
				files = append(files, e.GetDownloadURL())
			}
		}
	}
	for _, a := range asset {
		files = append(files, fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, ref, a))
	}
	ctr := m.Container().
		WithWorkdir("/tmp/gen")
	for _, f := range files {
		ctr = ctr.WithExec([]string{"wget", f})
	}
	ctr = ctr.WithExec([]string{"cue", "import", "-fl", "strings.ToLower(kind)", "-l", "strings.ToLower(metadata.name)", "-o", "all.cue"}).
		WithExec([]string{"cue", "export", "-e", "customresourcedefinition", "-o", "crds.cue", "all.cue"})
	return ctr.File("crds.cue"), nil
}

// export Kubernetes CRDs from a sources.yaml file
func (m *CueSchemas) Export(ctx context.Context, file *dagger.File) (*dagger.Directory, error) {
	if err := m.Validate(ctx, file); err != nil {
		return nil, err
	}
	contents, _ := file.Contents(ctx)
	var sources Sources
	if err := yamlv3.Unmarshal([]byte(contents), &sources); err != nil {
		return nil, err
	}
	ctr := dag.Container()
	for _, s := range sources.Github {
		crds, err := m.ExportGithub(ctx, s.Tag, s.Ref, s.Owner, s.Repo, s.Files, s.Dirs, s.Assets)
		if err != nil {
			return nil, err
		}
		ctr = ctr.WithFile(s.Owner+"-"+s.Repo+".cue", crds)
	}
	return ctr.Directory("."), nil
}
