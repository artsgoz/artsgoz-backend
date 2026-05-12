#!/usr/bin/env bash
# Enforces the Clean Architecture dependency rule for internal/modules/*.
#
# Rules:
#   1. domain/      must NOT import pgx, fiber, validator, or any sibling
#                   layer (application/, infrastructure/, interface/).
#   2. application/ must NOT import pgx, fiber, or any sibling layer
#                   (infrastructure/, interface/). It MAY import domain/
#                   and internal/platform/{apperr,validator}.
#
# Exit code 0 = clean. Non-zero = violations found, with file:line printed.
#
# Run from repo root:  ./scripts/check-deps.sh
# CI / pre-commit:     bash scripts/check-deps.sh

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
RESET=$'\033[0m'

violations=0

# check_layer <layer-dir-name> <forbidden-pattern> <human-description>
# Greps every *.go file under internal/modules/*/<layer-dir-name>/ for the
# forbidden import pattern. Prints violations as file:line.
check_layer() {
    local layer="$1"
    local pattern="$2"
    local description="$3"

    # find files under all modules' <layer>/ subtrees
    local files
    files=$(find internal/modules -type d -name "$layer" 2>/dev/null \
        | xargs -I{} find {} -name '*.go' 2>/dev/null)

    if [[ -z "$files" ]]; then
        return 0
    fi

    local hits
    hits=$(echo "$files" | xargs grep -nE "$pattern" 2>/dev/null || true)

    if [[ -n "$hits" ]]; then
        echo "${RED}✗ ${layer}/ must not import ${description}${RESET}"
        echo "$hits" | sed 's/^/    /'
        echo
        violations=$((violations + 1))
    else
        echo "${GREEN}✓ ${layer}/ is free of ${description}${RESET}"
    fi
}

echo "${YELLOW}Checking Clean Architecture dependency rules...${RESET}"
echo

# --- domain/ rules: the innermost layer. May only depend on stdlib +
#     internal/platform/apperr.
check_layer "domain" \
    'github\.com/jackc/pgx' \
    "pgx (DB driver belongs in infrastructure/)"

check_layer "domain" \
    'github\.com/gofiber/fiber' \
    "fiber (HTTP framework belongs in interface/)"

check_layer "domain" \
    'github\.com/go-playground/validator' \
    "validator (input concern, lives in application/)"

check_layer "domain" \
    'artsgoz-backend/internal/modules/[^/]+/application' \
    "application/ (inward dependency violation)"

check_layer "domain" \
    'artsgoz-backend/internal/modules/[^/]+/infrastructure' \
    "infrastructure/ (inward dependency violation)"

check_layer "domain" \
    'artsgoz-backend/internal/modules/[^/]+/interface' \
    "interface/ (inward dependency violation)"

# --- application/ rules: orchestrates use cases. May depend on domain/
#     and internal/platform/{apperr,validator}.
check_layer "application" \
    'github\.com/jackc/pgx' \
    "pgx (DB driver belongs in infrastructure/)"

check_layer "application" \
    'github\.com/gofiber/fiber' \
    "fiber (HTTP framework belongs in interface/)"

check_layer "application" \
    'artsgoz-backend/internal/modules/[^/]+/infrastructure' \
    "infrastructure/ (use the domain interface instead)"

check_layer "application" \
    'artsgoz-backend/internal/modules/[^/]+/interface' \
    "interface/ (inward dependency violation)"

echo
if [[ $violations -eq 0 ]]; then
    echo "${GREEN}All dependency rules satisfied.${RESET}"
    exit 0
else
    echo "${RED}${violations} rule(s) violated. See https://github.com/your-org/artsgoz-backend/blob/main/docs/architecture.md#4-dependency-rule-the-one-rule${RESET}"
    exit 1
fi
