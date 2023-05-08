#!/bin/bash
nm=`grep  "name:" charts/Chart.yaml |cut -d ' ' -f 2`
ver=`grep  "version:" charts/Chart.yaml |cut -d ' ' -f 2`
k6_operator_release_tag=`echo $nm-$ver`
export k6_operator_release_tag
echo $k6_operator_release_tag
