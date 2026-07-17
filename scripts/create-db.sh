#!/usr/bin/env bash
# Create PostgreSQL role and database for Hublio if they do not exist.
#
# Application credentials (.env):
#   DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
#
# Bootstrap admin (must be able to CREATE ROLE / CREATE DATABASE — usually a superuser):
#   DB_ADMIN_USER     (host default: postgres; docker default: DB_USER)
#   DB_ADMIN_PASSWORD
#   DB_ADMIN_DB       (default: postgres)
#
# Docker:
#   DB_VIA_DOCKER=1
#   DB_DOCKER_CONTAINER=postgres
#
# Tip: if Postgres was first initialized by docker-compose with POSTGRES_USER=root,
# set DB_ADMIN_USER=root and DB_ADMIN_PASSWORD to that password when creating a new app role.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -f "${ROOT_DIR}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source <(sed 's/\r$//' "${ROOT_DIR}/.env")
  set +a
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-hublio}"
DB_PASSWORD="${DB_PASSWORD:-hublio}"
DB_NAME="${DB_NAME:-hublio}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

DB_DOCKER_CONTAINER="${DB_DOCKER_CONTAINER:-postgres}"

use_docker=0
if [[ "${DB_VIA_DOCKER:-}" == "1" ]] || { [[ -z "${DB_VIA_DOCKER:-}" ]] && command -v docker >/dev/null 2>&1 && docker ps --format '{{.Names}}' 2>/dev/null | grep -qx "${DB_DOCKER_CONTAINER}"; }; then
  use_docker=1
fi

if [[ "${use_docker}" -eq 0 && ( "${DB_HOST}" == "db" || "${DB_HOST}" == "postgres" || "${DB_HOST}" == "redis" ) ]]; then
  echo "note: DB_HOST=${DB_HOST} looks like a Compose service name; using localhost for host-side psql"
  DB_HOST="localhost"
fi

if [[ "${use_docker}" -eq 1 ]]; then
  DB_ADMIN_USER="${DB_ADMIN_USER:-${DB_USER}}"
  DB_ADMIN_PASSWORD="${DB_ADMIN_PASSWORD:-${DB_PASSWORD}}"
  DB_ADMIN_DB="${DB_ADMIN_DB:-postgres}"
else
  DB_ADMIN_USER="${DB_ADMIN_USER:-postgres}"
  DB_ADMIN_PASSWORD="${DB_ADMIN_PASSWORD:-}"
  DB_ADMIN_DB="${DB_ADMIN_DB:-postgres}"
fi

# On Ubuntu/WSL, local Postgres often allows peer auth as OS user `postgres`.
use_peer_postgres=0
if [[ "${use_docker}" -eq 0 ]] && command -v sudo >/dev/null 2>&1; then
  if sudo -n -u postgres true 2>/dev/null || { [[ "$(id -u)" -eq 0 ]] && sudo -u postgres true 2>/dev/null; }; then
    # Prefer peer postgres when configured admin is missing/weak, or when explicitly asked.
    if [[ "${DB_ADMIN_USE_PEER:-}" == "1" ]] || [[ "${DB_ADMIN_USER}" != "postgres" && -z "${DB_ADMIN_PASSWORD}" ]]; then
      use_peer_postgres=1
    fi
  fi
fi

escape_literal() {
  printf "%s" "$1" | sed "s/'/''/g"
}

USER_LIT="$(escape_literal "${DB_USER}")"
PASS_LIT="$(escape_literal "${DB_PASSWORD}")"
NAME_LIT="$(escape_literal "${DB_NAME}")"

