export DOCKER_BUILDKIT=1
mkdir -p /tmp/input /tmp/output 

rm -rf /tmp/output/*

curl https://storage.googleapis.com/cdnjs-incoming-prod/design-system-2.15.4.tgz > /tmp/input/new-version.tgz
curl https://raw.githubusercontent.com/cdnjs/packages/master/packages/d/design-system.json > /tmp/input/config.json

ls -lh /tmp/input

docker build -f docker/process-version/Dockerfile -t sandbox .
docker run -it -v /tmp/input:/input -v /tmp/output:/output sandbox

ls -lh /tmp/output
