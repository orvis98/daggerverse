# cue-schemas

See [sources.example.yaml](./sources.example.yaml) for an example configuration.

## Publish to GHCR

```bash
# publish 
dagger -m github.com/orvis98/daggerverse/cue-schemas call publish --file ./sources.yaml --registry ghcr.io/$OWNER/$REPO --password "env:GITHUB_TOKEN"
```

## Export CRDs

```bash
# export
dagger -m github.com/orvis98/daggerverse/cue-schemas call export --file ./sources.yaml export --path ./crds
```
