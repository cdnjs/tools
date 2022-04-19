export DOCKER_BUILDKIT=1

package=$1
version=$2

set -e

echo "processing $package $version"

mkdir -p /tmp/input /tmp/output 
rm -rf /tmp/output/* /tmp/input/*

echo "loading new version files"
curl --fail https://storage.googleapis.com/cdnjs-incoming-prod/$package-$version.tgz > /tmp/input/new-version.tgz
echo "loading package configuration"
curl --fail https://raw.githubusercontent.com/cdnjs/packages/master/packages/${package::1}/$package.json > /tmp/input/config.json

cat /tmp/input/config.json

echo "----------------- input files -----------------"
ls -lh /tmp/input

docker build -f docker/process-version/Dockerfile -t sandbox .
docker run -it -v /tmp/input:/input -v /tmp/output:/output sandbox

echo "----------------- output files -----------------"
ls -lh /tmp/output

echo "checking SRIs"
for f in /tmp/output/*.sri
do
  (echo $f | sed s/.sri/.gz/ | xargs cat | gzip -d > /tmp/file)

  expected=$(shasum -b -a 512 /tmp/file | awk '{ print $1 }' | xxd -r -p | base64 -w0)
  actual=$(cat $f)

  if [ "sha512-$expected" != "$actual" ]; then
    echo "SRI mismatch for $f"
    exit 1
  fi
done

echo "OK"
