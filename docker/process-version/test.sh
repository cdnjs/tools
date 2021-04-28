ls -lh /input

tar -xvf /input/new-version.tgz -C /tmp

cp -v /tmp/packages/* /output
ls -lh /output
