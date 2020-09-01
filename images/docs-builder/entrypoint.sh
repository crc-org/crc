#!/bin/bash
set -e

if [ $# -lt 1 ]; then
	echo "You need to provide an argument [build_docs, build_docs_pdf, docs_serve]"
fi

case $1 in
	build_docs)
		echo "Generating html from docs source"
		shift 
		asciidoctor $@
		;;
	build_docs_pdf)
		echo "Generating pdf from docs source"
		shift 
		asciidoctor-pdf $@
		;;
	docs_serve)
		cd build
		python3 -m http.server 8088
		;;
	*)
		echo "Need to provide one of: build_docs, build_docs_pdf,  docs_serve"
		exit 1
		;;
esac
