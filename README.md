# k8s-remote-config
The use-case is remote configuration of a k8s pod that don't (or can't) watch their own configs on disk, which is common for networked storage that don't support file system watchers.

So we have a dedicated pod with a API (like HTTP or SQS) that allows for arbitrary updates to a configmap mounted on some target pod. Post config update, we annotate the target pod deployment to trigger a rollout and upon restarting the target pod will read the updated config.

This is a POC.

# Usage
Requirements: minikube, docker, go, stern.

````sh
# start the cluster
$ minikube start
# build and push images to the cluster
$ make images
# deploy all the stuff to -n test
$ make deploy
# update the config
$ export IP=$(minikube ip)
$ export PORT=$(minikube kubectl -- get service manager -n test -o jsonpath='{.spec.ports[0].nodePort}')
$ curl -i "http://$IP:$PORT/update?k=day&v=friday"
# verify change
$ stern . -n test
````

# TODOs
- [x] images
- [x] manifests
- [x] k8s API permissions
- [ ] consider https://github.com/stakater/Reloader
- [ ] API auth*
