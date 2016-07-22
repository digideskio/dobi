#!/bin/bash

set -eu

mkdir -p /code/db

# TODO: why do I need this?
bundle install --deployment

sleep 3
# TODO: wait for db to be ready
bin/rails db:migrate
pg_dump --user postgres --host postgres postgres > /code/db/export.sql
