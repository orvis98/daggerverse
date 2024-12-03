# talos

# Run Locally

Ensure [privileged execution](https://docs.dagger.io/configuration/custom-runner/#privileged-execution) is enabled in the engine.

```bash
# see configuration options
dagger -m github.com/orvis98/daggerverse/talos call bootstrap -h
# bootstrap the cluster and export the talosconfig
dagger -m github.com/orvis98/daggerverse/talos call bootstrap file --path talosconfig export --path ~/.talos/config
# export kubeconfig
dagger -m github.com/orvis98/daggerverse/talos call kubeconfig export --path ~/.kube/config
# expose the controlplane on localhost
dagger -m github.com/orvis98/daggerverse/talos call proxy up --ports 6443:6443,50000:50000
```
