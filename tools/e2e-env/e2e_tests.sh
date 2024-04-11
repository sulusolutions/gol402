#!/bin/bash

go test -tags=e2e ./... -v -count=1 -timeout 30s 