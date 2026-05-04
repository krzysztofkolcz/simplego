#!/usr/bin/env bash

set -e

# Funkcja zwiększająca patch (np. 0.0.16 -> 0.0.17)
bump_patch() {
  local version=$1
  IFS='.' read -r major minor patch <<< "$version"
  patch=$((patch + 1))
  echo "${major}.${minor}.${patch}"
}

# --- Makefile ---
CURRENT_VERSION=$(grep -E '^VERSION=' Makefile | cut -d '=' -f2)
NEW_VERSION=$(bump_patch "$CURRENT_VERSION")

echo "Makefile: $CURRENT_VERSION -> $NEW_VERSION"

sed -i "s/^VERSION=${CURRENT_VERSION}/VERSION=${NEW_VERSION}/" Makefile

# --- values.yaml ---
CURRENT_TAG=$(grep -E 'tag:' charts/values.yaml | sed -E 's/.*"(.+)".*/\1/')
echo "values.yaml: $CURRENT_TAG -> $NEW_VERSION"

sed -i "s/tag: \"${CURRENT_TAG}\"/tag: \"${NEW_VERSION}\"/" charts/values.yaml

# --- Chart.yaml ---
CURRENT_CHART_VERSION=$(grep '^version:' charts/Chart.yaml | awk '{print $2}')
NEW_CHART_VERSION=$(bump_patch "$CURRENT_CHART_VERSION")

echo "Chart.yaml version: $CURRENT_CHART_VERSION -> $NEW_CHART_VERSION"

sed -i "s/^version: ${CURRENT_CHART_VERSION}/version: ${NEW_CHART_VERSION}/" charts/Chart.yaml

CURRENT_APP_VERSION=$(grep '^appVersion:' charts/Chart.yaml | sed -E 's/.*"(.+)".*/\1/')
echo "Chart.yaml appVersion: $CURRENT_APP_VERSION -> $NEW_VERSION"

sed -i "s/appVersion: \"${CURRENT_APP_VERSION}\"/appVersion: \"${NEW_VERSION}\"/" charts/Chart.yaml