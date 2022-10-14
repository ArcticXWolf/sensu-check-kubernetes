FROM ubuntu:latest

COPY ./dist/sensu-check-kubernetes-count_linux_amd64_v1/bin/sensu-check-kubernetes-count /bin/sensu-check-kubernetes-count

COPY ./dist/sensu-check-kubernetes-compare_linux_amd64_v1/bin/sensu-check-kubernetes-compare /bin/sensu-check-kubernetes-compare

