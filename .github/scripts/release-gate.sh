#!/usr/bin/env bash
set -euo pipefail

: "${GH_TOKEN:?GH_TOKEN must be set}"
: "${GITHUB_OUTPUT:?GITHUB_OUTPUT must be set}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"
: "${HEAD_SHA:?HEAD_SHA must be set}"

mark_not_ready() {
	printf 'ready=false\n' >>"${GITHUB_OUTPUT}"
}

for workflow in verify codeql quality; do
	runs="$(gh api \
		"repos/${GITHUB_REPOSITORY}/actions/workflows/${workflow}.yml/runs" \
		-f "head_sha=${HEAD_SHA}" \
		-f 'event=push' \
		-f 'branch=main' \
		-F 'per_page=100')"
	run="$(jq -c --arg sha "${HEAD_SHA}" '
		[.workflow_runs[] | select(.head_sha == $sha and .event == "push" and .head_branch == "main")]
		| sort_by([.run_number, .run_attempt])
		| last // null
	' <<<"${runs}")"

	if [[ "${run}" == 'null' ]]; then
		echo "${workflow} has not reported for ${HEAD_SHA} yet."
		mark_not_ready
		exit 0
	fi

	status="$(jq -r '.status' <<<"${run}")"
	if [[ "${status}" != 'completed' ]]; then
		echo "${workflow} is still ${status}."
		mark_not_ready
		exit 0
	fi

	conclusion="$(jq -r '.conclusion // ""' <<<"${run}")"
	if [[ "${conclusion}" != 'success' ]]; then
		echo "${workflow} concluded ${conclusion:-without a conclusion}; refusing to release."
		exit 1
	fi
done

main_sha="$(gh api "repos/${GITHUB_REPOSITORY}/commits/main" --jq '.sha')"
if [[ "${main_sha}" != "${HEAD_SHA}" ]]; then
	echo "${HEAD_SHA} is no longer the tip of main; waiting for the newest commit's checks."
	mark_not_ready
	exit 0
fi

printf 'ready=true\n' >>"${GITHUB_OUTPUT}"
