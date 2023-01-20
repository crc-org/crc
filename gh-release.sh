#!/bin/bash

set -xeuo pipefail

lastReleasedVersion="$(gh release view -R crc-org/crc --json tagName --jq .tagName)"
currentRepoTag="$(git describe --tags --abbrev=0 --candidates=0 || echo "v0.0.0")"

NONINTERACTIVE=${NONINTERACTIVE:-0}

if ! grep -qP "^v\d\.\d+\.\d+$" <<< "$currentRepoTag"; then
	echo "Tag $currentRepoTag doesn't follow the expected tag format of v0.0.0" >&2
	exit 1
fi

if [[ "$currentRepoTag" = "v0.0.0" ]]; then
	echo "Please tag before publishing release or fetch latest main from upstream" >&2
	exit 1
fi

if [[ "$currentRepoTag" = "$lastReleasedVersion" ]]; then
	echo "$currentRepoTag is already published on github" >&2
	exit 1
fi

function prepare_release_notes() {
	read -r -d '' rn_template << EOF
Downloads are available at: https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/crc/%s
To use these binaries follow the instructions at https://console.redhat.com/openshift/create/local to obtain the needed pull-secret.

-------

Notable Changes
---

- OpenShift %s
- Podman %s
- OKD %s
%s


git shortlog
---

%s
EOF

	changelog="$(git log --format="%h %s" "$lastReleasedVersion..$currentRepoTag")"
	printf -v release_notes "$rn_template" "${currentRepoTag:1}" "$(ocp_version)" "$(podman_version)" "$(okd_version)" "$(notable_changes)" "$changelog"
	echo "$release_notes"
}

function ocp_version() {
	grep -oP "^OPENSHIFT_VERSION\s\?=\s\K\d\.\d+\.\d+$" Makefile
}

function podman_version() {
	grep -oP "^PODMAN_VERSION\s\?=\s\K\d\.\d+\.\d+$" Makefile
}

function okd_version() {
	grep -oP "^OKD_VERSION\s\?=\s\K\d\.\d+\..*$" Makefile
}

function notable_changes() {
	if [[ -e notable-changes.txt ]]; then
		cat notable-changes.txt
	else
		if @confirm "Notable changes file not found, would you still like to release?"; then
			echo "Preparing release without notable changes.." >&2
		else
			exit 2
		fi
	fi
}

@confirm() {
	# return true if NONINTERACTIVE is set
	if [[ "${NONINTERACTIVE}" -eq 1 ]]; then
		return 0
	fi
	local message="$*"
	local result=3

	echo -n "> $message (y/n) " >&2

	while [[ $result -gt 1 ]] ; do
	read -r -s -n 1 choice
	case "$choice" in
	  y|Y ) result=0 ;;
	  n|N ) result=1 ;;
	esac
	done

	return $result
}

release_txt=$(prepare_release_notes)

printf "#####CRC Release notes#####\n\n%s\n\n#####CRC Release notes#####\n" "$release_txt"

if @confirm "Push release to github?"; then
	gh release create "$currentRepoTag" --verify-tag --draft --notes "$release_txt" --title "${currentRepoTag:1}-$(ocp_version)"
else
	echo "Release aborted!!"
	exit 3
fi

