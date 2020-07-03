FROM alpine 

ENV ANSIBLE_HOST_KEY_CHECKING=False

RUN apk add ansible gcc python3 python3-dev libc-dev libffi-dev openssl-dev git py3-pip openssh-client
RUN pip3 install --upgrade paramiko
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
