#!/bin/sh

echo "$BOT_DEPLOY_KEY" | base64 -d > ~/.ssh/id_rsa

git clone https://github.com/cdnjs/bot-ansible.git .

ansible-playbook -i prod autoupdater.yml --tags tools --user=deploy
