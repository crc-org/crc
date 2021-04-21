#!/bin/bash
set -e

if [ $# -lt 1 ]; then
	echo "You need to provide an argument [build_docs, docs_check_links, docs_serve]"
fi

case $1 in
	build_docs)
		echo "Generating html from docs source"
		shift 
		asciidoctor $@
		;;
	docs_check_links)
		echo "Checking if all links are alive in docs source"
		find . -name \*.adoc | xargs -n1 asciidoc-link-check -c /links_ignorelist.json -q
		;;
	docs_serve)
		cd build
		python3 -m http.server 8088
		;;
	*)
		echo "Need to provide one of: build_docs, docs_check_links, docs_serve"
		exit 1
		;;
esac
