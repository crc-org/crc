#!/usr/bin/env python3
"""Extract URLs from stdin and verify they all use HTTPS and return 200."""

import ipaddress
import re
import socket
import sys
import urllib.error
import urllib.parse
import urllib.request

URL_PATTERN = re.compile(r'https?://[^\s\)\]\}>",]+')

def extract_urls(content):
    return URL_PATTERN.findall(content)

def is_safe_host(url):
    """
    Validate that URL points to a safe, non-private destination.
    Blocks private IPs, loopback, link-local, and similar unsafe targets.
    Returns (is_safe, error_message).
    """
    try:
        parsed = urllib.parse.urlparse(url)
        hostname = parsed.hostname
        if not hostname:
            return False, "no hostname"

        # Try to resolve hostname to IP and check if it's private/internal
        try:
            # Get all addresses for this hostname
            addr_info = socket.getaddrinfo(hostname, None)
            for info in addr_info:
                ip_str = info[4][0]
                try:
                    ip = ipaddress.ip_address(ip_str)
                    # Block private, loopback, link-local, multicast, and reserved ranges
                    if (ip.is_private or ip.is_loopback or ip.is_link_local or
                        ip.is_multicast or ip.is_reserved):
                        return False, f"unsafe IP {ip_str} ({hostname})"
                except ValueError:
                    continue
        except socket.gaierror:
            return False, f"cannot resolve {hostname}"

        return True, None
    except Exception as e:
        return False, f"validation error: {e}"

def main():
    content = sys.stdin.read()
    urls = extract_urls(content)
    if not urls:
        print("No URLs found in input.")
        sys.exit(1)

    errors = []
    for url in urls:
        if not url.startswith("https://"):
            errors.append(f"NOT HTTPS: {url}")
            continue

        # Validate host is safe before making any requests
        is_safe, err_msg = is_safe_host(url)
        if not is_safe:
            errors.append(f"UNSAFE HOST ({err_msg}): {url}")
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
