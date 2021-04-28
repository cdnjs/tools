export DOCKER_BUILDKIT=1

rm -rf /tmp/output/*

curl https://storage.googleapis.com/cdnjs-incoming-staging/fontawesome-free-5.15.3.tgz > /tmp/input/new-version.tgz
curl https://raw.githubusercontent.com/cdnjs/packages/master/packages/f/font-awesome.json > /tmp/input/config.json

ls -lh /tmp/input

docker build -f docker/process-version/Dockerfile -t sandbox .
docker run -it -v /tmp/input:/input -v /tmp/output:/output sandbox

ls -lh /tmp/output
