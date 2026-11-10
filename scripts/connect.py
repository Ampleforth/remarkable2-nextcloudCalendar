#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-only
"""Authorize a dedicated tablet credential using Nextcloud's browser login flow v2.

Only the browser authorization URL is printed. Credentials go directly to a
private file for transfer to the tablet; they are never printed or put in argv.
"""
import argparse
import json
import os
import pathlib
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request
from zoneinfo import ZoneInfo


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--server", required=True)
    parser.add_argument("--timezone", default="UTC")
    parser.add_argument("--output", type=pathlib.Path, required=True)
    args = parser.parse_args()
    ZoneInfo(args.timezone)
    origin = urllib.parse.urlsplit(args.server)
    if origin.scheme != "https" or not origin.hostname or origin.username or origin.query or origin.fragment:
        parser.error("server must be an HTTPS URL")
    if args.output.exists():
        parser.error("output exists; use a new path to avoid replacing an account")
    opener = urllib.request.build_opener(NoRedirect())

    def safe_url(value):
        url = urllib.parse.urlsplit(value)
        if (url.scheme, url.netloc) != (origin.scheme, origin.netloc) or url.username or url.fragment:
            raise ValueError("Nextcloud returned a URL outside the configured HTTPS origin")
        return value

    def post(url, data=b""):
        request = urllib.request.Request(safe_url(url), data=data, method="POST", headers={
            "User-Agent": "reMarkable 2 Nextcloud Calendar",
            "Content-Type": "application/x-www-form-urlencoded",
            "Accept": "application/json",
        })
        with opener.open(request, timeout=30) as response:
            return json.loads(response.read(1024 * 1024))

    flow = post(args.server.rstrip("/") + "/index.php/login/v2")
    login_url = safe_url(flow["login"])
    poll_url = safe_url(flow["poll"]["endpoint"])
    token = urllib.parse.urlencode({"token": flow["poll"]["token"]}).encode()
    print("Open this URL, select the intended Nextcloud account, and grant access:", flush=True)
    print(login_url, flush=True)
    deadline = time.monotonic() + 19 * 60
    while time.monotonic() < deadline:
        try:
            credentials = post(poll_url, token)
        except urllib.error.HTTPError as error:
            if error.code != 404:
                raise RuntimeError(f"Nextcloud login poll returned HTTP {error.code}") from None
            time.sleep(2)
            continue
        safe_url(credentials["server"])
        if not credentials.get("loginName") or not credentials.get("appPassword"):
            raise ValueError("Nextcloud did not return app credentials")
        args.output.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
        fd, name = tempfile.mkstemp(prefix=".calendar-auth-", dir=args.output.parent)
        try:
            with os.fdopen(fd, "w") as output:
                os.fchmod(output.fileno(), 0o600)
                json.dump({"server": credentials["server"], "username": credentials["loginName"],
                           "app_password": credentials["appPassword"], "timezone": args.timezone}, output)
                output.flush()
                os.fsync(output.fileno())
            # link is exclusive: a concurrent setup must not replace another config.
            os.link(name, args.output)
        finally:
            os.unlink(name)
        print("Authorization saved privately. Transfer the file to the tablet config path.", flush=True)
        return
    raise RuntimeError("Authorization timed out; run setup again")


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, KeyError, RuntimeError) as error:
        # Do not print remote response bodies, which may contain credentials.
        raise SystemExit(f"Calendar setup failed ({type(error).__name__}); check network, server, and output path.")
