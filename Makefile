LOCAL_REGISTRY_HOSTNAME=localhost
LOCAL_REGISTRY_CONTAINERNAME=k3d-local-registry
LOCAL_REGISTRY_PORT=$(shell docker port ${LOCAL_REGISTRY_CONTAINERNAME} 5000 | cut -f2 -d":")

ifeq ($(TEST_ARGS),)
TEST_ARGS := -t "pod" -r "testpod" -n "default"
endif

build-snapshot:
	goreleaser build --snapshot --rm-dist

test-create-dockerimage: build-snapshot
	docker build -t sensu-kubetest .
 
test-push-dockerimage-local: test-create-dockerimage
	docker tag sensu-kubetest ${LOCAL_REGISTRY_HOSTNAME}:${LOCAL_REGISTRY_PORT}/sensu-kubetest
	docker push ${LOCAL_REGISTRY_HOSTNAME}:${LOCAL_REGISTRY_PORT}/sensu-kubetest

test-compare: test-push-dockerimage-local
	@[ -n "$(findstring k3d,$(shell kubectl config current-context))" ] || (echo "ERROR: your context is not targeting a k3d (local) cluster. Please switch."; false)
	kubectl run testpod --rm -it --image ${LOCAL_REGISTRY_CONTAINERNAME}:${LOCAL_REGISTRY_PORT}/sensu-kubetest --restart Never -- sensu-check-kubernetes-compare ${TEST_ARGS}