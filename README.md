# apiServer-kubebuilder-operator
This is a simple operator for apiserver written using kubebuilder. 

## Necessary Make Command
```shell
make manifests # This will create necessary crd,rbac yaml
make deploy # Deploy Controller to the cluster
make undeploy # Undeploy Controller from the cluster. Run it when testing is done
```

## Running controller from the outside of the cluster
```shell
make run
```

## Create Sample Custom Resource
```shell
kubectl apply -f config/samples/appscode_v1_apiserver.yaml # Create
kubectl delete -f config/samples/appscode_v1_apiserver.yaml # Delete
```