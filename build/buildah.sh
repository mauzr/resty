#!/usr/bin/bash

set -eu

buildah manifest create $IMAGE:$GITTAG || true
for arch in amd64 arm64
do
  case $arch in
  amd64)
    dockerarch="amd64";;
  arm64)
    dockerarch="arm64v8";;
  *)
    echo "AAAAAAAAAAAAAAAAAAH"; exit 1;;
  esac

  container=$(buildah from scratch)
  buildah copy $container ./build/passwd /etc/passwd > /dev/null
  buildah config -l org.label-schema.build-date=$BUILD_DATE -l org.label-schema.name=mauzr \
    -l org.label-schema.vcs-url="https://go.eqrx.net/mauzr" --user nobody --entrypoint /usr/local/bin/mauzr \
    --port 443/tcp --healthcheck "/usr/local/bin/mauzr healthcheck" --healthcheck-interval 1m --healthcheck-retries 1 \
    --healthcheck-start-period 30s --healthcheck-timeout 5s $container
  buildah copy $container ./dist/$arch/mauzr /usr/local/bin/mauzr > /dev/null
  buildah commit --squash $container ${IMAGE}:${dockerarch}-${GITTAG} > /dev/null
  buildah manifest add --arch $dockerarch $IMAGE:$GITTAG ${IMAGE}:${dockerarch}-${GITTAG} > /dev/null
done
