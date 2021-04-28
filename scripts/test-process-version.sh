export DOCKER_BUILDKIT=1

rm -rf /tmp/output/*

curl https://raw.githubusercontent.com/cdnjs/packages/master/packages/j/jquery.json > /tmp/input/config.json

docker build -f docker/process-version/Dockerfile -t sandbox .
docker run -it -v /tmp/input:/input -v /tmp/output:/output sandbox

ls -lh /tmp/output
