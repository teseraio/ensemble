#!/bin/bash

tmp_dir=$(mktemp -d -t ci-XXXXXXXXXX)
echo $tmp_dir

# Package and move to website
helm package charts/operator --destination $tmp_dir
mv $tmp_dir/* website/public/charts

# Index the helm repo again
helm repo index website/public/charts
