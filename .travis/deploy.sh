#!/bin/bash

eval "$(ssh-agent -s)"
chmod 600 .travis/id_rsa
echo -e "Host $DEPLOY_URL\n\tStrictHostKeyChecking no\n" >> ~/.ssh/config
ssh-add .travis/id_rsa

git config --global push.default matching
git remote add deploy ssh://$DEPLOY_USER@$DEPLOY_URL:$DEPLOY_DIR
git push deploy master

ssh $DEPLOY_USER@$DEPLOY_URL <<EOF
    cd $DEPLOY_DIR
    make build
EOF