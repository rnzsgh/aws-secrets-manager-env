#!/bin/bash

TAG=$(git log -1 --pretty=%H)

REPOSITORY=rnzdocker1/aws-secrets-manager-env

docker build --tag $REPOSITORY:$TAG .

docker push $REPOSITORY:$TAG


