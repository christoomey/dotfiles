#!/usr/bin/env python3
"""The regression the jobs rework exists for: a parent that detaches a worker
and exits immediately, with SIGHUP sent to its process group right afterwards
(what herdr does when the popup closes). A `nohup ... &!` fork loses about a
third of the workers here; herdr-detach must lose none.

  race_test.py [ITERATIONS] [OUT_DIR]
"""
import os
import signal
import subprocess
import sys
import tempfile
import time

here = os.path.dirname(os.path.abspath(__file__))
iterations = int(sys.argv[1]) if len(sys.argv) > 1 else 200
out = sys.argv[2] if len(sys.argv) > 2 else tempfile.mkdtemp(prefix="herdr-race-")

# The workers only append; the first one must not race the directory into being.
os.makedirs(out, exist_ok=True)

for i in range(iterations):
    parent = subprocess.Popen(
        [os.path.join(here, "race_parent.zsh"), out, str(i)],
        start_new_session=True,
        stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
    )
    parent.wait()
    try:
        os.killpg(parent.pid, signal.SIGHUP)
    except ProcessLookupError:
        pass

time.sleep(2)


def count(name):
    try:
        with open(os.path.join(out, name)) as fh:
            return len([line for line in fh if line.strip()])
    except FileNotFoundError:
        return 0


started, done = count("started.log"), count("done.log")
print("iterations=%d started=%d done=%d (%s)" % (iterations, started, done, out))
if started != iterations or done != iterations:
    print("FAIL: workers were lost to SIGHUP")
    sys.exit(1)
print("ok: every detached worker survived")
