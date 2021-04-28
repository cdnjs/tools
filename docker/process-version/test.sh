ls -lh /input

tar -xvf /input/new-version.tgz -C /tmp

cp -v /tmp/package/* /output
ls -lh /output
