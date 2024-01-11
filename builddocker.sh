#!/bin/bash
if [ -z ${1} ]; then
  echo "Need version param";
  exit 1;
else
  ver=$1;
fi
builddate="$(date '+%Y-%m-%d_%H:%M:%S')"
echo $builddate

docker build -t wabackend:$ver --build-arg VER=$ver --build-arg BUILDDATE=$builddate .
docker tag wabackend:$ver mygit.imitra.com:5020/wabackend:$ver
docker tag wabackend:$ver mygit.imitra.com:5020/wabackend:latest
docker push mygit.imitra.com:5020/wabackend:$ver
docker push mygit.imitra.com:5020/wabackend:latest
docker image prune --filter label=stage=builder --force
