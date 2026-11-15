#!/usr/bin/env sh
set -eu

# Migrations remain disabled until the migration tool and first approved
# schema are added. Enable only after setting RUN_MIGRATIONS=true in Liara.
if [ "${RUN_MIGRATIONS:-false}" = "true" ]; then
  go run ./cmd/migrate up
fi
