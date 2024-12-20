.PHONY: test manager server images deploy rollout upgrade

test:
	go test ./...

manager:
	CGO_ENABLED=0 go build -o ./manager/bin/ ./manager/
	docker build manager -t sampling-manager

server:
	CGO_ENABLED=0 go build -o ./server/bin/ ./server/
	docker build server -t simple-server

images: manager server
	minikube image load sampling-manager:latest --overwrite=true
	minikube image load simple-server:latest --overwrite=true

deploy:
	minikube kubectl -- apply -f deploy

destroy:
	minikube kubectl -- delete deployment.apps/manager -n test
	minikube kubectl -- delete deployment.apps/server -n test

rollout:
	minikube kubectl -- rollout restart deployment.apps/server -n test
	minikube kubectl -- rollout restart deployment.apps/manager -n test

# Note!
#
# Must destroy before pushing new images if pods are running.
upgrade: destroy images deploy
	echo "upgrade done"