run_psql() {
  local database="$1"
  shift

  if [[ "${use_docker}" -eq 1 ]]; then
    if [[ -n "${DB_ADMIN_PASSWORD}" ]]; then
      docker exec -e PGPASSWORD="${DB_ADMIN_PASSWORD}" -i "${DB_DOCKER_CONTAINER}" \
        psql -v ON_ERROR_STOP=1 -U "${DB_ADMIN_USER}" -d "${database}" "$@"
    else
      docker exec -i "${DB_DOCKER_CONTAINER}" \
        psql -v ON_ERROR_STOP=1 -U "${DB_ADMIN_USER}" -d "${database}" "$@"
    fi
    return
  fi

  if [[ "${use_peer_postgres}" -eq 1 ]]; then
    sudo -u postgres psql -v ON_ERROR_STOP=1 -d "${database}" "$@"
    return
  fi

  if ! command -v psql >/dev/null 2>&1; then
    echo "error: psql not found and docker container '${DB_DOCKER_CONTAINER}' is not running" >&2
    echo "hint: start infra with 'make noapp' or install PostgreSQL client tools" >&2
    exit 1
  fi

  if [[ -n "${DB_ADMIN_PASSWORD}" ]]; then
    PGPASSWORD="${DB_ADMIN_PASSWORD}" psql \
      -v ON_ERROR_STOP=1 \
      -h "${DB_HOST}" \
      -p "${DB_PORT}" \
      -U "${DB_ADMIN_USER}" \
      -d "${database}" \
      "$@"
  else
    psql \
      -v ON_ERROR_STOP=1 \
      -h "${DB_HOST}" \
      -p "${DB_PORT}" \
      -U "${DB_ADMIN_USER}" \
      -d "${database}" \
      "$@"
  fi
}

# If configured admin lacks CREATEROLE, fall back to local peer `postgres` (WSL/Ubuntu).
probe_admin_caps() {
  run_psql "${DB_ADMIN_DB}" -tAc \
    "SELECT CASE WHEN rolsuper THEN 'super' WHEN rolcreaterole THEN 'createrole' ELSE 'none' END
     FROM pg_roles WHERE rolname = current_user" | tr -d '[:space:]'
}

echo "==> Admin connection: user=${DB_ADMIN_USER} db=${DB_ADMIN_DB} host=${DB_HOST} docker=${use_docker} peer=${use_peer_postgres}"

ADMIN_CAPS="$(probe_admin_caps || true)"

if [[ "${ADMIN_CAPS}" == "none" || -z "${ADMIN_CAPS}" ]] && [[ "${use_docker}" -eq 0 ]] && [[ "${use_peer_postgres}" -eq 0 ]]; then
  if command -v sudo >/dev/null 2>&1 && { sudo -n -u postgres true 2>/dev/null || [[ "$(id -u)" -eq 0 ]]; }; then
    echo "note: admin '${DB_ADMIN_USER}' lacks CREATEROLE; switching to peer auth via 'sudo -u postgres'"
    use_peer_postgres=1
    DB_ADMIN_USER=postgres
    DB_ADMIN_PASSWORD=
    DB_ADMIN_DB=postgres
    ADMIN_CAPS="$(probe_admin_caps)"
  fi
fi

if [[ -z "${ADMIN_CAPS}" ]]; then
  echo "error: could not determine privileges for admin user '${DB_ADMIN_USER}'" >&2
  exit 1
fi

echo "==> Admin privileges: ${ADMIN_CAPS}"

ROLE_EXISTS="$(
  run_psql "${DB_ADMIN_DB}" -tAc \
    "SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = '${USER_LIT}'" | tr -d '[:space:]'
)"

if [[ "${DB_USER}" == "${DB_ADMIN_USER}" ]]; then
  echo "==> Role '${DB_USER}' is the admin user; skipping CREATE ROLE"
  if [[ "${ADMIN_CAPS}" == "none" ]]; then
    echo "error: admin user '${DB_ADMIN_USER}' cannot create roles/databases." >&2
    echo "hint: set DB_ADMIN_USER to a PostgreSQL superuser (often 'postgres' or the original POSTGRES_USER)." >&2
    exit 1
  fi
