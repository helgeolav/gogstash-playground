#!/bin/bash
githash="$(git rev-parse HEAD | cut -c1-8)"
buildtime="$(date +%Y-%m-%d)"
LDFLAGS="${LDFLAGS} -X bitbucket.org/HelgeOlav/utils/version.BUILDTIME=${buildtime}"
LDFLAGS="${LDFLAGS} -X bitbucket.org/HelgeOlav/utils/version.GITCOMMIT=${githash}"
go build -ldflags "${LDFLAGS}" .
