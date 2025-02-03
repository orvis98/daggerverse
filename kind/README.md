# kind

A Dagger module for managing kind clusters.

## Example

```bash
# Create the cluster
export DOCKER_SOCKET="/var/run/docker.sock"
export CLUSTER_NAME="example-cluster"
export KIND_CONFIG_FILE="./kindconfig.yaml"
export KIND_REPOSITORY="kindest/node"
export KUBERNETES_VERSION="v1.32.1"
dagger -m github.com/orvis98/daggerverse/kind call --docker-socket $DOCKER_SOCKET create-cluster \
  --name $CLUSTER_NAME \
  --config $KIND_CONFIG_FILE \
  --image "${KIND_REPOSITORY}:${KUBERNETES_VERSION}"
# Export the kubeconfig
export KUBECONFIG="~/.kube/config"
dagger -m github.com/orvis98/daggerverse/kind call --docker-socket /var/run/docker.sock kubeconfig \
  --name $CLUSTER_NAME export --path $KUBECONFIG
# Export logs to a host directory
dagger -m github.com/orvis98/daggerverse/kind call --docker-socket /var/run/docker.sock logs \
  --name $CLUSTER_NAME export --path ./logs
# Delete the cluster
dagger -m github.com/orvis98/daggerverse/kind call --docker-socket /var/run/docker.sock delete-cluster \
  --name $CLUSTER_NAME
```
