#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-only
"""Exercise a built host backend with the actual AppLoad packet framing."""
import json
import pathlib
import socket
import struct
import subprocess
import sys
import tempfile
import urllib.request

with tempfile.TemporaryDirectory(prefix="calendar-lifecycle-") as tmp:
    path = pathlib.Path(tmp)
    with socket.socket(socket.AF_UNIX, socket.SOCK_SEQPACKET) as listener:
        listener.bind(str(path / "lifecycle.sock"))
        listener.listen(1)
        listener.settimeout(5)
        with socket.socket() as port:
            port.bind(("127.0.0.1", 0))
            address = f"127.0.0.1:{port.getsockname()[1]}"
        process = subprocess.Popen([
            sys.argv[1], "-listen", address, "-cache", str(path / "cache.json"),
            "-config", str(path / "missing.json"), str(path / "lifecycle.sock"),
        ], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            connection, _ = listener.accept()
            with connection:
                with urllib.request.urlopen(f"http://{address}/status", timeout=5) as response:
                    status = json.load(response)
                    assert status["pending"] == [] and not status["busy"], status
                connection.send(struct.pack("=II", 1, 3))
                connection.send(b"abc")
                assert process.poll() is None, "backend exited on ordinary lifecycle packet"
                connection.send(struct.pack("=II", 0xffffffff, 0))
                assert process.wait(timeout=5) == 0
            print("AppLoad lifecycle + HTTP integration passed")
        finally:
            if process.poll() is None:
                process.kill()
            process.communicate(timeout=5)
