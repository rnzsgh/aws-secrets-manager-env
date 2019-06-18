#!/bin/bash

set -e

/aws-secrets-manager-env --secret=/prod/test --secret=prod/test

/test
