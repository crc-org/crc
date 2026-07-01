#!/usr/bin/env python3
"""Extract URLs from a text file and verify they all use HTTPS and return 200."""

import re
import sys
import urllib.request
import urllib.error

URL_PATTERN = re.compile(r'https?://[^\s\)\]\}>",]+')

def extract_urls(filepath):
    with open(filepath) as f:
        return URL_PATTERN.findall(f.read())

def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <file>", file=sys.stderr)
        sys.exit(1)

    urls = extract_urls(sys.argv[1])
    if not urls:
        print("No URLs found in file.")
        sys.exit(1)

    errors = []
    for url in urls:
        if not url.startswith("https://"):
            errors.append(f"NOT HTTPS: {url}")
            continue
        try:
            req = urllib.request.Request(url, method="HEAD", headers={"User-Agent": "url-verify/1.0"})
            with urllib.request.urlopen(req, timeout=10) as resp:
                if resp.status != 200:
                    errors.append(f"HTTP {resp.status}: {url}")
                else:
                    print(f"OK: {url}")
        except urllib.error.HTTPError as e:
            if e.code == 405:
                try:
                    req = urllib.request.Request(url, headers={"User-Agent": "url-verify/1.0"})
                    with urllib.request.urlopen(req, timeout=10) as resp:
                        if resp.status != 200:
                            errors.append(f"HTTP {resp.status}: {url}")
                        else:
                            print(f"OK: {url}")
                except Exception as e2:
                    errors.append(f"FAILED ({e2}): {url}")
            else:
                errors.append(f"HTTP {e.code}: {url}")
        except Exception as e:
            errors.append(f"FAILED ({e}): {url}")
    if errors:
        print("\nErrors:")
        for err in errors:
            print(f"  {err}")
        sys.exit(1)

    print(f"\nAll {len(urls)} URLs are valid HTTPS and returned 200.")

if __name__ == "__main__":
    main()
