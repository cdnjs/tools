#!/bin/sh
set -e

eval $(ssh-agent -s)
echo "$BOT_DEPLOY_KEY" | base64 -d | ssh-add -

echo "$BOT_HOST" >> /etc/hosts

git clone https://github.com/cdnjs/bot-ansible.git

ansible-playbook \
  -i bot-ansible/prod \
  bot-ansible/autoupdater.yml \
  --tags tools \
  --user=deploy
