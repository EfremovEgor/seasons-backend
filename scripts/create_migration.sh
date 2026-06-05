#!/bin/bash

set -e
cd data/sql/migrations
goose -s create $1 sql