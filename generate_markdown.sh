#!/bin/bash
godoc-md . > README.md

# Add license info
echo "*Copyright (c) 2013 Freek Kalter.  All rights reserved.
See the LICENSE file.*" >> README.md
