#!/bin/sh

ssh-agent bash
echo "$BOT_DEPLOY_KEY" | base64 -d > id_rsa
ssh-add id_rsa

git clone https://github.com/cdnjs/bot-ansible.git

ansible-playbook \
  -i bot-ansible/prod \
  bot-ansible/autoupdater.yml \
  --tags tools \
  --user=deploy