elif [[ "${ROLE_EXISTS}" == "1" ]]; then
  echo "==> Role '${DB_USER}' already exists"
  if [[ "${ADMIN_CAPS}" != "none" ]]; then
    run_psql "${DB_ADMIN_DB}" -c "ALTER ROLE \"${USER_LIT}\" WITH LOGIN PASSWORD '${PASS_LIT}';" >/dev/null
    echo "password refreshed for role ${DB_USER}"
  else
    echo "note: admin lacks CREATEROLE; leaving existing role password unchanged"
  fi
else
  echo "==> Creating role '${DB_USER}'"
  if [[ "${ADMIN_CAPS}" == "none" ]]; then
    echo "error: permission denied to create role '${DB_USER}'." >&2
    echo "admin user '${DB_ADMIN_USER}' is not a superuser and has no CREATEROLE." >&2
    echo "" >&2
    echo "fix .env bootstrap credentials, for example:" >&2
    echo "  DB_ADMIN_USER=postgres" >&2
    echo "  DB_ADMIN_PASSWORD=<postgres-password>" >&2
    echo "  DB_ADMIN_DB=postgres" >&2
    echo "" >&2
    echo "If this Postgres volume was created by docker-compose with POSTGRES_USER=root:" >&2
    echo "  DB_ADMIN_USER=root" >&2
    echo "  DB_ADMIN_PASSWORD=<that-password>" >&2
    exit 1
  fi

  run_psql "${DB_ADMIN_DB}" <<SQL
CREATE ROLE "${USER_LIT}" LOGIN PASSWORD '${PASS_LIT}';
SQL
  echo "created role ${DB_USER}"
fi

echo "==> Ensuring database '${DB_NAME}' exists"
DB_EXISTS="$(
  run_psql "${DB_ADMIN_DB}" -tAc "SELECT 1 FROM pg_database WHERE datname = '${NAME_LIT}'" | tr -d '[:space:]'
)"

if [[ "${DB_EXISTS}" != "1" ]]; then
  if [[ "${ADMIN_CAPS}" == "none" ]]; then
    echo "error: database '${DB_NAME}' does not exist and admin cannot CREATE DATABASE" >&2
    exit 1
  fi
  run_psql "${DB_ADMIN_DB}" -c "CREATE DATABASE \"${NAME_LIT}\" OWNER \"${USER_LIT}\";"
  echo "created database ${DB_NAME}"
else
  echo "database ${DB_NAME} already exists"
  if [[ "${ADMIN_CAPS}" != "none" && "${DB_USER}" != "${DB_ADMIN_USER}" ]]; then
    run_psql "${DB_ADMIN_DB}" -c "ALTER DATABASE \"${NAME_LIT}\" OWNER TO \"${USER_LIT}\";" >/dev/null || true
  fi
fi

echo "==> Granting privileges on '${DB_NAME}' to '${DB_USER}'"
if [[ "${ADMIN_CAPS}" != "none" ]]; then
  run_psql "${DB_NAME}" <<SQL
GRANT ALL PRIVILEGES ON DATABASE "${NAME_LIT}" TO "${USER_LIT}";
GRANT ALL ON SCHEMA public TO "${USER_LIT}";
ALTER SCHEMA public OWNER TO "${USER_LIT}";
SQL
else
  echo "note: skipped GRANT (admin lacks privileges); ensure '${DB_USER}' can already use '${DB_NAME}'"
fi

echo "==> Verifying app login"
if [[ "${use_docker}" -eq 1 ]]; then
  docker exec -e PGPASSWORD="${DB_PASSWORD}" -i "${DB_DOCKER_CONTAINER}" \
    psql -v ON_ERROR_STOP=1 -U "${DB_USER}" -d "${DB_NAME}" -c "SELECT current_user, current_database();" >/dev/null
else
  PGPASSWORD="${DB_PASSWORD}" psql \
    -v ON_ERROR_STOP=1 \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -c "SELECT current_user, current_database();" >/dev/null
fi

echo "==> Done"
echo "connection: host=${DB_HOST} port=${DB_PORT} user=${DB_USER} dbname=${DB_NAME} sslmode=${DB_SSLMODE}"
