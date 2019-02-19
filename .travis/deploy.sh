#!/bin/bash

eval "$(ssh-agent -s)"
chmod 600 .travis/id_rsa
ssh-add .travis/id_rsa

ssh-keyscan -t $TRAVIS_SSH_KEY_TYPES -H $DEPLOY_URL 2>&1 | tee -a $HOME/.ssh/known_hosts

git config --global push.default matching
git remote add deploy ssh://$DEPLOY_USER@$DEPLOY_URL:$DEPLOY_DIR
git push deploy master

ssh $DEPLOY_USER@$DEPLOY_URL <<EOF
    cd $DEPLOY_DIR
    make build
EOF