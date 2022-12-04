#!/bin/bash 
declare array 
array=($(grep -i 'rancher' /etc/hosts  |cut -d '.' -f 6 |cut -d ' ' -f 2))

printf ${array[0]}
