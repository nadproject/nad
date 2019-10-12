#!/usr/bin/env bash
# dump_schema.sh dumps the current system's nad schema
set -eux

sqlite3 ~/.nad/nad.db .schema
