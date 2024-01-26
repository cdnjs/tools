package=$1
version=$2

set -e

echo "processing $package $version"

export DOCKER_BUILDKIT=1
rm -rf /tmp/output /tmp/input
mkdir -p /tmp/input /tmp/output 

echo "loading new version files"
curl --fail https://storage.googleapis.com/cdnjs-incoming-prod/$package-$version.tgz > /tmp/input/new-version.tgz
echo "loading package configuration"
curl --fail https://raw.githubusercontent.com/cdnjs/packages/master/packages/${package::1}/$package.json > /tmp/input/config.json

cat /tmp/input/config.json | jq .

echo "----------------- input files -----------------"
ls -lh /tmp/input

tar -tvf /tmp/input/new-version.tgz

docker build -f docker/process-version/Dockerfile -t sandbox .
docker run -it -v /tmp/input:/input -v /tmp/output:/output sandbox

echo "----------------- output files -----------------"
ls -lh /tmp/output
